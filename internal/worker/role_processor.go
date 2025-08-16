package worker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

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
	aiClient        *ai.AIClient
	taskRouter      *TaskRouter
}

// NewRoleBasedProcessor creates a processor for a specific role
func NewRoleBasedProcessor(role types.WorkerRole, ragService *rag.Service, modelManager *localmodels.Manager, contentAnalyzer *ContentAnalyzer, aiConfig *ai.AIHelperConfig) *RoleBasedProcessor {
	capabilities := GetCapabilitiesForRole(role)
	taskRouter := NewTaskRouter(modelManager, aiConfig)
	
	return &RoleBasedProcessor{
		role:            role,
		capabilities:    capabilities,
		ragService:      ragService,
		modelManager:    modelManager,
		contentAnalyzer: contentAnalyzer,
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
		ragContext, err := p.ragService.GetRelevantContext(ctx, workflowTask.Type,
			fmt.Sprintf("%s %s", workflowTask.Type, workflowTask.Payload["document_type"]))
		if err == nil {
			// Add RAG context to task payload for execution
			if workflowTask.Payload == nil {
				workflowTask.Payload = make(map[string]string)
			}
			workflowTask.Payload["rag_context"] = ragContext
		}
	}

	// Execute using the determined strategy
	result, err := execution.Execute(ctx, p.modelManager, p.aiClient)
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

// processDeveloperTask handles initial content creation
func (p *RoleBasedProcessor) processDeveloperTask(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	switch taskContext.Task.Type {
	case "create_document":
		return p.createDocument(ctx, taskContext)
	default:
		return "", fmt.Errorf("developer role doesn't support task type: %s", taskContext.Task.Type)
	}
}

// processReviewerTask handles content review and improvement
func (p *RoleBasedProcessor) processReviewerTask(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	if taskContext.Task.PreviousOutput == "" {
		return "", fmt.Errorf("reviewer task requires previous output")
	}

	switch taskContext.Task.Type {
	case "create_document":
		return p.reviewDocument(ctx, taskContext)
	default:
		return "", fmt.Errorf("reviewer role doesn't support task type: %s", taskContext.Task.Type)
	}
}

// processApproverTask handles final approval
func (p *RoleBasedProcessor) processApproverTask(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	if taskContext.Task.PreviousOutput == "" {
		return "", fmt.Errorf("approver task requires previous output")
	}

	switch taskContext.Task.Type {
	case "create_document":
		return p.approveDocument(ctx, taskContext)
	default:
		return "", fmt.Errorf("approver role doesn't support task type: %s", taskContext.Task.Type)
	}
}

// processTesterTask handles testing and validation
func (p *RoleBasedProcessor) processTesterTask(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	switch taskContext.Task.Type {
	case "create_document":
		return p.testDocument(ctx, taskContext)
	default:
		return "", fmt.Errorf("tester role doesn't support task type: %s", taskContext.Task.Type)
	}
}

// createDocument creates initial document content with optimized context
func (p *RoleBasedProcessor) createDocument(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	documentType := taskContext.Task.Payload["document_type"]

	// Try local model first if available
	if p.modelManager != nil && taskContext.ModelAnalysis != nil {
		result, err := p.processWithLocalModel(ctx, taskContext, taskContext.ModelAnalysis.RecommendedModel)
		if err == nil {
			return result, nil
		}
		// Fall back to external AI helper if local model fails
	}

	// Build optimized prompt using system prompt and RAG context
	optimizedPrompt := p.buildOptimizedPrompt(taskContext, "create", documentType)

	// Use most appropriate AI helper for initial creation
	aiHelper := p.selectAIHelper("development")
	cmd := exec.CommandContext(ctx, aiHelper, optimizedPrompt)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("AI helper %s failed: %w", aiHelper, err)
	}

	return string(output), nil
}

// reviewDocument reviews and improves existing content with optimized context
func (p *RoleBasedProcessor) reviewDocument(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	// Build optimized prompt for review phase
	optimizedPrompt := p.buildOptimizedPrompt(taskContext, "review", taskContext.Task.Payload["document_type"])

	// Use AI helper focused on review/analysis
	aiHelper := p.selectAIHelper("review")
	cmd := exec.CommandContext(ctx, aiHelper, optimizedPrompt)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("review AI helper %s failed: %w", aiHelper, err)
	}

	return string(output), nil
}

// approveDocument performs final approval check with optimized context
func (p *RoleBasedProcessor) approveDocument(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	// Build optimized prompt for approval phase
	optimizedPrompt := p.buildOptimizedPrompt(taskContext, "approve", taskContext.Task.Payload["document_type"])

	// Use AI helper best for final analysis
	aiHelper := p.selectAIHelper("approval")
	cmd := exec.CommandContext(ctx, aiHelper, optimizedPrompt)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("approval AI helper %s failed: %w", aiHelper, err)
	}

	return string(output), nil
}

