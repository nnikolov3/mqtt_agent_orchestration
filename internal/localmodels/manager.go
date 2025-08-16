package localmodels

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LRUEntry represents an entry in the LRU cache
type LRUEntry struct {
	modelName string
	lastUsed  time.Time
}

// Manager manages local models with GPU memory constraints and LRU eviction
type Manager struct {
	mu              sync.RWMutex
	models          map[string]Model
	modelConfigs    map[string]ModelConfig
	gpuMemory       GPUMemoryInfo
	maxGPUMemory    uint64
	nvidiaSMIPath   string
	monitorInterval time.Duration
	stopMonitoring  chan struct{}

	// LRU cache management
	lruList         *list.List
	lruMap          map[string]*list.Element
	maxLoadedModels int
}

// NewManager creates a new local model manager
func NewManager(config ModelManagerConfig) (*Manager, error) {
	m := &Manager{
		models:          make(map[string]Model),
		modelConfigs:    config.Models,
		maxGPUMemory:    config.MaxGPUMemory,
		nvidiaSMIPath:   config.NvidiaSMIPath,
		monitorInterval: config.MonitorInterval,
		stopMonitoring:  make(chan struct{}),

		// LRU cache initialization
		lruList:         list.New(),
		lruMap:          make(map[string]*list.Element),
		maxLoadedModels: 3, // Limit simultaneous loaded models based on GPU memory
	}

	// Initialize GPU memory monitoring
	if err := m.updateGPUMemoryInfo(); err != nil {
		log.Printf("Warning: Failed to initialize GPU memory info: %v", err)
		// Continue without GPU monitoring for CPU-only setups
	}

	// Start background GPU monitoring
	go m.monitorGPUMemory()

	log.Printf("Local model manager initialized with %d model configs", len(config.Models))
	return m, nil
}

// LoadModel loads a specific model if memory allows
func (m *Manager) LoadModel(ctx context.Context, modelName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already loaded
	if model, exists := m.models[modelName]; exists && model.IsLoaded() {
		log.Printf("Model %s already loaded", modelName)
		return nil
	}

	// Get model config
	config, exists := m.modelConfigs[modelName]
	if !exists {
		return fmt.Errorf("model config not found for %s", modelName)
	}

	// Check GPU memory availability and LRU constraints
	if !m.canLoadModel(config) {
		// Try to free memory by evicting LRU models
		if err := m.evictLRUModels(ctx, config.MemoryLimit); err != nil {
			return fmt.Errorf("insufficient GPU memory to load model %s (requires ~%dMB, available: %dMB): %w",
				modelName, config.MemoryLimit, m.gpuMemory.Free, err)
		}
	}

	// Create model instance based on type
	var model Model
	var err error

	switch config.Type {
	case ModelTypeText:
		model, err = NewQwenTextModel(config)
	case ModelTypeMultimodal:
		if strings.Contains(modelName, "minicpm") {
			model, err = NewMiniCPMModel(config)
		} else if strings.Contains(modelName, "qwen") {
			model, err = NewQwenMultimodalModel(config)
		} else {
			model, err = NewMiniCPMModel(config)
		}
	default:
		return fmt.Errorf("unsupported model type: %s", config.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to create model %s: %w", modelName, err)
	}

	// Load the model
	if err := model.Load(ctx); err != nil {
		return fmt.Errorf("failed to load model %s: %w", modelName, err)
	}

	m.models[modelName] = model

	// Add to LRU cache
	m.addToLRU(modelName)

	log.Printf("✅ Model %s loaded successfully", modelName)
	return nil
}

// UnloadModel unloads a specific model
func (m *Manager) UnloadModel(ctx context.Context, modelName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	model, exists := m.models[modelName]
	if !exists {
		return fmt.Errorf("model %s not loaded", modelName)
	}

	if err := model.Unload(ctx); err != nil {
		return fmt.Errorf("failed to unload model %s: %w", modelName, err)
	}

	delete(m.models, modelName)

	// Remove from LRU cache
	m.removeFromLRU(modelName)

	log.Printf("✅ Model %s unloaded successfully", modelName)
	return nil
}

// GetModel returns a loaded model and updates LRU
func (m *Manager) GetModel(modelName string) (Model, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	model, exists := m.models[modelName]
	if !exists || !model.IsLoaded() {
		return nil, fmt.Errorf("model %s not loaded", modelName)
	}

	// Update LRU on access
	m.updateLRU(modelName)

	return model, nil
}

// GetAvailableModels returns list of available model configurations
func (m *Manager) GetAvailableModels() []string {
	var models []string
	for name := range m.modelConfigs {
		models = append(models, name)
	}
	return models
}

// GetLoadedModels returns list of currently loaded models
func (m *Manager) GetLoadedModels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var models []string
	for name, model := range m.models {
		if model.IsLoaded() {
			models = append(models, name)
		}
	}
	return models
}

