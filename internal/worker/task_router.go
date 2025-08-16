package worker

import (
	"context"
	"fmt"
	"strings"

	"github.com/niko/mqtt-agent-orchestration/internal/ai"
	"github.com/niko/mqtt-agent-orchestration/internal/localmodels"
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// TaskComplexity represents the complexity level of a task
type TaskComplexity int

const (
	ComplexitySimple TaskComplexity = iota
	ComplexityMedium
	ComplexityHigh
)

// TaskRouter intelligently routes tasks between local models and external APIs
type TaskRouter struct {
	localModelManager *localmodels.Manager
	aiConfig          *ai.AIHelperConfig
	mcpEnabled        bool
}

// NewTaskRouter creates a new task router
func NewTaskRouter(localManager *localmodels.Manager, aiConfig *ai.AIHelperConfig) *TaskRouter {
	return &TaskRouter{
		localModelManager: localManager,
		aiConfig:          aiConfig,
		mcpEnabled:        true, // Enable MCP for local operations
	}
}

// RouteTask determines the best execution strategy for a task
func (tr *TaskRouter) RouteTask(ctx context.Context, task *types.WorkflowTask) (*TaskExecution, error) {
	complexity := tr.analyzeTaskComplexity(task)
	
	switch complexity {
	case ComplexitySimple:
		return tr.routeToLocalModel(ctx, task)
	case ComplexityMedium:
		// Try local first, fallback to API
		if execution, err := tr.routeToLocalModel(ctx, task); err == nil {
			return execution, nil
		}
		return tr.routeToExternalAPI(ctx, task, "medium")
	case ComplexityHigh:
		return tr.routeToExternalAPI(ctx, task, "high")
	default:
		return tr.routeToLocalModel(ctx, task)
	}
}

// analyzeTaskComplexity determines task complexity based on content and requirements
func (tr *TaskRouter) analyzeTaskComplexity(task *types.WorkflowTask) TaskComplexity {
	taskType := strings.ToLower(task.Type)
	content := strings.ToLower(fmt.Sprintf("%v", task.Payload))
	
	// High complexity indicators
	highComplexityKeywords := []string{
		"architecture", "design", "review", "security", "analysis", 
		"refactor", "optimization", "performance", "complex", "comprehensive",
		"strategy", "planning", "evaluation", "assessment",
	}
	
	// Medium complexity indicators  
	mediumComplexityKeywords := []string{
		"implement", "create", "generate", "modify", "update",
		"integration", "testing", "validation", "debugging",
	}
	
	// Simple task indicators (MCP/local model territory)
	simpleTaskKeywords := []string{
		"format", "lint", "syntax", "simple", "basic", "quick",
		"echo", "status", "info", "list", "search", "read", "write",
		"mcp", "tool", "file operation", "git operation",
	}
	
	// Check for high complexity
	for _, keyword := range highComplexityKeywords {
		if strings.Contains(content, keyword) || strings.Contains(taskType, keyword) {
			return ComplexityHigh
		}
	}
	
	// Check for simple tasks (MCP territory)
	for _, keyword := range simpleTaskKeywords {
		if strings.Contains(content, keyword) || strings.Contains(taskType, keyword) {
			return ComplexitySimple
		}
	}
	
	// Check for medium complexity
	for _, keyword := range mediumComplexityKeywords {
		if strings.Contains(content, keyword) || strings.Contains(taskType, keyword) {
			return ComplexityMedium
		}
	}
	
	// Default to medium complexity
	return ComplexityMedium
}

// routeToLocalModel routes task to local model with MCP capabilities
func (tr *TaskRouter) routeToLocalModel(ctx context.Context, task *types.WorkflowTask) (*TaskExecution, error) {
	if tr.localModelManager == nil {
		return nil, fmt.Errorf("local model manager not available")
	}
	
	// Select appropriate local model based on task type
	modelName := tr.selectLocalModel(task)
	
	execution := &TaskExecution{
		Strategy:  ExecutionStrategyLocal,
		ModelName: modelName,
		Task:      task,
		MCPEnabled: tr.mcpEnabled && tr.isMCPTask(task),
		Reasoning: fmt.Sprintf("Task complexity: simple/medium, using local model %s", modelName),
	}
	
	return execution, nil
}

// routeToExternalAPI routes task to external AI API
func (tr *TaskRouter) routeToExternalAPI(ctx context.Context, task *types.WorkflowTask, complexity string) (*TaskExecution, error) {
	if tr.aiConfig == nil {
		return nil, fmt.Errorf("AI configuration not available")
	}
	
	// Get preferred API based on complexity
	provider, apiConfig, err := tr.aiConfig.GetPreferredAPI(complexity)
	if err != nil {
		return nil, fmt.Errorf("no suitable API found: %w", err)
	}
	
	execution := &TaskExecution{
		Strategy:    ExecutionStrategyAPI,
		APIProvider: provider,
		APIConfig:   apiConfig,
		Task:        task,
		MCPEnabled:  false, // External APIs don't use MCP directly
		Reasoning:   fmt.Sprintf("Task complexity: %s, using %s API", complexity, provider),
	}
	
	return execution, nil
}

// selectLocalModel chooses the best local model for a task
func (tr *TaskRouter) selectLocalModel(task *types.WorkflowTask) string {
	taskType := strings.ToLower(task.Type)
	
	// Task-specific model selection
	switch {
	case strings.Contains(taskType, "embed") || strings.Contains(taskType, "search"):
		return "qwen-embedding-4b"
	case strings.Contains(taskType, "code") || strings.Contains(taskType, "develop"):
		return "qwen-omni-3b" // Good for coding tasks
	case strings.Contains(taskType, "visual") || strings.Contains(taskType, "image"):
		return "qwen-vl-7b" // Vision-language model
	default:
		return "qwen-omni-3b" // Default general purpose model
	}
}

// isMCPTask determines if a task should use MCP tools
func (tr *TaskRouter) isMCPTask(task *types.WorkflowTask) bool {
	mcpKeywords := []string{
		"file", "git", "search", "read", "write", "list",
		"directory", "repository", "database", "tool",
	}
	
	content := strings.ToLower(fmt.Sprintf("%s %v", task.Type, task.Payload))
	for _, keyword := range mcpKeywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	return false
}

// ExecutionStrategy defines how a task should be executed
type ExecutionStrategy int

const (
	ExecutionStrategyLocal ExecutionStrategy = iota
	ExecutionStrategyAPI
	ExecutionStrategyHybrid
)

// TaskExecution contains the execution plan for a task
type TaskExecution struct {
	Strategy    ExecutionStrategy
	ModelName   string // For local execution
	APIProvider string // For API execution
	APIConfig   ai.APIConfig
	Task        *types.WorkflowTask
	MCPEnabled  bool
	Reasoning   string
}

// Execute runs the task according to the execution plan
func (te *TaskExecution) Execute(ctx context.Context, localManager *localmodels.Manager, aiClient *ai.AIClient) (string, error) {
	switch te.Strategy {
	case ExecutionStrategyLocal:
		return te.executeLocal(ctx, localManager)
	case ExecutionStrategyAPI:
		return te.executeAPI(ctx, aiClient)
	default:
		return "", fmt.Errorf("unsupported execution strategy: %v", te.Strategy)
	}
}

// executeLocal executes task using local model
func (te *TaskExecution) executeLocal(ctx context.Context, localManager *localmodels.Manager) (string, error) {
	if localManager == nil {
		return "", fmt.Errorf("local model manager not available")
	}
	
	// Load model if needed
	if err := localManager.LoadModel(ctx, te.ModelName); err != nil {
		return "", fmt.Errorf("failed to load model %s: %w", te.ModelName, err)
	}
	
	// Get model instance
	model, err := localManager.GetModel(te.ModelName)
	if err != nil {
		return "", fmt.Errorf("failed to get model %s: %w", te.ModelName, err)
	}
	
	// Prepare input
	prompt := te.buildLocalPrompt()
	input := localmodels.ModelInput{
		Text:        prompt,
		Temperature: 0.7,
		MaxTokens:   te.getMaxTokensForTask(),
	}
	
	// Add MCP context if enabled
	if te.MCPEnabled {
		// MCP tools would be handled by the model implementation
		// For now, just add MCP context to the prompt
		input.Text = fmt.Sprintf("MCP Tools Available: %v\n\n%s", te.getRequiredMCPTools(), input.Text)
	}
	
	// Execute
	output, err := model.Predict(ctx, input)
	if err != nil {
		return "", fmt.Errorf("local model prediction failed: %w", err)
	}
	
	return output.Text, nil
}

// executeAPI executes task using external API
func (te *TaskExecution) executeAPI(ctx context.Context, aiClient *ai.AIClient) (string, error) {
	if aiClient == nil {
		return "", fmt.Errorf("AI client not available")
	}
	
	// For now, return a placeholder since we need to implement the full AI client
	// This would integrate with the AI helper scripts via MQTT
	return fmt.Sprintf("API execution not yet implemented for provider: %s, model: %s, prompt: %s", 
		te.APIProvider, te.APIConfig.Models[0], te.buildDetailedPrompt()), nil
}

// buildLocalPrompt creates a prompt optimized for local models
func (te *TaskExecution) buildLocalPrompt() string {
	var prompt strings.Builder
	
	// Keep it simple and direct for local models
	prompt.WriteString(fmt.Sprintf("Task: %s\n", te.Task.Type))
	
	if te.Task.PreviousOutput != "" {
		prompt.WriteString(fmt.Sprintf("Previous output: %s\n", te.Task.PreviousOutput))
	}
	
	for key, value := range te.Task.Payload {
		prompt.WriteString(fmt.Sprintf("%s: %v\n", key, value))
	}
	
	prompt.WriteString("\nPlease provide a clear, concise response.")
	
	return prompt.String()
}

// buildAPIMessages would create messages for API calls when implemented
// For now, we use buildDetailedPrompt() directly

// buildDetailedPrompt creates a comprehensive prompt for API calls
func (te *TaskExecution) buildDetailedPrompt() string {
	var prompt strings.Builder
	
	prompt.WriteString(fmt.Sprintf("Task Type: %s\n", te.Task.Type))
	prompt.WriteString(fmt.Sprintf("Required Role: %s\n\n", te.Task.RequiredRole))
	
	if te.Task.PreviousOutput != "" {
		prompt.WriteString(fmt.Sprintf("Previous Output to Review/Improve:\n%s\n\n", te.Task.PreviousOutput))
	}
	
	prompt.WriteString("Task Details:\n")
	for key, value := range te.Task.Payload {
		prompt.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
	}
	
	prompt.WriteString("\nPlease provide a comprehensive, high-quality response that demonstrates expertise in this domain.")
	
	return prompt.String()
}

// getMaxTokensForTask returns appropriate token limits based on task complexity
func (te *TaskExecution) getMaxTokensForTask() int {
	if te.MCPEnabled {
		return 1024 // Shorter responses for MCP tasks
	}
	
	switch te.Task.Type {
	case "echo", "status", "simple":
		return 512
	case "implement", "create", "generate":
		return 2048
	default:
		return 1024
	}
}

// getRequiredMCPTools returns MCP tools needed for the task
func (te *TaskExecution) getRequiredMCPTools() []string {
	var tools []string
	
	taskContent := strings.ToLower(fmt.Sprintf("%s %v", te.Task.Type, te.Task.Payload))
	
	if strings.Contains(taskContent, "file") || strings.Contains(taskContent, "read") || strings.Contains(taskContent, "write") {
		tools = append(tools, "filesystem")
	}
	
	if strings.Contains(taskContent, "git") || strings.Contains(taskContent, "repository") {
		tools = append(tools, "git")
	}
	
	if strings.Contains(taskContent, "search") || strings.Contains(taskContent, "query") {
		tools = append(tools, "qdrant_search")
	}
	
	return tools
}