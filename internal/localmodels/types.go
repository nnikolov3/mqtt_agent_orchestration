package localmodels

import (
	"context"
	"time"
)

// ModelType represents the type of local model
type ModelType string

const (
	ModelTypeText       ModelType = "text"
	ModelTypeMultimodal ModelType = "multimodal"
)

// Model interface defines common operations for all local models
type Model interface {
	Load(ctx context.Context) error
	Unload(ctx context.Context) error
	IsLoaded() bool
	Predict(ctx context.Context, input ModelInput) (*ModelOutput, error)
	GetName() string
	GetType() ModelType
	GetMemoryUsage() uint64 // Returns estimated memory usage in MB
}

// ModelConfig holds configuration for a specific model
type ModelConfig struct {
	Name          string            `yaml:"name"`
	BinaryPath    string            `yaml:"binary_path"`
	ModelPath     string            `yaml:"model_path"`
	ProjectorPath string            `yaml:"projector_path,omitempty"` // For multimodal models
	Type          ModelType         `yaml:"type"`
	MemoryLimit   uint64            `yaml:"memory_limit"` // MB
	Parameters    map[string]string `yaml:"parameters,omitempty"`
}

// ModelInput represents input to a model
type ModelInput struct {
	Text        string   `json:"text"`
	ImagePaths  []string `json:"image_paths,omitempty"` // For multimodal
	ImageData   [][]byte `json:"image_data,omitempty"`  // Base64 decoded image data
	Temperature float64  `json:"temperature,omitempty"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
}

// ModelOutput represents output from a model
type ModelOutput struct {
	Text           string            `json:"text"`
	ProcessingTime time.Duration     `json:"processing_time"`
	TokensUsed     int               `json:"tokens_used,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// GPUMemoryInfo stores GPU memory usage information
type GPUMemoryInfo struct {
	Total     uint64    `json:"total"` // Total GPU memory (MB)
	Used      uint64    `json:"used"`  // Used GPU memory (MB)
	Free      uint64    `json:"free"`  // Free GPU memory (MB)
	Timestamp time.Time `json:"timestamp"`
}

// ModelManagerConfig holds manager-level configuration
type ModelManagerConfig struct {
	MaxGPUMemory    uint64                 `yaml:"max_gpu_memory"` // MB
	NvidiaSMIPath   string                 `yaml:"nvidia_smi_path"`
	MonitorInterval time.Duration          `yaml:"monitor_interval"`
	Models          map[string]ModelConfig `yaml:"models"`
}

// LoadingState represents the current state of model loading/unloading
type LoadingState int

const (
	StateUnloaded LoadingState = iota
	StateLoading
	StateLoaded
	StateUnloading
	StateError
)

// ModelStatus represents the current status of a model
type ModelStatus struct {
	Name         string       `json:"name"`
	State        LoadingState `json:"state"`
	MemoryUsage  uint64       `json:"memory_usage"` // MB
	LastUsed     time.Time    `json:"last_used"`
	ErrorMessage string       `json:"error_message,omitempty"`
}
