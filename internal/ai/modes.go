package ai

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// OperationalMode defines how the AI system operates
type OperationalMode int

const (
	LOCAL_ONLY OperationalMode = iota
	REMOTE_ONLY
	HYBRID
)

func (m OperationalMode) String() string {
	switch m {
	case LOCAL_ONLY:
		return "LOCAL_ONLY"
	case REMOTE_ONLY:
		return "REMOTE_ONLY"
	case HYBRID:
		return "HYBRID"
	default:
		return "UNKNOWN"
	}
}

// ModeManager handles operational mode switching and provider selection
type ModeManager struct {
	mu              sync.RWMutex
	currentMode     OperationalMode
	config          *AIHelperConfig
	providers       map[string]Provider
	localModels     []string
	remoteModels    []string
	circuitBreakers *CircuitBreakerManager
}

// NewModeManager creates a new operational mode manager
func NewModeManager(config *AIHelperConfig) *ModeManager {
	return &ModeManager{
		currentMode:     HYBRID,
		config:          config,
		providers:       make(map[string]Provider),
		localModels:     []string{"qwen-omni-3b", "minicpm-v-4"},
		remoteModels:    []string{"cerebras", "nvidia", "gemini", "grok", "groq"},
		circuitBreakers: NewCircuitBreakerManager(DefaultCircuitBreakerConfig()),
	}
}

// SetMode changes the operational mode
func (m *ModeManager) SetMode(mode OperationalMode) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch mode {
	case LOCAL_ONLY:
		if !m.hasLocalProviders() {
			return fmt.Errorf("no local providers available for LOCAL_ONLY mode")
		}
	case REMOTE_ONLY:
		if !m.hasRemoteProviders() {
			return fmt.Errorf("no remote providers available for REMOTE_ONLY mode")
		}
	case HYBRID:
		if !m.hasLocalProviders() && !m.hasRemoteProviders() {
			return fmt.Errorf("no providers available for HYBRID mode")
		}
	default:
		return fmt.Errorf("invalid operational mode: %v", mode)
	}

	m.currentMode = mode
	return nil
}

// GetMode returns the current operational mode
func (m *ModeManager) GetMode() OperationalMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentMode
}

// SelectProvider chooses the best provider based on current mode and task complexity
func (m *ModeManager) SelectProvider(ctx context.Context, taskComplexity string) (string, APIConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	available := m.config.GetAvailableAPIs()
	if len(available) == 0 {
		return "", APIConfig{}, fmt.Errorf("no providers available")
	}

	// Filter out providers with open circuits
	healthy := m.filterHealthyProviders(available)
	if len(healthy) == 0 {
		return "", APIConfig{}, fmt.Errorf("no healthy providers available")
	}

	var candidates map[string]APIConfig

	switch m.currentMode {
	case LOCAL_ONLY:
		candidates = m.filterLocalProviders(healthy)
	case REMOTE_ONLY:
		candidates = m.filterRemoteProviders(healthy)
	case HYBRID:
		candidates = m.selectHybridProviders(healthy, taskComplexity)
	default:
		return "", APIConfig{}, fmt.Errorf("invalid operational mode: %v", m.currentMode)
	}

	if len(candidates) == 0 {
		return "", APIConfig{}, fmt.Errorf("no suitable healthy providers for mode %s", m.currentMode)
	}

	return m.selectBestProvider(candidates, taskComplexity)
}

// filterLocalProviders returns only local model providers
func (m *ModeManager) filterLocalProviders(available map[string]APIConfig) map[string]APIConfig {
	local := make(map[string]APIConfig)
	for name, config := range available {
		if m.isLocalProvider(name) {
			local[name] = config
		}
	}
	return local
}

// filterRemoteProviders returns only remote API providers
func (m *ModeManager) filterRemoteProviders(available map[string]APIConfig) map[string]APIConfig {
	remote := make(map[string]APIConfig)
	for name, config := range available {
		if !m.isLocalProvider(name) {
			remote[name] = config
		}
	}
	return remote
}