// GetModelStatus returns status of all models
func (m *Manager) GetModelStatus() map[string]ModelStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := make(map[string]ModelStatus)

	for name := range m.modelConfigs {
		modelStatus := ModelStatus{
			Name:        name,
			State:       StateUnloaded,
			MemoryUsage: 0,
		}

		if model, exists := m.models[name]; exists {
			if model.IsLoaded() {
				modelStatus.State = StateLoaded
				modelStatus.MemoryUsage = model.GetMemoryUsage()
			}
		}

		status[name] = modelStatus
	}

	return status
}

// canLoadModel checks if there's enough GPU memory to load a model
func (m *Manager) canLoadModel(config ModelConfig) bool {
	// Check if we have enough free GPU memory
	requiredMemory := config.MemoryLimit
	return m.gpuMemory.Free >= requiredMemory
}

// monitorGPUMemory monitors GPU memory usage in background
func (m *Manager) monitorGPUMemory() {
	ticker := time.NewTicker(m.monitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := m.updateGPUMemoryInfo(); err != nil {
				log.Printf("Warning: Failed to update GPU memory info: %v", err)
				continue
			}

			// Check if memory usage is critical
			if m.gpuMemory.Used > m.maxGPUMemory {
				log.Printf("🚨 GPU memory usage critical: %dMB used, %dMB max",
					m.gpuMemory.Used, m.maxGPUMemory)
				m.handleMemoryPressure()
			}

		case <-m.stopMonitoring:
			return
		}
	}
}

// updateGPUMemoryInfo updates GPU memory information using nvidia-smi
func (m *Manager) updateGPUMemoryInfo() error {
	cmd := exec.Command(m.nvidiaSMIPath,
		"--query-gpu=memory.total,memory.used,memory.free",
		"--format=csv,noheader,nounits")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("nvidia-smi execution failed: %w, output: %s", err, string(output))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 1 {
		return fmt.Errorf("unexpected nvidia-smi output: %s", string(output))
	}

	// Parse first GPU (index 0)
	values := strings.Split(strings.TrimSpace(lines[0]), ", ")
	if len(values) != 3 {
		return fmt.Errorf("unexpected nvidia-smi format: %s", string(output))
	}

	total, err := parseMemoryValue(values[0])
	if err != nil {
		return fmt.Errorf("failed to parse total memory: %w", err)
	}

	used, err := parseMemoryValue(values[1])
	if err != nil {
		return fmt.Errorf("failed to parse used memory: %w", err)
	}

	free, err := parseMemoryValue(values[2])
	if err != nil {
		return fmt.Errorf("failed to parse free memory: %w", err)
	}

	m.mu.Lock()
	m.gpuMemory = GPUMemoryInfo{
		Total:     total,
		Used:      used,
		Free:      free,
		Timestamp: time.Now(),
	}
	m.mu.Unlock()

	return nil
}

