package localmodels

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// QwenTextModel implements Qwen2.5-Omni-3B for text-only tasks
type QwenTextModel struct {
	config   ModelConfig
	isLoaded bool
	lastUsed time.Time
}

// NewQwenTextModel creates a new Qwen text model wrapper
func NewQwenTextModel(config ModelConfig) (*QwenTextModel, error) {
	// Validate paths exist
	if _, err := os.Stat(config.BinaryPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("binary not found: %s", config.BinaryPath)
	}
	if _, err := os.Stat(config.ModelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("model file not found: %s", config.ModelPath)
	}

	return &QwenTextModel{
		config:   config,
		isLoaded: false,
	}, nil
}

// Load initializes the model
func (q *QwenTextModel) Load(ctx context.Context) error {
	log.Printf("Qwen2.5-Omni-3B (Text): Preparing model for use")
	q.isLoaded = true
	q.lastUsed = time.Now()
	log.Printf("✅ Qwen2.5-Omni-3B (Text) ready for inference")
	return nil
}

// Unload releases model resources
func (q *QwenTextModel) Unload(ctx context.Context) error {
	log.Printf("Qwen2.5-Omni-3B (Text): Releasing model resources")
	q.isLoaded = false
	return nil
}

// IsLoaded returns whether the model is ready
func (q *QwenTextModel) IsLoaded() bool {
	return q.isLoaded
}

// Predict performs text inference
func (q *QwenTextModel) Predict(ctx context.Context, input ModelInput) (*ModelOutput, error) {
	if !q.isLoaded {
		return nil, fmt.Errorf("model not loaded")
	}

	startTime := time.Now()
	q.lastUsed = startTime

	// Start llama-server if not already running
	serverPort := "8082"
	serverURL := fmt.Sprintf("http://localhost:%s", serverPort)

	// Check if server is already running
	if !q.isServerRunning(serverURL) {
		log.Printf("Qwen2.5-Omni-3B (Text): Starting llama-server on port %s", serverPort)
		args := q.buildTextCommandArgs(input)
		cmd := exec.CommandContext(ctx, q.config.BinaryPath, args...)

		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start llama-server: %w", err)
		}

		// Wait for server to start
		time.Sleep(10 * time.Second)
	}

	log.Printf("Qwen2.5-Omni-3B (Text): Making inference request")

	// Make API call to llama-server
	requestBody := map[string]interface{}{
		"prompt":         input.Text,
		"n_predict":      getMaxTokens(input.MaxTokens),
		"temperature":    getTemperature(input.Temperature),
		"top_k":          40,
		"top_p":          0.9,
		"repeat_penalty": 1.1,
	}

	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request to llama-server
	resp, err := http.Post(serverURL+"/completion", "application/json", bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to make inference request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	output, ok := response["content"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid response format: %s", string(body))
	}

	processingTime := time.Since(startTime)

	// Estimate token usage
	promptTokens := len(strings.Fields(input.Text))
	completionTokens := len(strings.Fields(output))

	log.Printf("Qwen2.5-Omni-3B (Text): Inference completed in %v", processingTime)

	return &ModelOutput{
		Text:           output,
		ProcessingTime: processingTime,
		TokensUsed:     promptTokens + completionTokens,
		Metadata: map[string]string{
			"model_name":     q.config.Name,
			"model_type":     string(q.config.Type),
			"inference_time": processingTime.String(),
			"mode":           "text_only",
		},
	}, nil
}

// buildTextCommandArgs constructs arguments for text-only inference
func (q *QwenTextModel) buildTextCommandArgs(input ModelInput) []string {
	// For text-only, use llama-server for inference
	args := []string{
		"--model", q.config.ModelPath,
		"--port", "8082", // Use different port to avoid conflicts
		"-ngl", "37", // GPU layers for 3B model
		"--ctx-size", "8192",
	}

	return args
}