// testDocument validates the document
func (p *RoleBasedProcessor) testDocument(ctx context.Context, taskContext *EnhancedTaskContext) (string, error) {
	content := taskContext.Task.PreviousOutput

	// For coding standards, test by checking examples compile/lint
	if taskContext.Task.Payload["document_type"] == "go_coding_standards" {
		return p.testGoCodingStandards(ctx, content)
	}

	return "Document testing not implemented for this type", nil
}

// buildOptimizedPrompt creates token-efficient prompts using system prompt and RAG context
func (p *RoleBasedProcessor) buildOptimizedPrompt(taskContext *EnhancedTaskContext, phase, documentType string) string {
	var prompt strings.Builder

	// Start with system prompt for role context (reduces token usage by providing clear role definition)
	if taskContext.SystemPrompt != "" {
		prompt.WriteString(taskContext.SystemPrompt)
		prompt.WriteString("\n\n")
	}

	// Add relevant RAG context (reduces token usage by providing specific domain knowledge)
	if taskContext.RAGContext != "" {
		prompt.WriteString("Relevant Context:\n")
		prompt.WriteString(taskContext.RAGContext)
		prompt.WriteString("\n\n")
	}

	// Add phase-specific instructions (concise and targeted)
	switch phase {
	case "create":
		prompt.WriteString(fmt.Sprintf("Create a comprehensive %s document.", documentType))

	case "review":
		prompt.WriteString(fmt.Sprintf("Review and improve this %s document.\n\nPrevious version:\n%s",
			documentType, taskContext.Task.PreviousOutput))

	case "approve":
		prompt.WriteString(fmt.Sprintf("Perform final approval for this %s document.\n\nContent to approve:\n%s\n\nRespond with APPROVED: [reason] or REJECTED: [issues]",
			documentType, taskContext.Task.PreviousOutput))
	}

	return prompt.String()
}

// processWithLocalModel processes a task using a local model
func (p *RoleBasedProcessor) processWithLocalModel(ctx context.Context, taskContext *EnhancedTaskContext, modelName string) (string, error) {
	if p.modelManager == nil {
		return "", fmt.Errorf("model manager not available")
	}

	// Load the model if not already loaded
	if err := p.modelManager.LoadModel(ctx, modelName); err != nil {
		return "", fmt.Errorf("failed to load model %s: %w", modelName, err)
	}

	// Get model instance
	model, err := p.modelManager.GetModel(modelName)
	if err != nil {
		return "", fmt.Errorf("failed to get model %s: %w", modelName, err)
	}

	// Build optimized prompt
	documentType := taskContext.Task.Payload["document_type"]
	optimizedPrompt := p.buildOptimizedPrompt(taskContext, "create", documentType)

	// Create model input
	input := localmodels.ModelInput{
		Text:        optimizedPrompt,
		Temperature: 0.7,
		MaxTokens:   2048,
	}

	// Process with local model
	output, err := model.Predict(ctx, input)
	if err != nil {
		return "", fmt.Errorf("local model prediction failed: %w", err)
	}

	return output.Text, nil
}

// Helper methods

func (p *RoleBasedProcessor) buildGoCodingStandardsPrompt(ragContext string) string {
	return fmt.Sprintf(`Create comprehensive Go coding standards document. Include:

1. Core Principles (explicit over implicit, composition over inheritance)
2. Package Management (naming, organization, imports)
3. Variable/Constant Declaration (naming, scoping, zero values)
4. Function/Method Design (naming, parameters, returns, receivers)
5. Error Handling (explicit handling, wrapping, checking)
6. Struct/Interface Design (composition, embedding, small interfaces)
7. Concurrency Patterns (goroutines, channels, context)
8. Testing Standards (table-driven, mocking, benchmarks)
9. Code Organization (directory structure, files)
10. Performance Guidelines (allocation, strings, profiling)
11. Documentation (godoc, comments, examples)
12. Security (validation, secrets)
13. Compliance Checklist

Use this context: %s

Format as markdown with clear good/bad examples. Make it comprehensive but practical.`, ragContext)
}

func (p *RoleBasedProcessor) selectAIHelper(phase string) string {
	switch phase {
	case "development":
		return "gemini_code_analyzer" // Best for comprehensive content
	case "review":
		return "cerebras_code_analyzer" // Fast and good for improvements
	case "approval":
		return "groq_fast_analyzer" // Quick final checks
	default:
		return "gemini_code_analyzer"
	}
}

func (p *RoleBasedProcessor) testGoCodingStandards(ctx context.Context, content string) (string, error) {
	// Extract Go code examples and test them
	// For now, just validate the document structure
	requiredSections := []string{
		"Core Principles",
		"Error Handling",
		"Testing Standards",
		"Compliance Checklist",
	}

	var missingSection []string
	for _, section := range requiredSections {
		if !strings.Contains(content, section) {
			missingSection = append(missingSection, section)
		}
	}

	if len(missingSection) > 0 {
		return fmt.Sprintf("FAILED: Missing required sections: %s", strings.Join(missingSection, ", ")), nil
	}

	return "PASSED: Document structure validates successfully", nil
}

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