// parseMemoryValue parses memory value from nvidia-smi output
func parseMemoryValue(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// handleMemoryPressure attempts to free GPU memory by unloading idle models
func (m *Manager) handleMemoryPressure() {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Printf("Handling GPU memory pressure...")

	// Find least recently used model to unload using proper LRU
	oldestModel := m.getLRUModel()
	if oldestModel == "" {
		log.Printf("No models available to unload")
		return
	}

	log.Printf("Unloading LRU model %s to free GPU memory", oldestModel)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := m.models[oldestModel].Unload(ctx); err != nil {
		log.Printf("Failed to unload model %s: %v", oldestModel, err)
		return
	}

	delete(m.models, oldestModel)
	m.removeFromLRU(oldestModel)
	log.Printf("✅ Model %s unloaded to free memory", oldestModel)
}

// addToLRU adds a model to the LRU cache
func (m *Manager) addToLRU(modelName string) {
	now := time.Now()

	// Remove if already exists
	if elem, exists := m.lruMap[modelName]; exists {
		m.lruList.Remove(elem)
	}

	// Add to front
	entry := &LRUEntry{
		modelName: modelName,
		lastUsed:  now,
	}
	elem := m.lruList.PushFront(entry)
	m.lruMap[modelName] = elem

	// Enforce max loaded models limit
	for m.lruList.Len() > m.maxLoadedModels {
		m.evictOldest()
	}
}

// updateLRU updates the last used time for a model
func (m *Manager) updateLRU(modelName string) {
	if elem, exists := m.lruMap[modelName]; exists {
		entry := elem.Value.(*LRUEntry)
		entry.lastUsed = time.Now()
		m.lruList.MoveToFront(elem)
	}
}

// removeFromLRU removes a model from the LRU cache
func (m *Manager) removeFromLRU(modelName string) {
	if elem, exists := m.lruMap[modelName]; exists {
		m.lruList.Remove(elem)
		delete(m.lruMap, modelName)
	}
}

// getLRUModel returns the least recently used model name
func (m *Manager) getLRUModel() string {
	if m.lruList.Len() == 0 {
		return ""
	}

	back := m.lruList.Back()
	if back == nil {
		return ""
	}

	entry := back.Value.(*LRUEntry)
	return entry.modelName
}

// evictOldest evicts the oldest model to enforce cache limits
func (m *Manager) evictOldest() {
	oldestModel := m.getLRUModel()
	if oldestModel == "" {
		return
	}

	// Unload the model
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if model, exists := m.models[oldestModel]; exists {
		if err := model.Unload(ctx); err != nil {
			log.Printf("Failed to unload LRU model %s: %v", oldestModel, err)
			return
		}
		delete(m.models, oldestModel)
	}

	m.removeFromLRU(oldestModel)
	log.Printf("✅ Evicted LRU model %s to enforce cache limits", oldestModel)
}

// evictLRUModels evicts LRU models to free the required memory
func (m *Manager) evictLRUModels(ctx context.Context, requiredMemory uint64) error {
	freedMemory := uint64(0)

	for freedMemory < requiredMemory && m.lruList.Len() > 0 {
		oldestModel := m.getLRUModel()
		if oldestModel == "" {
			break
		}

		// Get model memory usage before unloading
		if model, exists := m.models[oldestModel]; exists {
			modelMemory := model.GetMemoryUsage()

			if err := model.Unload(ctx); err != nil {
				log.Printf("Failed to unload LRU model %s: %v", oldestModel, err)
				continue
			}

			delete(m.models, oldestModel)
			m.removeFromLRU(oldestModel)
			freedMemory += modelMemory

			log.Printf("✅ Evicted LRU model %s, freed %dMB", oldestModel, modelMemory)
		}
	}

	// Update GPU memory info
	if err := m.updateGPUMemoryInfo(); err != nil {
		log.Printf("Warning: Failed to update GPU memory after eviction: %v", err)
	}

	if freedMemory < requiredMemory {
		return fmt.Errorf("could not free enough memory: freed %dMB, required %dMB", freedMemory, requiredMemory)
	}

	return nil
}

// GetGPUMemoryInfo returns current GPU memory information
func (m *Manager) GetGPUMemoryInfo() GPUMemoryInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.gpuMemory
}

// Shutdown gracefully shuts down the model manager
func (m *Manager) Shutdown(ctx context.Context) error {
	log.Printf("Shutting down local model manager...")

	// Stop monitoring
	close(m.stopMonitoring)

	// Unload all models
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, model := range m.models {
		if model.IsLoaded() {
			log.Printf("Unloading model %s...", name)
			if err := model.Unload(ctx); err != nil {
				log.Printf("Failed to unload model %s: %v", name, err)
			}
		}
	}

	log.Printf("✅ Local model manager shutdown complete")
	return nil
}
