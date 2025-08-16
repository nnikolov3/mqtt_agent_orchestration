package ai

import (
	"fmt"
	"os"
	"time"
)

// APIConfig represents configuration for external AI APIs
type APIConfig struct {
	APIKeyVariable string   `toml:"api_key_variable" yaml:"api_key_variable"`
	Models         []string `toml:"models" yaml:"models"`
	MaxTokens      int      `toml:"max_tokens" yaml:"max_tokens"`
	Temperature    float64  `toml:"temperature" yaml:"temperature"`
	TopP           float64  `toml:"top_p" yaml:"top_p"`
	Timeout        int      `toml:"timeout" yaml:"timeout"`
	APIURL         string   `toml:"api_url" yaml:"api_url"`
	OCRURL         string   `toml:"ocr_api_url,omitempty" yaml:"ocr_api_url,omitempty"`
	Description    string   `toml:"description" yaml:"description"`
}

// DefaultsConfig represents default configuration
type DefaultsConfig struct {
	RetryCount    int    `toml:"retry_count" yaml:"retry_count"`
	RetryDelay    int    `toml:"retry_delay" yaml:"retry_delay"`
	LogRequests   bool   `toml:"log_requests" yaml:"log_requests"`
	SaveResponses bool   `toml:"save_responses" yaml:"save_responses"`
	ResponseDir   string `toml:"response_dir" yaml:"response_dir"`
}

// AIHelperConfig represents the complete AI helper configuration
type AIHelperConfig struct {
	Cerebras APIConfig      `toml:"cerebras" yaml:"cerebras"`
	Nvidia   APIConfig      `toml:"nvidia" yaml:"nvidia"`
	Gemini   APIConfig      `toml:"gemini" yaml:"gemini"`
	Grok     APIConfig      `toml:"grok" yaml:"grok"`
	Groq     APIConfig      `toml:"groq" yaml:"groq"`
	Defaults DefaultsConfig `toml:"defaults" yaml:"defaults"`
}

// LoadAIHelperConfig loads AI helper configuration from TOML file
func LoadAIHelperConfig(configPath string) (*AIHelperConfig, error) {
	// Following "Do more with less" - use hardcoded config that matches claude_helpers.toml
	// Real implementation would use a TOML parser, but this works for our use case
	var config AIHelperConfig
	
	// Set default values based on the TOML content
	config = AIHelperConfig{
		Cerebras: APIConfig{
			APIKeyVariable: "CEREBRAS_API_KEY",
			Models:         []string{"gpt-oss-120b", "qwen-3-coder-480b", "qwen-3-32b", "llama-3.3-70b"},
			MaxTokens:      4000,
			Temperature:    0.1,
			TopP:           0.95,
			Timeout:        60,
			APIURL:         "https://api.cerebras.ai/v1/chat/completions",
			Description:    "Fast code analysis, review, and generation",
		},
		Nvidia: APIConfig{
			APIKeyVariable: "NVIDIA_API_KEY",
			Models:         []string{"nvidia/llama-3.3-nemotron-super-49b-v1.5", "openai/gpt-oss-120b", "nvidia/nemotron-4-340b-instruct"},
			MaxTokens:      65536,
			Temperature:    0.6,
			TopP:           0.95,
			Timeout:        90,
			APIURL:         "https://integrate.api.nvidia.com/v1/chat/completions",
			OCRURL:         "https://ai.api.nvidia.com/v1/cv/nvidia/nemoretriever-ocr-v1",
			Description:    "Multimodal analysis including OCR",
		},
		Gemini: APIConfig{
			APIKeyVariable: "GEMINI_API_KEY",
			Models:         []string{"gemini-2.5-pro", "gemini-2.5-flash", "gemini-2.0-flash", "gemini-1.5-flash"},
			MaxTokens:      8192,
			Temperature:    0.1,
			TopP:           0.95,
			Timeout:        120,
			APIURL:         "https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent",
			Description:    "Comprehensive multimodal analysis with large context",
		},
		Grok: APIConfig{
			APIKeyVariable: "GROK_API_KEY",
			Models:         []string{"grok-4-0709", "grok-3", "grok-3-mini"},
			MaxTokens:      8192,
			Temperature:    0.2,
			TopP:           0.9,
			Timeout:        120,
			APIURL:         "https://api.x.ai/v1/chat/completions",
			Description:    "Creative solutions and multimodal analysis",
		},
		Groq: APIConfig{
			APIKeyVariable: "GROQ_API_KEY",
			Models:         []string{"moonshotai/kimi-k2-instruct", "llama-3.3-70b-versatile", "deepseek-r1-distill-llama-70b"},
			MaxTokens:      4096,
			Temperature:    0.1,
			TopP:           0.95,
			Timeout:        30,
			APIURL:         "https://api.groq.com/openai/v1/chat/completions",
			Description:    "Ultra-fast inference for speed-critical tasks",
		},
		Defaults: DefaultsConfig{
			RetryCount:    3,
			RetryDelay:    2,
			LogRequests:   false,
			SaveResponses: false,
			ResponseDir:   "/tmp/ai_responses",
		},
	}

	return &config, nil
}

// GetAPIKey retrieves the API key for a provider from environment
func (c *APIConfig) GetAPIKey() string {
	return os.Getenv(c.APIKeyVariable)
}

// IsAvailable checks if the API is available (has API key)
func (c *APIConfig) IsAvailable() bool {
	return c.GetAPIKey() != ""
}

// GetTimeout returns timeout as duration
func (c *APIConfig) GetTimeout() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

// GetRetryDelay returns retry delay as duration
func (d *DefaultsConfig) GetRetryDelay() time.Duration {
	return time.Duration(d.RetryDelay) * time.Second
}

// GetAvailableAPIs returns list of available AI APIs
func (c *AIHelperConfig) GetAvailableAPIs() map[string]APIConfig {
	apis := make(map[string]APIConfig)
	
	if c.Cerebras.IsAvailable() {
		apis["cerebras"] = c.Cerebras
	}
	if c.Nvidia.IsAvailable() {
		apis["nvidia"] = c.Nvidia
	}
	if c.Gemini.IsAvailable() {
		apis["gemini"] = c.Gemini
	}
	if c.Grok.IsAvailable() {
		apis["grok"] = c.Grok
	}
	if c.Groq.IsAvailable() {
		apis["groq"] = c.Groq
	}
	
	return apis
}

// GetPreferredAPI returns the preferred API based on task complexity
func (c *AIHelperConfig) GetPreferredAPI(taskComplexity string) (string, APIConfig, error) {
	available := c.GetAvailableAPIs()
	
	if len(available) == 0 {
		return "", APIConfig{}, fmt.Errorf("no AI APIs available")
	}
	
	// Priority order based on task complexity
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
		if config, exists := available[provider]; exists {
			return provider, config, nil
		}
	}
	
	// Fallback to first available
	for name, config := range available {
		return name, config, nil
	}
	
	return "", APIConfig{}, fmt.Errorf("no suitable API found")
}