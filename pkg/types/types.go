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
	Status       WorkflowStatus         `json:"status"`
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
	WorkflowID     string         `json:"workflow_id"`
	Stage          WorkflowStage  `json:"stage"`
	WorkerRole     WorkerRole     `json:"worker_role"`
	Approved       bool           `json:"approved"`
	RequiresRetry  bool           `json:"requires_retry"`
	ReviewFeedback string         `json:"review_feedback"`
	Status         WorkflowStatus `json:"status"`
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
	ID       string            `json:"id"`
	Content  string            `json:"content"`
	Score    float64           `json:"score"`
	Metadata map[string]string `json:"metadata"`
	Source   string            `json:"source"`
}

// WorkerRole represents the role of a worker
type WorkerRole string

// WorkflowStage represents the stage of a workflow
type WorkflowStage string

// WorkflowStatus represents the status of a workflow
type WorkflowStatus string

// Constants for worker roles and workflow stages
const (
	RoleDeveloper WorkerRole = "developer"
	RoleReviewer  WorkerRole = "reviewer"
	RoleApprover  WorkerRole = "approver"
	RoleTester    WorkerRole = "tester"

	// Cerebras model roles
	RoleCerebrasGptOss120b     WorkerRole = "cerebras_gpt_oss_120b"
	RoleCerebrasQwen3Coder480b WorkerRole = "cerebras_qwen_3_coder_480b"
	RoleCerebrasQwen332b       WorkerRole = "cerebras_qwen_3_32b"
	RoleCerebrasLlama3370b     WorkerRole = "cerebras_llama_3.3_70b"

	// NVIDIA model roles
	RoleNvidiaLlama33NemotronSuper49bV15 WorkerRole = "nvidia_llama_3.3_nemotron_super_49b_v1.5"
	RoleNvidiaOpenaiGptOss120b           WorkerRole = "nvidia_openai_gpt_oss_120b"
	RoleNvidiaNemotron4340bInstruct      WorkerRole = "nvidia_nemotron_4_340b_instruct"
	RoleNvidiaMetaLlama318bInstruct      WorkerRole = "nvidia_meta_llama_3.1_8b_instruct"

	// NVIDIA OCR role
	RoleNvidiaNemoretrieverOcrV1 WorkerRole = "nvidia_nemoretriever_ocr_v1"

	// Gemini model roles
	RoleGemini25Pro   WorkerRole = "gemini_2.5_pro"
	RoleGemini25Flash WorkerRole = "gemini_2.5_flash"
	RoleGemini20Flash WorkerRole = "gemini_2.0_flash"
	RoleGemini15Flash WorkerRole = "gemini_1.5_flash"

	// Grok model roles
	RoleGrok40709 WorkerRole = "grok_4_0709"
	RoleGrok3     WorkerRole = "grok_3"
	RoleGrok3Mini WorkerRole = "grok_3_mini"

	// Groq model roles
	RoleGroqMoonshotaiKimiK2Instruct  WorkerRole = "groq_moonshotai_kimi_k2_instruct"
	RoleGroqLlama3370bVersatile       WorkerRole = "groq_llama_3.3_70b_versatile"
	RoleGroqDeepseekR1DistillLlama70b WorkerRole = "groq_deepseek_r1_distill_llama_70b"
	RoleGroqLlama370b8192             WorkerRole = "groq_llama3_70b_8192"
	RoleGroqLlama318bInstant          WorkerRole = "groq_llama_3.1_8b_instant"

	StageDevelopment WorkflowStage = "development"
	StageReview      WorkflowStage = "review"
	StageApproval    WorkflowStage = "approval"
	StageTesting     WorkflowStage = "testing"
	StageCompleted   WorkflowStage = "completed"
	StageFailed      WorkflowStage = "failed"

	StatusInProgress WorkflowStatus = "in_progress"
	StatusCompleted  WorkflowStatus = "completed"
	StatusFailed     WorkflowStatus = "failed"

	EmbeddingRequestTopic  = "embedding/request"
	EmbeddingResponseTopic = "embedding/response"
)
