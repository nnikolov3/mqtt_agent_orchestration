package worker

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/niko/mqtt-agent-orchestration/internal/localmodels"
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// ContentAnalyzer analyzes task content to determine optimal model routing
type ContentAnalyzer struct {
	modelConfigs map[string]localmodels.ModelConfig
}

// NewContentAnalyzer creates a new content analyzer
func NewContentAnalyzer(modelConfigs map[string]localmodels.ModelConfig) *ContentAnalyzer {
	return &ContentAnalyzer{
		modelConfigs: modelConfigs,
	}
}

// AnalysisResult contains the routing decision and reasoning
type AnalysisResult struct {
	RecommendedModel  string   `json:"recommended_model"`
	Confidence        float64  `json:"confidence"`
	Reasoning         string   `json:"reasoning"`
	AlternativeModels []string `json:"alternative_models,omitempty"`
	ContentType       string   `json:"content_type"`
	Complexity        string   `json:"complexity"`
}

// AnalyzeContent determines the best model for a given task
func (ca *ContentAnalyzer) AnalyzeContent(ctx context.Context, task *types.WorkflowTask) (*AnalysisResult, error) {
	content := ca.extractContent(task)

	// Determine content type
	contentType := ca.detectContentType(content)

	// Determine complexity
	complexity := ca.assessComplexity(content)

	// Route based on content type and complexity
	result := ca.routeTask(contentType, complexity, content, task)

	return result, nil
}

// extractContent extracts relevant content from the task
func (ca *ContentAnalyzer) extractContent(task *types.WorkflowTask) string {
	var content strings.Builder

	// Add task type
	content.WriteString(task.Type)
	content.WriteString(" ")

	// Add document type if present
	if docType, ok := task.Payload["document_type"]; ok {
		content.WriteString(docType)
		content.WriteString(" ")
	}

	// Add previous output if present
	if task.PreviousOutput != "" {
		content.WriteString(task.PreviousOutput)
		content.WriteString(" ")
	}

	// Add any additional context
	for key, value := range task.Payload {
		if key != "document_type" {
			content.WriteString(fmt.Sprintf("%s: %s ", key, value))
		}
	}

	return content.String()
}

// detectContentType determines the type of content in the task
func (ca *ContentAnalyzer) detectContentType(content string) string {
	content = strings.ToLower(content)

	// Check for image-related content
	imagePatterns := []string{
		"image", "screenshot", "ui", "interface", "visual", "picture", "photo",
		"error.*screenshot", "ui.*analysis", "interface.*review",
	}

	for _, pattern := range imagePatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return "multimodal"
		}
	}

	// Check for code-specific content
	codePatterns := []string{
		"code.*generation", "refactor", "testing", "debug", "compile",
		"function", "class", "method", "api", "syntax", "algorithm",
		"go.*code", "python.*code", "javascript.*code",
	}

	for _, pattern := range codePatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return "code"
		}
	}

	// Check for documentation content
	docPatterns := []string{
		"documentation", "readme", "guide", "manual", "tutorial",
		"explanation", "description", "overview", "summary",
	}

	for _, pattern := range docPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			return "documentation"
		}
	}

	// Default to general text
	return "general"
}

// assessComplexity determines the complexity level of the task
func (ca *ContentAnalyzer) assessComplexity(content string) string {
	content = strings.ToLower(content)

	// Count technical terms as complexity indicators
	technicalTerms := []string{
		"architecture", "design", "pattern", "algorithm", "optimization",
		"performance", "scalability", "security", "testing", "deployment",
		"integration", "api", "framework", "library", "dependency",
	}

	complexityScore := 0
	for _, term := range technicalTerms {
		if strings.Contains(content, term) {
			complexityScore++
		}
	}

	// Assess based on content length and technical terms
	if complexityScore > 5 || len(content) > 2000 {
		return "high"
	} else if complexityScore > 2 || len(content) > 500 {
		return "medium"
	}

	return "low"
}

// routeTask determines the best model based on content type and complexity
func (ca *ContentAnalyzer) routeTask(contentType, complexity, content string, task *types.WorkflowTask) *AnalysisResult {
	result := &AnalysisResult{
		ContentType: contentType,
		Complexity:  complexity,
	}

	switch contentType {
	case "multimodal":
		// Qwen-VL is best for multimodal tasks
		result.RecommendedModel = "qwen-vl"
		result.Confidence = 0.95
		result.Reasoning = "Multimodal content detected - Qwen-VL specializes in image+text tasks"
		result.AlternativeModels = []string{"llava", "mimo"}

	case "code":
		// Qwen-omni for code tasks (no CodeLlama available)
		result.RecommendedModel = "qwen-omni"
		result.Confidence = 0.85
		result.Reasoning = "Code task - Qwen-omni provides good balance of code and general capabilities"
		result.AlternativeModels = []string{"qwen-omni"}

	case "documentation":
		// Qwen-omni for documentation tasks
		result.RecommendedModel = "qwen-omni"
		result.Confidence = 0.90
		result.Reasoning = "Documentation task - Qwen-omni excels at text generation and explanation"
		result.AlternativeModels = []string{"qwen-omni"}

	case "general":
		// Route based on complexity
		switch complexity {
		case "high":
			result.RecommendedModel = "qwen-omni"
			result.Confidence = 0.85
			result.Reasoning = "High-complexity general task - Qwen-omni provides comprehensive capabilities"
		case "medium":
			result.RecommendedModel = "qwen-omni"
			result.Confidence = 0.80
			result.Reasoning = "Medium-complexity general task - Qwen-omni provides good balance"
		case "low":
			result.RecommendedModel = "qwen-omni"
			result.Confidence = 0.90
			result.Reasoning = "Low-complexity task - Qwen-omni is efficient for simple tasks"
		}
		result.AlternativeModels = []string{"qwen-omni"}

	default:
		// Fallback to Qwen-omni
		result.RecommendedModel = "qwen-omni"
		result.Confidence = 0.70
		result.Reasoning = "Unknown content type - using general-purpose Qwen-omni as fallback"
		result.AlternativeModels = []string{"qwen-omni"}
	}

	// Adjust for worker role if needed
	result = ca.adjustForWorkerRole(result, task.RequiredRole)

	return result
}

// adjustForWorkerRole adjusts the model recommendation based on worker role
func (ca *ContentAnalyzer) adjustForWorkerRole(result *AnalysisResult, role types.WorkerRole) *AnalysisResult {
	switch role {
	case types.RoleTester:
		// Testers often need quick, simple analysis
		if result.Complexity == "low" {
			result.RecommendedModel = "qwen-omni"
			result.Confidence = 0.85
			result.Reasoning += " - Adjusted for tester role (prefer fast, simple models)"
		}
	case types.RoleApprover:
		// Approvers need comprehensive analysis
		if result.Complexity == "high" {
			result.RecommendedModel = "qwen-omni"
			result.Confidence = 0.90
			result.Reasoning += " - Adjusted for approver role (prefer comprehensive models)"
		}
	}

	return result
}

// GetModelConfig returns the configuration for a specific model
func (ca *ContentAnalyzer) GetModelConfig(modelName string) (localmodels.ModelConfig, bool) {
	config, exists := ca.modelConfigs[modelName]
	return config, exists
}

// ListAvailableModels returns all available model configurations
func (ca *ContentAnalyzer) ListAvailableModels() map[string]localmodels.ModelConfig {
	return ca.modelConfigs
}
