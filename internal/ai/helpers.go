package ai

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// HelperType represents different types of AI helpers
type HelperType string

const (
	HelperCerebras HelperType = "cerebras" // Code analysis and generation
	HelperNvidia   HelperType = "nvidia"   // Multimodal with OCR
	HelperGemini   HelperType = "gemini"   // Comprehensive multimodal
	HelperGrok     HelperType = "grok"     // Creative solutions
	HelperGroq     HelperType = "groq"     // Ultra-fast inference
)

// HelperConfig holds configuration for AI helpers
type HelperConfig struct {
	Type        HelperType    `yaml:"type"`
	ScriptPath  string        `yaml:"script_path"`
	Description string        `yaml:"description"`
	MaxTokens   int           `yaml:"max_tokens"`
	Timeout     time.Duration `yaml:"timeout"`
	Priority    int           `yaml:"priority"` // Lower number = higher priority
}

// HelperRequest represents a request to an AI helper
type HelperRequest struct {
	Prompt     string     `json:"prompt"`
	InputFile  string     `json:"input_file,omitempty"`
	ImageFile  string     `json:"image_file,omitempty"`
	HelperType HelperType `json:"helper_type"`
}

// HelperResponse represents a response from an AI helper
type HelperResponse struct {
	Content        string        `json:"content"`
	HelperType     HelperType    `json:"helper_type"`
	ProcessingTime time.Duration `json:"processing_time"`
	Error          string        `json:"error,omitempty"`
}

// HelperManager manages AI helper interactions
type HelperManager struct {
	configs map[HelperType]HelperConfig
}

// NewHelperManager creates a new helper manager
func NewHelperManager() *HelperManager {
	return &HelperManager{
		configs: map[HelperType]HelperConfig{
			HelperCerebras: {
				Type:        HelperCerebras,
				ScriptPath:  "/home/niko/.claude/cerebras_code_analyzer",
				Description: "Fast code analysis, review, and generation",
				MaxTokens:   4000,
				Timeout:     60 * time.Second,
				Priority:    1,
			},
			HelperNvidia: {
				Type:        HelperNvidia,
				ScriptPath:  "/home/niko/.claude/nvidia_enhance_helper",
				Description: "Multimodal analysis including OCR",
				MaxTokens:   65536,
				Timeout:     90 * time.Second,
				Priority:    2,
			},
			HelperGemini: {
				Type:        HelperGemini,
				ScriptPath:  "/home/niko/.claude/gemini_code_analyzer",
				Description: "Comprehensive multimodal analysis",
				MaxTokens:   8192,
				Timeout:     120 * time.Second,
				Priority:    3,
			},
			HelperGrok: {
				Type:        HelperGrok,
				ScriptPath:  "/home/niko/.claude/grok_code_helper",
				Description: "Creative solutions and multimodal analysis",
				MaxTokens:   8192,
				Timeout:     120 * time.Second,
				Priority:    4,
			},
			HelperGroq: {
				Type:        HelperGroq,
				ScriptPath:  "/home/niko/.claude/groq_fast_analyzer",
				Description: "Ultra-fast inference for speed-critical tasks",
				MaxTokens:   4096,
				Timeout:     30 * time.Second,
				Priority:    5,
			},
		},
	}
}

// GetHelperForTask determines the best helper for a given task
func (hm *HelperManager) GetHelperForTask(taskType string, complexity string, hasImages bool) HelperType {
	// Simple tasks use local models, complex tasks use helpers
	if complexity == "small" {
		return "" // Use local models
	}

	// Route based on task type and requirements
	switch {
	case strings.Contains(taskType, "code") && !hasImages:
		return HelperCerebras // Best for code analysis
	case hasImages && strings.Contains(taskType, "ocr"):
		return HelperNvidia // Best for OCR and visual tasks
	case hasImages && complexity == "high":
		return HelperGemini // Best for complex multimodal
	case strings.Contains(taskType, "creative") || strings.Contains(taskType, "solution"):
		return HelperGrok // Best for creative solutions
	case complexity == "medium" && !hasImages:
		return HelperGroq // Fast inference for medium tasks
	default:
		return HelperGemini // Default to comprehensive analysis
	}
}

// ExecuteHelper runs an AI helper script
func (hm *HelperManager) ExecuteHelper(ctx context.Context, req HelperRequest) (*HelperResponse, error) {
	config, exists := hm.configs[req.HelperType]
	if !exists {
		return nil, fmt.Errorf("unknown helper type: %s", req.HelperType)
	}

	start := time.Now()

	// Build command arguments
	args := []string{req.Prompt}
	if req.InputFile != "" {
		args = append(args, req.InputFile)
	}
	if req.ImageFile != "" {
		args = append(args, req.ImageFile)
	}

	// Execute the helper script
	cmd := exec.CommandContext(ctx, config.ScriptPath, args...)
	output, err := cmd.Output()

	processingTime := time.Since(start)

	response := &HelperResponse{
		HelperType:     req.HelperType,
		ProcessingTime: processingTime,
	}

	if err != nil {
		response.Error = fmt.Sprintf("helper execution failed: %v", err)
		return response, err
	}

	response.Content = strings.TrimSpace(string(output))
	return response, nil
}

// GetAvailableHelpers returns all available helper types
func (hm *HelperManager) GetAvailableHelpers() []HelperType {
	helpers := make([]HelperType, 0, len(hm.configs))
	for helperType := range hm.configs {
		helpers = append(helpers, helperType)
	}
	return helpers
}

// GetHelperConfig returns configuration for a specific helper
func (hm *HelperManager) GetHelperConfig(helperType HelperType) (HelperConfig, bool) {
	config, exists := hm.configs[helperType]
	return config, exists
}
