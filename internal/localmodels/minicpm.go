package localmodels

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// MiniCPMModel implements the MiniCPM-V-4 multimodal model
type MiniCPMModel struct {
	config        ModelConfig
	binaryPath    string
	modelPath     string
	projectorPath string
	isLoaded      bool
	lastUsed      time.Time
}

// NewMiniCPMModel creates a new MiniCPM-V-4 model wrapper
func NewMiniCPMModel(config ModelConfig) (*MiniCPMModel, error) {
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

	return &MiniCPMModel{
		config:        config,
		binaryPath:    config.BinaryPath,
		modelPath:     config.ModelPath,
		projectorPath: config.ProjectorPath,
		isLoaded:      false,
	}, nil
}

// Load initializes the model (stateless, so just validates)
func (m *MiniCPMModel) Load(ctx context.Context) error {
	log.Printf("MiniCPM-V-4: Preparing model for use")

	// For stateless models, we just mark as loaded
	// The actual loading happens per-inference
	m.isLoaded = true
	m.lastUsed = time.Now()

	log.Printf("✅ MiniCPM-V-4 ready for inference")
	return nil
}

// Unload releases model resources
func (m *MiniCPMModel) Unload(ctx context.Context) error {
	log.Printf("MiniCPM-V-4: Releasing model resources")
	m.isLoaded = false
	return nil
}

// IsLoaded returns whether the model is ready for inference
func (m *MiniCPMModel) IsLoaded() bool {
	return m.isLoaded
}

// Predict performs inference using the model
func (m *MiniCPMModel) Predict(ctx context.Context, input ModelInput) (*ModelOutput, error) {
	if !m.isLoaded {
		return nil, fmt.Errorf("model not loaded")
	}

	startTime := time.Now()
	m.lastUsed = startTime

	// Build command arguments for llama-mtmd-cli
	args := m.buildCommandArgs(input)

	log.Printf("MiniCPM-V-4: Running inference with %d args", len(args))

	// Execute the command
	cmd := exec.CommandContext(ctx, m.binaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("inference failed: %w\nStderr: %s", err, stderr.String())
	}

	output := strings.TrimSpace(stdout.String())

	// Check for errors in output
	if strings.Contains(output, "error") || strings.Contains(output, "failed") {
		return nil, fmt.Errorf("model inference error: %s", output)
	}

	// Calculate processing time
	processingTime := time.Since(startTime)

	// Estimate token usage (simplified)
	promptTokens := len(strings.Fields(input.Text))
	completionTokens := len(strings.Fields(output))

	log.Printf("MiniCPM-V-4: Inference completed in %v", processingTime)

	return &ModelOutput{
		Text:           output,
		ProcessingTime: processingTime,
		TokensUsed:     promptTokens + completionTokens,
		Metadata: map[string]string{
			"model_name":     m.config.Name,
			"model_type":     string(m.config.Type),
			"inference_time": processingTime.String(),
			"has_image":      strconv.FormatBool(len(input.ImagePaths) > 0),
		},
	}, nil
}

// buildCommandArgs constructs the command line arguments for llama-mtmd-cli
func (m *MiniCPMModel) buildCommandArgs(input ModelInput) []string {
	// Base arguments following your example
	args := []string{
		"--offline",
		"--mmproj", m.projectorPath,
		"-m", m.modelPath,
		"-ngl", "20", // GPU layers - will be adjusted by memory manager
		"-fa",        // Flash attention
		"-b", "2048", // Batch size (smaller for 4B model)
		"-t", "16", // Threads
		"-p", input.Text,
		"--temp", fmt.Sprintf("%.2f", getTemperature(input.Temperature)),
		"--ctx-size", "8192",
		"-np", "16", // Parallel processing
		"--prio-batch", "2", // Priority batch
		"--no-mmproj-offload", // Keep projector on GPU
		"--ignore-eos",        // Don't stop at end-of-sequence
	}

	// Add image if provided
	if len(input.ImagePaths) > 0 {
		args = append(args, "--image", input.ImagePaths[0])
	}

	// Add max tokens if specified
	if input.MaxTokens > 0 {
		args = append(args, "--n-predict", strconv.Itoa(input.MaxTokens))
	} else {
		args = append(args, "--n-predict", "-2") // Use default
	}

	return args
}

// GetName returns the model name
func (m *MiniCPMModel) GetName() string {
	return m.config.Name
}

// GetType returns the model type
func (m *MiniCPMModel) GetType() ModelType {
	return m.config.Type
}

// GetMemoryUsage estimates memory usage in MB
func (m *MiniCPMModel) GetMemoryUsage() uint64 {
	if !m.isLoaded {
		return 0
	}
	// Estimate for MiniCPM-V-4: model + projector
	// Based on file sizes: ggml-model-Q8_0.gguf (~3.6GB) + ggml-mmproj-model-f16.gguf (~0.9GB)
	return 4500 // ~4.5GB estimated usage
}

// getTemperature returns temperature value with default
func getTemperature(temp float64) float64 {
	if temp > 0 {
		return temp
	}
	return 0.5 // Default temperature
}
