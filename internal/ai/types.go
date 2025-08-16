package ai

import (
	"context"
	"time"
)

// Provider is the minimal contract every AI backend must satisfy
type Provider interface {
	// Name returns a short, human readable identifier (e.g. "local", "grok", "gemini")
	Name() string

	// ModelInfo returns static metadata for a model name
	ModelInfo(model string) (ModelInfo, error)

	// Generate runs inference with context for timeout/cancellation
	Generate(ctx context.Context, req Request) (Response, error)

	// IsAvailable checks if the provider is currently operational
	IsAvailable(ctx context.Context) bool

	// SupportedModels returns list of models this provider can handle
	SupportedModels() []string
}

// Request represents an AI inference request
type Request struct {
	// Primary prompt text
	Prompt string `json:"prompt"`

	// Model identifier (e.g. "minicpm-v-4", "grok-4", "gemini-2.5-pro")
	Model string `json:"model"`

	// Optional image paths for multimodal models
	ImagePaths []string `json:"image_paths,omitempty"`

	// Generation parameters
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`

	// Request metadata
	RequestID string    `json:"request_id"`
	TenantID  string    `json:"tenant_id"`  // For cost tracking
	CreatedAt time.Time `json:"created_at"`

	// MQTT context
	MQTTTopic     string `json:"mqtt_topic,omitempty"`
	CorrelationID string `json:"correlation_id,omitempty"`
}

// Response represents an AI inference response
type Response struct {
	// Generated content
	Content string `json:"content"`

	// Response metadata
	RequestID  string    `json:"request_id"`
	Model      string    `json:"model"`
	Provider   string    `json:"provider"`
	CreatedAt  time.Time `json:"created_at"`
	FinishedAt time.Time `json:"finished_at"`

	// Performance metrics
	Latency time.Duration `json:"latency"`
	Usage   TokenUsage    `json:"usage"`

	// Raw response for debugging
	Raw interface{} `json:"raw,omitempty"`

	// Error information if any
	Error string `json:"error,omitempty"`
}

// TokenUsage holds token counts and cost information
type TokenUsage struct {
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	TotalTokens  int     `json:"total_tokens"`
	CostUSD      float64 `json:"cost_usd"`
}

// ModelInfo contains metadata about a specific model
type ModelInfo struct {
	// Display name for UI/logs
	DisplayName string `json:"display_name"`

	// Model type: "text", "multimodal", "embedding"
	Type string `json:"type"`

	// Provider category: "local", "remote"
	Category string `json:"category"`

	// Pricing information
	PricePerKInputTokens  float64 `json:"price_per_k_input_tokens"`
	PricePerKOutputTokens float64 `json:"price_per_k_output_tokens"`

	// Technical limits
	MaxTokens        int  `json:"max_tokens"`
	SupportsImages   bool `json:"supports_images"`
	SupportsStreaming bool `json:"supports_streaming"`

	// Resource requirements (for local models)
	MemoryRequiredMB uint64 `json:"memory_required_mb"`
}

// ProviderConfig holds configuration for AI providers
type ProviderConfig struct {
	// Provider identification
	Name     string `yaml:"name"`
	Type     string `yaml:"type"` // "local", "grok", "gemini"
	Enabled  bool   `yaml:"enabled"`
	Priority int    `yaml:"priority"` // Lower number = higher priority

	// Authentication
	APIKey  string `yaml:"api_key,omitempty"`
	BaseURL string `yaml:"base_url,omitempty"`

	// Timeouts and limits
	TimeoutSeconds     int `yaml:"timeout_seconds"`
	MaxConcurrency     int `yaml:"max_concurrency"`
	RateLimitPerMinute int `yaml:"rate_limit_per_minute"`

	// Local model specific
	BinaryPath  string            `yaml:"binary_path,omitempty"`
	ModelPath   string            `yaml:"model_path,omitempty"`
	Models      map[string]string `yaml:"models,omitempty"` // model_name -> file_path

	// Cost controls
	MaxCostPerRequest float64 `yaml:"max_cost_per_request"`
	DailyCostLimit    float64 `yaml:"daily_cost_limit"`
}

// ManagerConfig holds configuration for the AI manager
type ManagerConfig struct {
	// Provider configurations
	Providers []ProviderConfig `yaml:"providers"`

	// Fallback behavior
	EnableFallback      bool          `yaml:"enable_fallback"`
	FallbackTimeout     time.Duration `yaml:"fallback_timeout"`
	MaxRetries          int           `yaml:"max_retries"`
	RetryDelay          time.Duration `yaml:"retry_delay"`

	// Cost tracking
	EnableCostTracking bool   `yaml:"enable_cost_tracking"`
	CostStorePath      string `yaml:"cost_store_path"`

	// Monitoring
	EnableMetrics   bool   `yaml:"enable_metrics"`
	MetricsAddr     string `yaml:"metrics_addr"`
	HealthCheckAddr string `yaml:"health_check_addr"`

	// MQTT integration
	MQTTBrokerURL   string `yaml:"mqtt_broker_url"`
	MQTTTopicPrefix string `yaml:"mqtt_topic_prefix"`
}

// CostEntry represents a cost tracking record
type CostEntry struct {
	Timestamp time.Time `json:"timestamp"`
	TenantID  string    `json:"tenant_id"`
	Provider  string    `json:"provider"`
	Model     string    `json:"model"`
	RequestID string    `json:"request_id"`
	
	InputTokens  int     `json:"input_tokens"`
	OutputTokens int     `json:"output_tokens"`
	CostUSD      float64 `json:"cost_usd"`
	
	Latency time.Duration `json:"latency"`
	Success bool          `json:"success"`
}

// ProviderStats holds statistics for a provider
type ProviderStats struct {
	Name             string        `json:"name"`
	IsAvailable      bool          `json:"is_available"`
	RequestCount     int64         `json:"request_count"`
	SuccessCount     int64         `json:"success_count"`
	ErrorCount       int64         `json:"error_count"`
	TotalCostUSD     float64       `json:"total_cost_usd"`
	AverageLatency   time.Duration `json:"average_latency"`
	LastRequestTime  time.Time     `json:"last_request_time"`
	SupportedModels  []string      `json:"supported_models"`
}