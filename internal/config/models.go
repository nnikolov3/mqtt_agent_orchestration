package config

import (
	"fmt"
	"os"
	"time"

	"github.com/niko/mqtt-agent-orchestration/internal/localmodels"
	"gopkg.in/yaml.v3"
)

// ModelConfig represents the complete model configuration structure
type ModelConfig struct {
	Models      map[string]localmodels.ModelConfig `yaml:"models"`
	Manager     ManagerConfig                      `yaml:"manager"`
	Fallback    FallbackConfig                     `yaml:"fallback"`
	Performance PerformanceConfig                  `yaml:"performance"`
}

// ManagerConfig holds manager-level configuration
type ManagerConfig struct {
	MaxGPUMemory    uint64        `yaml:"max_gpu_memory"`
	NvidiaSMIPath   string        `yaml:"nvidia_smi_path"`
	MonitorInterval time.Duration `yaml:"monitor_interval"`
}

// FallbackConfig holds fallback configuration
type FallbackConfig struct {
	EnableExternalAI    bool   `yaml:"enable_external_ai"`
	PreferredExternalAI string `yaml:"preferred_external_ai"`
	MaxRetries          int    `yaml:"max_retries"`
	RetryDelay          string `yaml:"retry_delay"`
}

// PerformanceConfig holds performance-related settings
type PerformanceConfig struct {
	EnableBatching     bool   `yaml:"enable_batching"`
	MaxBatchSize       int    `yaml:"max_batch_size"`
	EnableContextReuse bool   `yaml:"enable_context_reuse"`
	MaxContextAge      string `yaml:"max_context_age"`
}

// LoadModelConfig loads model configuration from a YAML file
func LoadModelConfig(configPath string) (*ModelConfig, error) {
	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("model configuration file not found: %s", configPath)
	}

	// Read configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read model configuration: %w", err)
	}

	// Parse YAML
	var config ModelConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse model configuration: %w", err)
	}

	// Validate configuration
	if err := validateModelConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid model configuration: %w", err)
	}

	return &config, nil
}

// validateModelConfig validates the model configuration
func validateModelConfig(config *ModelConfig) error {
	// Validate models section
	if len(config.Models) == 0 {
		return fmt.Errorf("no models defined in configuration")
	}

	// Validate each model configuration
	for name, modelConfig := range config.Models {
		if err := validateModelConfigEntry(name, modelConfig); err != nil {
			return err
		}
	}

	// Validate manager configuration
	if config.Manager.MaxGPUMemory == 0 {
		return fmt.Errorf("max_gpu_memory must be greater than 0")
	}

	if config.Manager.MonitorInterval == 0 {
		config.Manager.MonitorInterval = 30 * time.Second // Default value
	}

	// Validate fallback configuration
	if config.Fallback.MaxRetries < 0 {
		return fmt.Errorf("max_retries must be non-negative")
	}

	// Validate performance configuration
	if config.Performance.MaxBatchSize <= 0 {
		config.Performance.MaxBatchSize = 4 // Default value
	}

	return nil
}

// validateModelConfigEntry validates a single model configuration entry
func validateModelConfigEntry(name string, config localmodels.ModelConfig) error {
	if config.Name == "" {
		return fmt.Errorf("model %s: name is required", name)
	}

	if config.ModelPath == "" {
		return fmt.Errorf("model %s: model_path is required", name)
	}

	if config.Type == "" {
		return fmt.Errorf("model %s: type is required", name)
	}

	if config.MemoryLimit == 0 {
		return fmt.Errorf("model %s: memory_limit must be greater than 0", name)
	}

	// Check if model file exists
	if _, err := os.Stat(config.ModelPath); os.IsNotExist(err) {
		return fmt.Errorf("model %s: model file not found: %s", name, config.ModelPath)
	}

	return nil
}

// GetModelManagerConfig converts the configuration to ModelManagerConfig
func (mc *ModelConfig) GetModelManagerConfig() localmodels.ModelManagerConfig {
	return localmodels.ModelManagerConfig{
		MaxGPUMemory:    mc.Manager.MaxGPUMemory,
		NvidiaSMIPath:   mc.Manager.NvidiaSMIPath,
		MonitorInterval: mc.Manager.MonitorInterval,
		Models:          mc.Models,
	}
}

// GetModelConfig returns configuration for a specific model
func (mc *ModelConfig) GetModelConfig(modelName string) (localmodels.ModelConfig, bool) {
	config, exists := mc.Models[modelName]
	return config, exists
}

// ListAvailableModels returns all available model names
func (mc *ModelConfig) ListAvailableModels() []string {
	var models []string
	for name := range mc.Models {
		models = append(models, name)
	}
	return models
}
