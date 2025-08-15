package worker

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/niko/mqtt-agent-orchestration/internal/rag"
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// RoleBasedProcessor implements role-specific task processing
type RoleBasedProcessor struct {
	role        types.WorkerRole
	capabilities types.WorkerCapabilities
	ragService  *rag.Service
}

// NewRoleBasedProcessor creates a processor for a specific role
func NewRoleBasedProcessor(role types.WorkerRole, ragService *rag.Service) *RoleBasedProcessor {
	capabilities := GetCapabilitiesForRole(role)
	return &RoleBasedProcessor{
		role:         role,
		capabilities: capabilities,
		ragService:   ragService,
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
	
	// Get RAG context if available
	var ragContext string
	if p.ragService != nil && p.capabilities.RAGEnabled {
		context, err := p.ragService.GetRelevantContext(ctx, workflowTask.Type, 
			fmt.Sprintf("%s %s", workflowTask.Type, workflowTask.Payload["document_type"]))
		if err == nil {
			ragContext = context
		}
	}
	
	// Process based on role and task type
	switch p.role {
	case types.RoleDeveloper:
		return p.processDeveloperTask(ctx, workflowTask, ragContext)
	case types.RoleReviewer:
		return p.processReviewerTask(ctx, workflowTask, ragContext)
	case types.RoleApprover:
		return p.processApproverTask(ctx, workflowTask, ragContext)
	case types.RoleTester:
		return p.processTesterTask(ctx, workflowTask, ragContext)
	default:
		return "", fmt.Errorf("unsupported role: %s", p.role)
	}
}

// processDeveloperTask handles initial content creation
func (p *RoleBasedProcessor) processDeveloperTask(ctx context.Context, task *types.WorkflowTask, ragContext string) (string, error) {
	switch task.Type {
	case "create_document":
		return p.createDocument(ctx, task, ragContext)
	default:
		return "", fmt.Errorf("developer role doesn't support task type: %s", task.Type)
	}
}

// processReviewerTask handles content review and improvement
func (p *RoleBasedProcessor) processReviewerTask(ctx context.Context, task *types.WorkflowTask, ragContext string) (string, error) {
	if task.PreviousOutput == "" {
		return "", fmt.Errorf("reviewer task requires previous output")
	}
	
	switch task.Type {
	case "create_document":
		return p.reviewDocument(ctx, task, ragContext)
	default:
		return "", fmt.Errorf("reviewer role doesn't support task type: %s", task.Type)
	}
}

// processApproverTask handles final approval
func (p *RoleBasedProcessor) processApproverTask(ctx context.Context, task *types.WorkflowTask, ragContext string) (string, error) {
	if task.PreviousOutput == "" {
		return "", fmt.Errorf("approver task requires previous output")
	}
	
	switch task.Type {
	case "create_document":
		return p.approveDocument(ctx, task, ragContext)
	default:
		return "", fmt.Errorf("approver role doesn't support task type: %s", task.Type)
	}
}

// processTesterTask handles testing and validation
func (p *RoleBasedProcessor) processTesterTask(ctx context.Context, task *types.WorkflowTask, ragContext string) (string, error) {
	switch task.Type {
	case "create_document":
		return p.testDocument(ctx, task, ragContext)
	default:
		return "", fmt.Errorf("tester role doesn't support task type: %s", task.Type)
	}
}

// createDocument creates initial document content
func (p *RoleBasedProcessor) createDocument(ctx context.Context, task *types.WorkflowTask, ragContext string) (string, error) {
	documentType := task.Payload["document_type"]
	
	var prompt string
	switch documentType {
	case "go_coding_standards":
		prompt = p.buildGoCodingStandardsPrompt(ragContext)
	default:
		return "", fmt.Errorf("unknown document type: %s", documentType)
	}
	
	// Use most appropriate AI helper for initial creation
	aiHelper := p.selectAIHelper("development")
	cmd := exec.CommandContext(ctx, aiHelper, prompt)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("AI helper %s failed: %w", aiHelper, err)
	}
	
	return string(output), nil
}

// reviewDocument reviews and improves existing content
func (p *RoleBasedProcessor) reviewDocument(ctx context.Context, task *types.WorkflowTask, ragContext string) (string, error) {
	previousContent := task.PreviousOutput
	documentType := task.Payload["document_type"]
	
	prompt := fmt.Sprintf(`Review and improve this %s document. Focus on:
1. Completeness - are all important topics covered?
2. Accuracy - are the guidelines correct and up-to-date?
3. Clarity - is the content clear and well-organized?
4. Examples - are there good/bad code examples?
5. Consistency - is the style and format consistent?

Use this context for reference:
%s

Previous version to review:
%s

Provide the improved version with clear explanations of changes made.`, documentType, ragContext, previousContent)
	
	// Use AI helper focused on review/analysis
	aiHelper := p.selectAIHelper("review")
	cmd := exec.CommandContext(ctx, aiHelper, prompt)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("review AI helper %s failed: %w", aiHelper, err)
	}
	
	return string(output), nil
}

// approveDocument performs final approval check
func (p *RoleBasedProcessor) approveDocument(ctx context.Context, task *types.WorkflowTask, ragContext string) (string, error) {
	content := task.PreviousOutput
	documentType := task.Payload["document_type"]
	
	prompt := fmt.Sprintf(`Perform final approval check for this %s document. 
Check for:
1. Technical accuracy
2. Completeness 
3. Production readiness
4. Adherence to standards
5. No factual errors

Context:
%s

Document to approve:
%s

Respond with either:
APPROVED: [brief explanation]
REJECTED: [specific issues that need fixing]`, documentType, ragContext, content)
	
	// Use AI helper best for final analysis
	aiHelper := p.selectAIHelper("approval")
	cmd := exec.CommandContext(ctx, aiHelper, prompt)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("approval AI helper %s failed: %w", aiHelper, err)
	}
	
	return string(output), nil
}

// testDocument validates the document
func (p *RoleBasedProcessor) testDocument(ctx context.Context, task *types.WorkflowTask, ragContext string) (string, error) {
	content := task.PreviousOutput
	
	// For coding standards, test by checking examples compile/lint
	if task.Payload["document_type"] == "go_coding_standards" {
		return p.testGoCodingStandards(ctx, content)
	}
	
	return "Document testing not implemented for this type", nil
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