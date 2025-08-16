package localmodels

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// LoRAConfig represents LoRA training configuration
type LoRAConfig struct {
	ModelPath       string  `yaml:"model_path"`
	TrainingData    string  `yaml:"training_data"`
	OutputPath      string  `yaml:"output_path"`
	LoRAR           int     `yaml:"lora_r"`           // LoRA rank
	LoRAAlpha       int     `yaml:"lora_alpha"`       // LoRA alpha
	BatchSize       int     `yaml:"batch_size"`
	GradAccSteps    int     `yaml:"grad_acc_steps"`
	LearningRate    float64 `yaml:"learning_rate"`
	Epochs          int     `yaml:"epochs"`
	WarmupSteps     int     `yaml:"warmup_steps"`
	SaveSteps       int     `yaml:"save_steps"`
}

// TrainingExample represents a single training example
type TrainingExample struct {
	Input  string `json:"input"`
	Output string `json:"output"`
	Score  float64 `json:"score,omitempty"` // For reinforcement learning
}

// LoRATrainer handles LoRA fine-tuning operations
type LoRATrainer struct {
	llamaFinetunePath string
	llamaExportPath   string
	workingDir        string
}

// NewLoRATrainer creates a new LoRA trainer
func NewLoRATrainer(llamaBinPath, workingDir string) *LoRATrainer {
	return &LoRATrainer{
		llamaFinetunePath: filepath.Join(llamaBinPath, "llama-finetune"),
		llamaExportPath:   filepath.Join(llamaBinPath, "llama-export-lora"),
		workingDir:        workingDir,
	}
}

// PrepareTrainingData converts training examples to llama-finetune format
func (lt *LoRATrainer) PrepareTrainingData(examples []TrainingExample, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create training data file: %w", err)
	}
	defer file.Close()

	for _, example := range examples {
		// Convert to llama-finetune JSONL format
		entry := map[string]interface{}{
			"text": fmt.Sprintf("### Instruction:\n%s\n\n### Response:\n%s", example.Input, example.Output),
		}
		
		// Add reinforcement learning score if available
		if example.Score > 0 {
			entry["score"] = example.Score
		}

		data, err := json.Marshal(entry)
		if err != nil {
			return fmt.Errorf("failed to marshal training example: %w", err)
		}

		if _, err := file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("failed to write training example: %w", err)
		}
	}

	log.Printf("Prepared %d training examples in %s", len(examples), outputPath)
	return nil
}

// TrainLoRA performs LoRA fine-tuning using llama-finetune
func (lt *LoRATrainer) TrainLoRA(ctx context.Context, config LoRAConfig) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(config.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build llama-finetune command
	args := []string{
		"--model", config.ModelPath,
		"--train-data", config.TrainingData,
		"--lora-out", config.OutputPath,
		"--lora-r", fmt.Sprintf("%d", config.LoRAR),
		"--lora-alpha", fmt.Sprintf("%d", config.LoRAAlpha),
		"--batch-size", fmt.Sprintf("%d", config.BatchSize),
		"--grad-acc", fmt.Sprintf("%d", config.GradAccSteps),
		"--learning-rate", fmt.Sprintf("%.6f", config.LearningRate),
		"--epochs", fmt.Sprintf("%d", config.Epochs),
		"--warmup-steps", fmt.Sprintf("%d", config.WarmupSteps),
		"--save-steps", fmt.Sprintf("%d", config.SaveSteps),
		"--verbose",
	}

	log.Printf("Starting LoRA training with command: %s %v", lt.llamaFinetunePath, args)

	cmd := exec.CommandContext(ctx, lt.llamaFinetunePath, args...)
	cmd.Dir = lt.workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		return fmt.Errorf("LoRA training failed after %v: %w", duration, err)
	}

	log.Printf("LoRA training completed successfully in %v, output: %s", duration, config.OutputPath)
	return nil
}

// ExportMergedModel exports a merged model with LoRA weights
func (lt *LoRATrainer) ExportMergedModel(ctx context.Context, baseModelPath, loraPath, outputPath string) error {
	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	args := []string{
		"--model", baseModelPath,
		"--lora", loraPath,
		"--output", outputPath,
	}

	log.Printf("Exporting merged model with command: %s %v", lt.llamaExportPath, args)

	cmd := exec.CommandContext(ctx, lt.llamaExportPath, args...)
	cmd.Dir = lt.workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	if err != nil {
		return fmt.Errorf("model export failed after %v: %w", duration, err)
	}

	log.Printf("Model export completed successfully in %v, output: %s", duration, outputPath)
	return nil
}

// DefaultLoRAConfig returns a sensible default LoRA configuration
func DefaultLoRAConfig() LoRAConfig {
	return LoRAConfig{
		LoRAR:        16,   // Common LoRA rank
		LoRAAlpha:    32,   // 2x the rank
		BatchSize:    4,    // Small batch for GPU memory
		GradAccSteps: 4,    // Effective batch size: 4 * 4 = 16
		LearningRate: 1e-4, // Conservative learning rate
		Epochs:       3,    // Few epochs to avoid overfitting
		WarmupSteps:  100,  // Warmup for stable training
		SaveSteps:    500,  // Save checkpoints frequently
	}
}

// ValidateLoRABinaries checks if required llama.cpp binaries exist
func (lt *LoRATrainer) ValidateLoRABinaries() error {
	binaries := []string{lt.llamaFinetunePath, lt.llamaExportPath}
	
	for _, binary := range binaries {
		if _, err := os.Stat(binary); os.IsNotExist(err) {
			return fmt.Errorf("required binary not found: %s", binary)
		}
	}
	
	return nil
}

// GetLoRAInfo returns information about available LoRA adapters
func GetLoRAInfo(adapterDir string) ([]string, error) {
	var adapters []string
	
	err := filepath.Walk(adapterDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if filepath.Ext(path) == ".bin" && info.Mode().IsRegular() {
			relPath, err := filepath.Rel(adapterDir, path)
			if err != nil {
				return err
			}
			adapters = append(adapters, relPath)
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to scan LoRA adapter directory: %w", err)
	}
	
	return adapters, nil
}