// isServerRunning checks if llama-server is running on the given URL
func (q *QwenTextModel) isServerRunning(serverURL string) bool {
	resp, err := http.Get(serverURL + "/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

// GetName returns the model name
func (q *QwenTextModel) GetName() string {
	return q.config.Name
}

// GetType returns the model type
func (q *QwenTextModel) GetType() ModelType {
	return q.config.Type
}

// GetMemoryUsage estimates memory usage in MB
func (q *QwenTextModel) GetMemoryUsage() uint64 {
	if !q.isLoaded {
		return 0
	}
	// Qwen2.5-Omni-3B text-only: ~3.4GB
	return 3400
}

// QwenMultimodalModel implements Qwen2.5-Omni-3B for multimodal tasks
type QwenMultimodalModel struct {
	config        ModelConfig
	projectorPath string
	isLoaded      bool
	lastUsed      time.Time
}

// NewQwenMultimodalModel creates a new Qwen multimodal model wrapper
func NewQwenMultimodalModel(config ModelConfig) (*QwenMultimodalModel, error) {
	// Validate paths exist
	if _, err := os.Stat(config.BinaryPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("binary not found: %s", config.BinaryPath)
	}
	if _, err := os.Stat(config.ModelPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("model file not found: %s", config.ModelPath)
	}
	if _, err := os.Stat(config.ProjectorPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("projector file not found: %s", config.ProjectorPath)
	}

	return &QwenMultimodalModel{
		config:        config,
		projectorPath: config.ProjectorPath,
		isLoaded:      false,
	}, nil
}

// Load initializes the multimodal model
func (q *QwenMultimodalModel) Load(ctx context.Context) error {
	log.Printf("Qwen2.5-Omni-3B (Multimodal): Preparing model for use")
	q.isLoaded = true
	q.lastUsed = time.Now()
	log.Printf("✅ Qwen2.5-Omni-3B (Multimodal) ready for inference")
	return nil
}

// Unload releases model resources
func (q *QwenMultimodalModel) Unload(ctx context.Context) error {
	log.Printf("Qwen2.5-Omni-3B (Multimodal): Releasing model resources")
	q.isLoaded = false
	return nil
}

// IsLoaded returns whether the model is ready
func (q *QwenMultimodalModel) IsLoaded() bool {
	return q.isLoaded
}

// Predict performs multimodal inference
func (q *QwenMultimodalModel) Predict(ctx context.Context, input ModelInput) (*ModelOutput, error) {
	if !q.isLoaded {
		return nil, fmt.Errorf("model not loaded")
	}

	startTime := time.Now()
	q.lastUsed = startTime

	// Build command arguments for multimodal inference
	args := q.buildMultimodalCommandArgs(input)

	log.Printf("Qwen2.5-Omni-3B (Multimodal): Running multimodal inference")

	// Execute the command using llama-mtmd-cli
	cmd := exec.CommandContext(ctx, q.config.BinaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("multimodal inference failed: %w\nStderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())

	// Check for errors in output
	if strings.Contains(output, "error") || strings.Contains(output, "failed") {
		return nil, fmt.Errorf("model inference error: %s", output)
	}

	processingTime := time.Since(startTime)

	// Estimate token usage
	promptTokens := len(strings.Fields(input.Text))
	completionTokens := len(strings.Fields(output))

	log.Printf("Qwen2.5-Omni-3B (Multimodal): Inference completed in %v", processingTime)

	return &ModelOutput{
		Text:           output,
		ProcessingTime: processingTime,
		TokensUsed:     promptTokens + completionTokens,
		Metadata: map[string]string{
			"model_name":     q.config.Name,
			"model_type":     string(q.config.Type),
			"inference_time": processingTime.String(),
			"mode":           "multimodal",
			"has_image":      strconv.FormatBool(len(input.ImagePaths) > 0),
		},
	}, nil
}

// buildMultimodalCommandArgs constructs arguments for multimodal inference
func (q *QwenMultimodalModel) buildMultimodalCommandArgs(input ModelInput) []string {
	// Based on your example for Qwen multimodal
	args := []string{
		"--offline",
		"--mmproj", q.projectorPath,
		"-m", q.config.ModelPath,
		"-ngl", "10", // Lower for 3B model
		"-fa",        // Flash attention
		"-b", "1024", // Smaller batch for 3B
		"-t", "16", // Threads
		"-p", input.Text,
		"--temp", fmt.Sprintf("%.2f", getTemperature(input.Temperature)),
		"--ctx-size", "8192",
		"-np", "16",
		"--prio-batch", "2",
		"--no-mmproj-offload",
		"--ignore-eos",
		"-ub", "4096",
		"--prio", "3",
	}

	// Add image if provided
	if len(input.ImagePaths) > 0 {
		args = append(args, "--image", input.ImagePaths[0])
	}

	// Add max tokens
	if input.MaxTokens > 0 {
		args = append(args, "--n-predict", strconv.Itoa(input.MaxTokens))
	} else {
		args = append(args, "--n-predict", "-2")
	}

	return args
}

// GetName returns the model name
func (q *QwenMultimodalModel) GetName() string {
	return q.config.Name
}

// GetType returns the model type
func (q *QwenMultimodalModel) GetType() ModelType {
	return q.config.Type
}

// GetMemoryUsage estimates memory usage in MB
func (q *QwenMultimodalModel) GetMemoryUsage() uint64 {
	if !q.isLoaded {
		return 0
	}
	// Qwen2.5-Omni-3B + projector: ~3.4GB + 1.4GB = ~4.8GB
	return 4800
}

// Helper functions
func getMaxTokens(tokens int) int {
	if tokens > 0 {
		return tokens
	}
	return 512 // Default max tokens
}
