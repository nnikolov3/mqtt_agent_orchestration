package worker

import (
	"context"
	"fmt"

	"github.com/niko/mqtt-agent-orchestration/internal/ai"
	"github.com/niko/mqtt-agent-orchestration/internal/localmodels"
	"github.com/niko/mqtt-agent-orchestration/internal/rag"
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// RoleBasedProcessor implements role-specific task processing
type RoleBasedProcessor struct {
	role            types.WorkerRole
	capabilities    types.WorkerCapabilities
	ragService      *rag.Service
	modelManager    *localmodels.Manager
	contentAnalyzer *ContentAnalyzer
	helperManager   *ai.HelperManager
	taskRouter      *TaskRouter
}

// NewRoleBasedProcessor creates a processor for a specific role
func NewRoleBasedProcessor(role types.WorkerRole, ragService *rag.Service, modelManager *localmodels.Manager, contentAnalyzer *ContentAnalyzer, aiConfig *ai.AIHelperConfig) *RoleBasedProcessor {
	capabilities := GetCapabilitiesForRole(role)
	taskRouter := NewTaskRouter(modelManager, aiConfig)
	helperManager := ai.NewHelperManager()

	return &RoleBasedProcessor{
		role:            role,
		capabilities:    capabilities,
		ragService:      ragService,
		modelManager:    modelManager,
		contentAnalyzer: contentAnalyzer,
		helperManager:   helperManager,
		taskRouter:      taskRouter,
	}
}

// ProcessTask processes tasks according to the worker's role
func (p *RoleBasedProcessor) ProcessTask(ctx context.Context, task types.Task) (string, error) {
	// For now, this will be called with regular tasks and we'll extend them
	// In a real implementation, we'd use a proper interface or type switch
	return p.processSimpleTask(ctx, task)
}

// processSimpleTask handles basic task processing for backward compatibility
func (p *RoleBasedProcessor) processSimpleTask(ctx context.Context, task types.Task) (string, error) {
	switch task.Type {
	case "echo":
		if message, ok := task.Payload["message"]; ok {
			return fmt.Sprintf("Echo from %s: %s", p.role, message), nil
		}
		return fmt.Sprintf("Echo from %s: (no message)", p.role), nil
	default:
		return "", fmt.Errorf("simple task type %s not supported by role %s", task.Type, p.role)
	}
}

// ProcessWorkflowTask processes workflow tasks according to the worker's role
func (p *RoleBasedProcessor) ProcessWorkflowTask(ctx context.Context, workflowTask *types.WorkflowTask) (string, error) {
	// Verify role match
	if workflowTask.RequiredRole != p.role {
		return "", fmt.Errorf("task requires role %s, but worker is %s", workflowTask.RequiredRole, p.role)
	}

	// Use task router to determine optimal execution strategy
	execution, err := p.taskRouter.RouteTask(ctx, workflowTask)
	if err != nil {
		return "", fmt.Errorf("task routing failed: %w", err)
	}

	// Log routing decision for monitoring
	fmt.Printf("Task routed: %s (Strategy: %v, Model: %s, API: %s)\n",
		execution.Reasoning, execution.Strategy, execution.ModelName, execution.APIProvider)

	// Get RAG context if enabled
	if p.ragService != nil && p.capabilities.RAGEnabled {
		var docTypeStr string
		if v, ok := workflowTask.Payload["document_type"]; ok {
			if s, ok := v.(string); ok {
				docTypeStr = s
			} else {
				docTypeStr = fmt.Sprint(v)
			}
		}

		ragContext, err := p.ragService.GetRelevantContext(ctx, workflowTask.Type,
			fmt.Sprintf("%s %s", workflowTask.Type, docTypeStr))
		if err == nil {
			// Add RAG context to task payload for execution
			if workflowTask.Payload == nil {
				workflowTask.Payload = make(map[string]interface{})
			}
			workflowTask.Payload["rag_context"] = ragContext
		}
	}

	// Execute using the determined strategy
	result, err := execution.Execute(ctx, p.modelManager, p.helperManager)
	if err != nil {
		return "", fmt.Errorf("task execution failed: %w", err)
	}

	return result, nil
}

// EnhancedTaskContext provides optimized context for AI API calls
type EnhancedTaskContext struct {
	SystemPrompt  string
	RAGContext    string
	Task          *types.WorkflowTask
	ModelAnalysis *AnalysisResult
}

/*
// Role-specific processing methods - preserved for future role-specific implementations

// processDeveloperTask handles initial content creation
func (p *RoleBasedProcessor) processDeveloperTask(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	// Initial development task - create the document
	if taskContext.Task.Type == "development" || taskContext.Task.Type == "creation" {
		return p.createDocument(ctx, taskContext)
	}
	return "", fmt.Errorf("unsupported task type for developer: %s", taskContext.Task.Type)
}

// processReviewerTask handles content review and improvement
func (p *RoleBasedProcessor) processReviewerTask(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	if taskContext.Task.Type == "review" {
		content, exists := taskContext.Task.Payload["content"].(string)
		if !exists {
			return "", fmt.Errorf("no content to review in task payload")
		}
		taskContext.Task.Payload["existing_content"] = content
		return p.reviewDocument(ctx, taskContext)
	}
	return "", fmt.Errorf("unsupported task type for reviewer: %s", taskContext.Task.Type)
}

// processApproverTask handles final approval
func (p *RoleBasedProcessor) processApproverTask(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	if taskContext.Task.Type == "approval" {
		content, exists := taskContext.Task.Payload["content"].(string)
		if !exists {
			return "", fmt.Errorf("no content to approve in task payload")
		}
		taskContext.Task.Payload["existing_content"] = content
		return p.approveDocument(ctx, taskContext)
	}
	return "", fmt.Errorf("unsupported task type for approver: %s", taskContext.Task.Type)
}

// processTesterTask handles testing and validation
func (p *RoleBasedProcessor) processTesterTask(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	// Test tasks - validate the document
	if taskContext.Task.Type == "testing" || taskContext.Task.Type == "validation" {
		return p.testDocument(ctx, taskContext)
	}
	return "", fmt.Errorf("unsupported task type for tester: %s", taskContext.Task.Type)
}

// createDocument creates initial document content with optimized context
func (p *RoleBasedProcessor) createDocument(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	// Get document type from task
	documentType := "general"
	if docType, ok := taskContext.Task.Payload["document_type"].(string); ok {
		documentType = docType
	}

	// Check if we should use local model based on analysis
	if taskContext.ModelAnalysis != nil && taskContext.ModelAnalysis.UseLocalModel {
		fmt.Printf("Using local model %s for document creation\n", taskContext.ModelAnalysis.RecommendedModel)
		result, err := p.processWithLocalModel(ctx, taskContext, taskContext.ModelAnalysis.RecommendedModel)
		if err == nil {
			return result, nil
		}
		// Fall back to cloud API if local model fails
		fmt.Printf("Local model failed, falling back to cloud API: %v\n", err)
	}

	optimizedPrompt := p.buildOptimizedPrompt(taskContext, "create", documentType)

	// Use AI helper to create document
	aiHelper := p.selectAIHelper("development")
	response, err := p.helperManager.ProcessWithHelper(ctx, aiHelper, optimizedPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to create document: %w", err)
	}

	return response, nil
}

// reviewDocument reviews and improves existing content with optimized context
func (p *RoleBasedProcessor) reviewDocument(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	// Get existing content
	existingContent, exists := taskContext.Task.Payload["existing_content"].(string)
	if !exists {
		return "", fmt.Errorf("no existing content to review")
	}

	docTypeReview := fmt.Sprintf("review of: %s", existingContent[:min(100, len(existingContent))])
	optimizedPrompt := p.buildOptimizedPrompt(taskContext, "review", docTypeReview)

	// Use AI helper to review
	aiHelper := p.selectAIHelper("review")
	response, err := p.helperManager.ProcessWithHelper(ctx, aiHelper, optimizedPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to review document: %w", err)
	}

	return response, nil
}

// approveDocument performs final approval check with optimized context
func (p *RoleBasedProcessor) approveDocument(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	// Get content to approve
	contentToApprove, exists := taskContext.Task.Payload["existing_content"].(string)
	if !exists {
		return "", fmt.Errorf("no content to approve")
	}

	docTypeApprove := fmt.Sprintf("approval of: %s", contentToApprove[:min(100, len(contentToApprove))])
	optimizedPrompt := p.buildOptimizedPrompt(taskContext, "approve", docTypeApprove)

	// Use AI helper for approval
	aiHelper := p.selectAIHelper("approval")
	response, err := p.helperManager.ProcessWithHelper(ctx, aiHelper, optimizedPrompt)
	if err != nil {
		return "", fmt.Errorf("failed to approve document: %w", err)
	}

	return response, nil
}

// testDocument validates the document
func (p *RoleBasedProcessor) testDocument(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	// Get content to test
	content, exists := taskContext.Task.Payload["content"].(string)
	if exists && strings.Contains(strings.ToLower(content), "go") {
		// Special handling for Go code testing
		return p.testGoCodingStandards(ctx, content)
	}
	return "Test validation completed - no specific tests required", nil
}

// buildOptimizedPrompt creates token-efficient prompts using system prompt and RAG context
func (p *RoleBasedProcessor) buildOptimizedPrompt(taskContext *EnhancedTaskContext, phase, documentType string) string {
	var promptBuilder strings.Builder

	// Add phase-specific instruction
	switch phase {
	case "create":
		promptBuilder.WriteString("Create a comprehensive document for the following request:\n\n")
	case "review":
		promptBuilder.WriteString("Review and improve the following content:\n\n")
	case "approve":
		promptBuilder.WriteString("Evaluate the following content for final approval:\n\n")
	default:
		promptBuilder.WriteString("Process the following task:\n\n")
	}

	// Add task context
	promptBuilder.WriteString(fmt.Sprintf("Task Type: %s\n", taskContext.Task.Type))
	promptBuilder.WriteString(fmt.Sprintf("Document Type: %s\n", documentType))

	// Add RAG context if available (limited to save tokens)
	if taskContext.RAGContext != "" {
		promptBuilder.WriteString("\nRelevant Context:\n")
		// Limit RAG context to 500 characters for token efficiency
		if len(taskContext.RAGContext) > 500 {
			promptBuilder.WriteString(taskContext.RAGContext[:500] + "...")
		} else {
			promptBuilder.WriteString(taskContext.RAGContext)
		}
		promptBuilder.WriteString("\n\n")
	}

	// Add task-specific payload information
	if desc, ok := taskContext.Task.Payload["description"].(string); ok {
		promptBuilder.WriteString(fmt.Sprintf("Description: %s\n", desc))
	}
	if content, ok := taskContext.Task.Payload["existing_content"].(string); ok {
		promptBuilder.WriteString(fmt.Sprintf("\nExisting Content:\n%s\n", content))
	}

	return promptBuilder.String()
}

// processWithLocalModel processes a task using a local model
func (p *RoleBasedProcessor) processWithLocalModel(ctx context.Context, taskContext *EnhancedTaskContext, modelName string) (string, error) {
	if p.modelManager == nil {
		return "", fmt.Errorf("local model manager not available")
	}

	// Check if model is loaded
	if !p.modelManager.IsModelLoaded(modelName) {
		// Try to load the model
		if err := p.modelManager.LoadModel(ctx, modelName); err != nil {
			return "", fmt.Errorf("failed to load model %s: %w", modelName, err)
		}
	}

	// Get document type for context
	documentType := "general"
	if docType, ok := taskContext.Task.Payload["document_type"].(string); ok {
		documentType = docType
	}

	// Build optimized prompt for local model
	// Local models often need more explicit instructions
	optimizedPrompt := p.buildOptimizedPrompt(taskContext, "create", documentType)

	// Process with local model
	response, err := p.modelManager.GenerateWithModel(ctx, modelName, optimizedPrompt)
	if err != nil {
		return "", fmt.Errorf("local model generation failed: %w", err)
	}

	return response, nil
}

func (p *RoleBasedProcessor) buildGoCodingStandardsPrompt(ragContext string) string {
	var promptBuilder strings.Builder
	promptBuilder.WriteString("Review the following Go code against coding standards:\n\n")

	if ragContext != "" {
		promptBuilder.WriteString("Relevant Standards:\n")
		promptBuilder.WriteString(ragContext)
		promptBuilder.WriteString("\n\n")
	}

	promptBuilder.WriteString("Please check for:\n")
	promptBuilder.WriteString("- Code formatting and style\n")
	promptBuilder.WriteString("- Error handling patterns\n")
	promptBuilder.WriteString("- Concurrency safety\n")
	promptBuilder.WriteString("- Performance considerations\n")
	promptBuilder.WriteString("- Security issues\n")
	promptBuilder.WriteString("- Best practices adherence\n")

	return promptBuilder.String()
}

func (p *RoleBasedProcessor) selectAIHelper(phase string) string {
	// Select appropriate AI helper based on phase
	// This could be made more sophisticated with configuration
	switch phase {
	case "development":
		return "cerebras_code_generator"
	case "review":
		return "gemini_code_analyzer"
	case "approval":
		return "openai_validator"
	default:
		return "claude_haiku_assistant"
	}
}

func (p *RoleBasedProcessor) testGoCodingStandards(ctx context.Context, content string) (string, error) {
	// Get Go coding standards from RAG if available
	var ragContext string
	if p.ragService != nil {
		context, _ := p.ragService.GetRelevantContext(ctx, "testing", "go coding standards")
		ragContext = context
	}

	prompt := p.buildGoCodingStandardsPrompt(ragContext)
	prompt += "\n\nCode to review:\n" + content

	// Use a code analysis helper
	response, err := p.helperManager.ProcessWithHelper(ctx, "gemini_code_analyzer", prompt)
	if err != nil {
		return "", fmt.Errorf("failed to test Go coding standards: %w", err)
	}

	return response, nil
}
*/

// min returns the minimum of two integers

// GetCapabilitiesForRole returns capabilities for each role (exported)
func GetCapabilitiesForRole(role types.WorkerRole) types.WorkerCapabilities {
	switch role {
	case types.RoleDeveloper:
		return types.WorkerCapabilities{
			Roles:          []types.WorkerRole{types.RoleDeveloper},
			AIHelpers:      []string{"gemini_code_analyzer", "cerebras_code_analyzer"},
			Languages:      []string{"go", "python", "bash"},
			Specialization: "content_creation",
			RAGEnabled:     true,
		}
	case types.RoleReviewer:
		return types.WorkerCapabilities{
			Roles:          []types.WorkerRole{types.RoleReviewer},
			AIHelpers:      []string{"cerebras_code_analyzer", "groq_fast_analyzer"},
			Languages:      []string{"go", "python", "bash"},
			Specialization: "content_review",
			RAGEnabled:     true,
		}
	case types.RoleApprover:
		return types.WorkerCapabilities{
			Roles:          []types.WorkerRole{types.RoleApprover},
			AIHelpers:      []string{"groq_fast_analyzer", "gemini_code_analyzer"},
			Languages:      []string{"go", "python", "bash"},
			Specialization: "final_approval",
			RAGEnabled:     true,
		}
	case types.RoleTester:
		return types.WorkerCapabilities{
			Roles:          []types.WorkerRole{types.RoleTester},
			AIHelpers:      []string{"cerebras_code_analyzer"},
			Languages:      []string{"go", "python", "bash"},
			Specialization: "validation",
			RAGEnabled:     false,
		}
	default:
		return types.WorkerCapabilities{}
	}
}
