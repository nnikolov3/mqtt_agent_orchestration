package types

import "time"

// SubmitWorkflowRequest represents the request to submit a new workflow
type SubmitWorkflowRequest struct {
	Content    string                 `json:"content"`
	Complexity string                 `json:"complexity"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// Workflow represents a multi-stage workflow
type Workflow struct {
	ID           string                 `json:"id"`
	Content      string                 `json:"content"`
	Complexity   string                 `json:"complexity"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
	Tasks        []*WorkflowTask        `json:"tasks"`
	CurrentStage WorkflowStage          `json:"current_stage"`
}

// Task represents a single task for a worker
type Task struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Payload   map[string]string `json:"payload"`
	CreatedAt time.Time         `json:"created_at"`
	Priority  int               `json:"priority"`
}

// TaskResult represents the result of a task
type TaskResult struct {
	TaskID      string    `json:"task_id"`
	WorkerID    string    `json:"worker_id"`
	Success     bool      `json:"success"`
	Result      string    `json:"result,omitempty"`
	Error       string    `json:"error,omitempty"`
	ProcessedAt time.Time `json:"processed_at"`
	Duration    int64     `json:"duration"`
}

// WorkerStatus represents the status of a worker
type WorkerStatus struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	LastSeen    time.Time `json:"last_seen"`
	TasksTotal  int       `json:"tasks_total"`
	TasksError  int       `json:"tasks_error"`
	CurrentTask string    `json:"current_task"`
}

// WorkflowTask represents a task within a workflow
type WorkflowTask struct {
	ID             string                 `json:"id"`
	WorkflowID     string                 `json:"workflow_id"`
	Stage          WorkflowStage          `json:"stage"`
	RequiredRole   WorkerRole             `json:"required_role"`
	Payload        map[string]interface{} `json:"payload"`
	PreviousOutput string                 `json:"previous_output"`
	Type           string                 `json:"type"`
}

// WorkflowResult represents the result of a workflow task
type WorkflowResult struct {
	TaskResult
	WorkflowID     string        `json:"workflow_id"`
	Stage          WorkflowStage `json:"stage"`
	WorkerRole     WorkerRole    `json:"worker_role"`
	Approved       bool          `json:"approved"`
	RequiresRetry  bool          `json:"requires_retry"`
	ReviewFeedback string        `json:"review_feedback"`
}

// ExtendedWorkerStatus represents the extended status of a worker
type ExtendedWorkerStatus struct {
	WorkerStatus
	Role         WorkerRole `json:"role"`
	Capabilities []string   `json:"capabilities"`
}

// WorkerCapabilities defines what a worker can do
type WorkerCapabilities struct {
	Roles          []WorkerRole `json:"roles"`
	AIHelpers      []string     `json:"ai_helpers"`
	Languages      []string     `json:"languages"`
	Specialization string       `json:"specialization"`
	RAGEnabled     bool         `json:"rag_enabled"`
}

// EmbeddingRequest represents a request to generate an embedding
type EmbeddingRequest struct {
	Text      string `json:"text"`
	RequestID string `json:"request_id"`
}

// EmbeddingResponse represents the response of an embedding request
type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
	RequestID string    `json:"request_id"`
	Error     string    `json:"error,omitempty"`
}

// RAGQuery represents a query to the RAG service
type RAGQuery struct {
	Query      string  `json:"query"`
	Collection string  `json:"collection"`
	TopK       int     `json:"top_k"`
	Threshold  float64 `json:"threshold"`
}

// RAGResponse represents the response of a RAG query
type RAGResponse struct {
	Query     string        `json:"query"`
	Documents []RAGDocument `json:"documents"`
	TotalHits int           `json:"total_hits"`
}

// RAGDocument represents a document in the RAG service
type RAGDocument struct {
	Content  string            `json:"content"`
	Score    float64           `json:"score"`
	Metadata map[string]string `json:"metadata"`
	Source   string            `json:"source"`
}

// WorkerRole represents the role of a worker
type WorkerRole string

// WorkflowStage represents the stage of a workflow
type WorkflowStage string

// Constants for worker roles and workflow stages
const (
	RoleDeveloper WorkerRole = "developer"
	RoleReviewer  WorkerRole = "reviewer"
	RoleApprover  WorkerRole = "approver"
	RoleTester    WorkerRole = "tester"

	StageDevelopment WorkflowStage = "development"
	StageReview      WorkflowStage = "review"
	StageApproval    WorkflowStage = "approval"
	StageTesting     WorkflowStage = "testing"
	StageCompleted   WorkflowStage = "completed"

	EmbeddingRequestTopic  = "embedding/request"
	EmbeddingResponseTopic = "embedding/response"
)
