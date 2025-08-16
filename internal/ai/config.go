package ai

import (
	"fmt"
	"os"
	"time"

	"github.com/BurntSushi/toml"
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
	Cerebras  APIConfig      `toml:"cerebras" yaml:"cerebras"`
	Nvidia    APIConfig      `toml:"nvidia" yaml:"nvidia"`
	NvidiaOCR APIConfig      `toml:"nvidia_ocr" yaml:"nvidia_ocr"`
	Gemini    APIConfig      `toml:"gemini" yaml:"gemini"`
	Grok      APIConfig      `toml:"grok" yaml:"grok"`
	Groq      APIConfig      `toml:"groq" yaml:"groq"`
	Defaults  DefaultsConfig `toml:"defaults" yaml:"defaults"`
}

// LoadAIHelperConfig loads AI helper configuration from TOML file
func LoadAIHelperConfig(configPath string) (*AIHelperConfig, error) {
	var config AIHelperConfig

	// Read and parse TOML file - following design principles: "Never hard code values"
	_, err := toml.DecodeFile(configPath, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to load AI helper config from %s: %w", configPath, err)
	}

	// Validate required fields
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// validateConfig ensures all required configuration fields are present
func validateConfig(config *AIHelperConfig) error {
	providers := map[string]APIConfig{
		"cerebras":   config.Cerebras,
		"nvidia":     config.Nvidia,
		"nvidia_ocr": config.NvidiaOCR,
		"gemini":     config.Gemini,
		"grok":       config.Grok,
		"groq":       config.Groq,
	}

	for name, provider := range providers {
		if provider.APIKeyVariable == "" {
			return fmt.Errorf("missing api_key_variable for %s", name)
		}
		if len(provider.Models) == 0 {
			return fmt.Errorf("missing models for %s", name)
		}
		if provider.APIURL == "" {
			return fmt.Errorf("missing api_url for %s", name)
		}
	}

	return nil
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
	if c.NvidiaOCR.IsAvailable() {
		apis["nvidia_ocr"] = c.NvidiaOCR
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
