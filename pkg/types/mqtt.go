package types

// MQTT Topic Constants
// These constants define the topic structure for the MQTT Agent Orchestration System
// Following the "Never hard code values" design principle

const (
	// Task-related topics
	TaskNewTopic        = "tasks/new"
	TaskResultTopic     = "tasks/result"
	TaskFailedTopic     = "tasks/failed"
	TaskStatusTopic     = "tasks/status"
	
	// Role-specific task topics
	TaskDeveloperTopic  = "tasks/developer"
	TaskReviewerTopic   = "tasks/reviewer"
	TaskApproverTopic   = "tasks/approver"
	TaskTesterTopic     = "tasks/tester"
	
	// Worker status topics
	WorkerStatusTopic   = "workers/status"
	WorkerHeartbeatTopic = "workers/heartbeat"
	
	// Embedding service topics
	EmbeddingRequestTopic  = "embeddings/request"
	EmbeddingResponseTopic = "embeddings/response"
	
	// AI service topics
	AIRequestTopicPrefix   = "ai/request/"   // e.g., "ai/request/cerebras"
	AIResponseTopicPrefix  = "ai/response/"  // e.g., "ai/response/cerebras"
	
	// RAG service topics
	RAGQueryTopic     = "rag/query"
	RAGResponseTopic  = "rag/response"
	RAGTrainTopic     = "rag/train"
	
	// System control topics
	SystemHealthTopic = "system/health"
	SystemConfigTopic = "system/config"
)

// GetRoleTaskTopic returns the topic for a specific role
func GetRoleTaskTopic(role string) string {
	switch role {
	case "developer":
		return TaskDeveloperTopic
	case "reviewer":
		return TaskReviewerTopic
	case "approver":
		return TaskApproverTopic
	case "tester":
		return TaskTesterTopic
	default:
		return TaskNewTopic
	}
}

// GetAIRequestTopic returns the AI request topic for a specific provider
func GetAIRequestTopic(provider string) string {
	return AIRequestTopicPrefix + provider
}

// GetAIResponseTopic returns the AI response topic for a specific provider
func GetAIResponseTopic(provider string) string {
	return AIResponseTopicPrefix + provider
}
