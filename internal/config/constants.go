package config

import "time"

// Configuration constants following "Never hard code values" principle
const (
	// Vector dimensions
	VectorDimension = 2560 // Qwen3-Embedding-4B dimension
	
	// Timeouts
	MQTTConnectTimeout     = 10 * time.Second
	MQTTPublishTimeout     = 5 * time.Second
	EmbeddingRequestTimeout = 30 * time.Second
	EmbeddingInitTimeout   = 5 * time.Second
	
	// Cache limits
	MaxTextSize = 64 * 1024 // 64KB limit for text input
	
	// Collection names
	AgentPromptsCollection    = "agent_prompts"
	CodingStandardsCollection = "coding_standards"
	DocumentationCollection   = "documentation"
	CodeExamplesCollection    = "code_examples"
	BookExpertCollection      = "book_expert"
	
	// MQTT topics
	EmbeddingRequestTopic  = "embedding/requests"
	EmbeddingResponseTopic = "embedding/responses"
	
	// Default ports
	DefaultQdrantPort = 6333
	DefaultMQTTPort   = 1883
	
	// Retry configuration
	MaxRetries          = 3
	RetryBackoffInitial = 1 * time.Second
	RetryBackoffMax     = 30 * time.Second
)
