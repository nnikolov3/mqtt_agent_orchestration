package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	TopP        float64   `json:"top_p,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// AIClient provides a unified interface to multiple AI APIs
type AIClient struct {
	config     *AIHelperConfig
	httpClient *http.Client
}

// NewAIClient creates a new AI client
func NewAIClient(configPath string) (*AIClient, error) {
	config, err := LoadAIHelperConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load AI config: %w", err)
	}

	client := &AIClient{
		config: config,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}

	return client, nil
}

// GenerateResponse generates a response using the best available AI API
func (c *AIClient) GenerateResponse(ctx context.Context, messages []Message, taskComplexity string) (string, error) {
	provider, apiConfig, err := c.config.GetPreferredAPI(taskComplexity)
	if err != nil {
		return "", fmt.Errorf("no AI API available: %w", err)
	}

	return c.generateWithProvider(ctx, provider, apiConfig, messages)
}

// GenerateWithProvider generates a response using a specific provider
func (c *AIClient) GenerateWithProvider(ctx context.Context, provider string, messages []Message) (string, error) {
	available := c.config.GetAvailableAPIs()
	apiConfig, exists := available[provider]
	if !exists {
		return "", fmt.Errorf("provider %s not available", provider)
	}

	return c.generateWithProvider(ctx, provider, apiConfig, messages)
}

// generateWithProvider handles the actual API call
func (c *AIClient) generateWithProvider(ctx context.Context, provider string, apiConfig APIConfig, messages []Message) (string, error) {
	// Retry logic
	var lastErr error
	for attempt := 0; attempt <= c.config.Defaults.RetryCount; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(c.config.Defaults.GetRetryDelay()):
				// Continue to retry
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		result, err := c.callAPI(ctx, provider, apiConfig, messages)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Try next model in the list
		if len(apiConfig.Models) > attempt+1 {
			continue
		}
	}

	return "", fmt.Errorf("all attempts failed for provider %s: %w", provider, lastErr)
}

// callAPI makes the actual HTTP request to the AI API
func (c *AIClient) callAPI(ctx context.Context, provider string, apiConfig APIConfig, messages []Message) (string, error) {
	// Select the first available model for this attempt
	if len(apiConfig.Models) == 0 {
		return "", fmt.Errorf("no models configured for provider %s", provider)
	}

	model := apiConfig.Models[0]

	// Create request
	request := ChatRequest{
		Model:       model,
		Messages:    messages,
		MaxTokens:   apiConfig.MaxTokens,
		Temperature: apiConfig.Temperature,
		TopP:        apiConfig.TopP,
		Stream:      false,
	}

	// Marshal request
	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	apiURL := apiConfig.APIURL
	if strings.Contains(apiURL, "{model}") {
		apiURL = strings.ReplaceAll(apiURL, "{model}", model)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	apiKey := apiConfig.GetAPIKey()
	if apiKey == "" {
		return "", fmt.Errorf("API key not found for provider %s", provider)
	}

	// Different providers use different auth headers
	switch provider {
	case "cerebras", "groq", "grok":
		req.Header.Set("Authorization", "Bearer "+apiKey)
	case "nvidia":
		req.Header.Set("Authorization", "Bearer "+apiKey)
	case "gemini":
		// Gemini uses API key as query parameter
		if !strings.Contains(apiURL, "key=") {
			separator := "?"
			if strings.Contains(apiURL, "?") {
				separator = "&"
			}
			apiURL += separator + "key=" + apiKey
			req.URL.RawQuery = "key=" + apiKey
		}
	}

	// Set timeout
	client := &http.Client{
		Timeout: apiConfig.GetTimeout(),
	}

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)
	if content == "" {
		return "", fmt.Errorf("empty response content")
	}

	return content, nil
}

// GetAvailableProviders returns list of available AI providers
func (c *AIClient) GetAvailableProviders() []string {
	available := c.config.GetAvailableAPIs()
	providers := make([]string, 0, len(available))
	for name := range available {
		providers = append(providers, name)
	}
	return providers
}

// IsProviderAvailable checks if a specific provider is available
func (c *AIClient) IsProviderAvailable(provider string) bool {
	available := c.config.GetAvailableAPIs()
	_, exists := available[provider]
	return exists
}