// selectHybridProviders uses intelligent routing for hybrid mode
func (m *ModeManager) selectHybridProviders(available map[string]APIConfig, taskComplexity string) map[string]APIConfig {
	switch taskComplexity {
	case "low":
		local := m.filterLocalProviders(available)
		if len(local) > 0 {
			return local
		}
		return m.filterRemoteProviders(available)
	case "high":
		remote := m.filterRemoteProviders(available)
		if len(remote) > 0 {
			return remote
		}
		return m.filterLocalProviders(available)
	default:
		return available
	}
}

// selectBestProvider chooses the optimal provider from candidates
func (m *ModeManager) selectBestProvider(candidates map[string]APIConfig, taskComplexity string) (string, APIConfig, error) {
	var priorities []string

	switch taskComplexity {
	case "high":
		priorities = []string{"cerebras", "nvidia", "gemini", "grok", "groq"}
	case "medium":
		priorities = []string{"nvidia", "cerebras", "groq", "gemini", "grok"}
	case "low":
		priorities = []string{"groq", "cerebras", "nvidia", "gemini", "grok"}
	default:
		priorities = []string{"cerebras", "nvidia", "gemini", "grok", "groq"}
	}

	for _, provider := range priorities {
		if config, exists := candidates[provider]; exists {
			return provider, config, nil
		}
	}

	for name, config := range candidates {
		return name, config, nil
	}

	return "", APIConfig{}, fmt.Errorf("no suitable provider found")
}

// isLocalProvider determines if a provider is local
func (m *ModeManager) isLocalProvider(name string) bool {
	localProviders := []string{"local", "llama", "qwen", "minicpm"}
	for _, local := range localProviders {
		if strings.Contains(strings.ToLower(name), local) {
			return true
		}
	}
	return false
}

// hasLocalProviders checks if local providers are available
func (m *ModeManager) hasLocalProviders() bool {
	available := m.config.GetAvailableAPIs()
	for name := range available {
		if m.isLocalProvider(name) {
			return true
		}
	}
	return false
}

// hasRemoteProviders checks if remote providers are available
func (m *ModeManager) hasRemoteProviders() bool {
	available := m.config.GetAvailableAPIs()
	for name := range available {
		if !m.isLocalProvider(name) {
			return true
		}
	}
	return false
}

// GetModeStats returns statistics about current mode usage
func (m *ModeManager) GetModeStats() ModeStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	available := m.config.GetAvailableAPIs()
	local := m.filterLocalProviders(available)
	remote := m.filterRemoteProviders(available)

	return ModeStats{
		CurrentMode:       m.currentMode,
		LocalProviders:    len(local),
		RemoteProviders:   len(remote),
		TotalProviders:    len(available),
		AvailableLocal:    getProviderNames(local),
		AvailableRemote:   getProviderNames(remote),
		LastModeChange:    time.Now(),
		ModeChangeCount:   0,
		LocalRequestCount: 0,
		RemoteRequestCount: 0,
	}
}

// ModeStats contains statistics about operational modes
type ModeStats struct {
	CurrentMode        OperationalMode `json:"current_mode"`
	LocalProviders     int             `json:"local_providers"`
	RemoteProviders    int             `json:"remote_providers"`
	TotalProviders     int             `json:"total_providers"`
	AvailableLocal     []string        `json:"available_local"`
	AvailableRemote    []string        `json:"available_remote"`
	LastModeChange     time.Time       `json:"last_mode_change"`
	ModeChangeCount    int64           `json:"mode_change_count"`
	LocalRequestCount  int64           `json:"local_request_count"`
	RemoteRequestCount int64           `json:"remote_request_count"`
}

// getProviderNames extracts provider names from API configs
func getProviderNames(configs map[string]APIConfig) []string {
	names := make([]string, 0, len(configs))
	for name := range configs {
		names = append(names, name)
	}
	return names
}

