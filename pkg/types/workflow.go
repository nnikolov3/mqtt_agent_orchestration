package types

// WorkerRole defines the role of a worker in the pipeline
type WorkerRole string

const (
	RoleDeveloper    WorkerRole = "developer"
	RoleReviewer     WorkerRole = "reviewer"
	RoleApprover     WorkerRole = "approver"
	RoleTester       WorkerRole = "tester"
	RoleOrchestrator WorkerRole = "orchestrator"
)

// WorkflowStage represents a stage in the autonomous pipeline
type WorkflowStage string

const (
	StageDevelopment WorkflowStage = "development"
	StageReview      WorkflowStage = "review"
	StageApproval    WorkflowStage = "approval"
	StageTesting     WorkflowStage = "testing"
	StageCompleted   WorkflowStage = "completed"
	StageFailed      WorkflowStage = "failed"
)

// WorkflowTask extends Task with role-based information
type WorkflowTask struct {
	Task
	WorkflowID     string        `json:"workflow_id"`
	Stage          WorkflowStage `json:"stage"`
	RequiredRole   WorkerRole    `json:"required_role"`
	PreviousOutput string        `json:"previous_output,omitempty"`
	ReviewFeedback string        `json:"review_feedback,omitempty"`
	RAGContext     string        `json:"rag_context,omitempty"`
	RetryCount     int           `json:"retry_count"`
	MaxRetries     int           `json:"max_retries"`
}

// WorkflowResult extends TaskResult with workflow information
type WorkflowResult struct {
	TaskResult
	WorkflowID     string        `json:"workflow_id"`
	Stage          WorkflowStage `json:"stage"`
	WorkerRole     WorkerRole    `json:"worker_role"`
	NextStage      WorkflowStage `json:"next_stage,omitempty"`
	ReviewFeedback string        `json:"review_feedback,omitempty"`
	Approved       bool          `json:"approved"`
	RequiresRetry  bool          `json:"requires_retry"`
}

// WorkerCapabilities defines what a worker can do
type WorkerCapabilities struct {
	Roles          []WorkerRole `json:"roles"`
	AIHelpers      []string     `json:"ai_helpers"`
	Languages      []string     `json:"languages"`
	Specialization string       `json:"specialization"`
	RAGEnabled     bool         `json:"rag_enabled"`
}

// ExtendedWorkerStatus extends WorkerStatus with role information
type ExtendedWorkerStatus struct {
	WorkerStatus
	Role          WorkerRole         `json:"role"`
	Capabilities  WorkerCapabilities `json:"capabilities"`
	AssignedStage WorkflowStage      `json:"assigned_stage,omitempty"`
	WorkflowID    string             `json:"workflow_id,omitempty"`
}

// RAGQuery represents a query to the knowledge base
type RAGQuery struct {
	Query      string   `json:"query"`
	Collection string   `json:"collection"`
	TopK       int      `json:"top_k"`
	Threshold  float64  `json:"threshold"`
	Filters    []string `json:"filters,omitempty"`
}

// RAGResponse represents response from knowledge base
type RAGResponse struct {
	Documents []RAGDocument `json:"documents"`
	Query     string        `json:"query"`
	TotalHits int           `json:"total_hits"`
}

// RAGDocument represents a document from the knowledge base
type RAGDocument struct {
	Content  string            `json:"content"`
	Score    float64           `json:"score"`
	Metadata map[string]string `json:"metadata"`
	Source   string            `json:"source"`
}