// filterHealthyProviders removes providers with open circuit breakers
func (m *ModeManager) filterHealthyProviders(available map[string]APIConfig) map[string]APIConfig {
	healthy := make(map[string]APIConfig)
	healthyProviders := m.circuitBreakers.GetHealthyProviders()
	
	// If no circuit breaker stats exist yet, consider all providers healthy
	if len(healthyProviders) == 0 {
		return available
	}
	
	healthySet := make(map[string]bool)
	for _, provider := range healthyProviders {
		healthySet[provider] = true
	}
	
	for name, config := range available {
		if healthySet[name] {
			healthy[name] = config
		}
	}
	
	return healthy
}

// RecordProviderResult records success/failure for circuit breaker
func (m *ModeManager) RecordProviderResult(providerName string, success bool) {
	breaker := m.circuitBreakers.GetBreaker(providerName)
	breaker.RecordResult(success)
}

// ExecuteWithCircuitBreaker executes a function with circuit breaker protection
func (m *ModeManager) ExecuteWithCircuitBreaker(ctx context.Context, providerName string, fn func(ctx context.Context) error) error {
	return m.circuitBreakers.ExecuteWithBreaker(ctx, providerName, fn)
}

// GetCircuitBreakerStats returns circuit breaker statistics
func (m *ModeManager) GetCircuitBreakerStats() map[string]CircuitBreakerStats {
	return m.circuitBreakers.GetAllStats()
}

// ResetCircuitBreaker manually resets a circuit breaker
func (m *ModeManager) ResetCircuitBreaker(providerName string) {
	m.circuitBreakers.ResetBreaker(providerName)
}

// ValidateMode checks if a mode switch is possible
func (m *ModeManager) ValidateMode(mode OperationalMode) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	switch mode {
	case LOCAL_ONLY:
		if !m.hasLocalProviders() {
			return fmt.Errorf("LOCAL_ONLY mode requires local providers")
		}
	case REMOTE_ONLY:
		if !m.hasRemoteProviders() {
			return fmt.Errorf("REMOTE_ONLY mode requires remote providers")
		}
	case HYBRID:
		if !m.hasLocalProviders() && !m.hasRemoteProviders() {
			return fmt.Errorf("HYBRID mode requires at least one provider")
		}
	default:
		return fmt.Errorf("invalid mode: %v", mode)
	}

	return nil
}

// GetProviderCapabilities returns capabilities for current mode
func (m *ModeManager) GetProviderCapabilities() ProviderCapabilities {
	m.mu.RLock()
	defer m.mu.RUnlock()

	available := m.config.GetAvailableAPIs()
	
	var capabilities ProviderCapabilities
	
	switch m.currentMode {
	case LOCAL_ONLY:
		local := m.filterLocalProviders(available)
		capabilities = m.analyzeCapabilities(local)
	case REMOTE_ONLY:
		remote := m.filterRemoteProviders(available)
		capabilities = m.analyzeCapabilities(remote)
	case HYBRID:
		capabilities = m.analyzeCapabilities(available)
	}

	return capabilities
}

// ProviderCapabilities describes what the current mode can handle
type ProviderCapabilities struct {
	SupportsText      bool     `json:"supports_text"`
	SupportsImages    bool     `json:"supports_images"`
	SupportsStreaming bool     `json:"supports_streaming"`
	MaxTokens         int      `json:"max_tokens"`
	AvailableModels   []string `json:"available_models"`
	CostPerKTokens    float64  `json:"cost_per_k_tokens"`
	AverageLatency    float64  `json:"average_latency_ms"`
}

// analyzeCapabilities examines provider configs to determine capabilities
func (m *ModeManager) analyzeCapabilities(providers map[string]APIConfig) ProviderCapabilities {
	caps := ProviderCapabilities{
		SupportsText:      true,
		SupportsImages:    false,
		SupportsStreaming: false,
		MaxTokens:         0,
		AvailableModels:   []string{},
		CostPerKTokens:    0.0,
		AverageLatency:    100.0,
	}

	for _, config := range providers {
		if config.MaxTokens > caps.MaxTokens {
			caps.MaxTokens = config.MaxTokens
		}
		caps.AvailableModels = append(caps.AvailableModels, config.Models...)
	}

	return caps
}