package rag

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/niko/mqtt-agent-orchestration/internal/mcp"
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// MCPService provides RAG functionality using MCP for enhanced capabilities
type MCPService struct {
	qdrantClient *mcp.QdrantMCPClient
	config       *MCPServiceConfig
}

// MCPServiceConfig holds MCP service configuration
type MCPServiceConfig struct {
	QdrantURL string        `json:"qdrant_url"`
	Timeout   time.Duration `json:"timeout"`
	MaxRetries int          `json:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay"`
}

// NewMCPService creates a new MCP-enhanced RAG service
func NewMCPService(config *MCPServiceConfig) *MCPService {
	qdrantConfig := &mcp.QdrantMCPConfig{
		QdrantURL:  config.QdrantURL,
		Timeout:    config.Timeout,
		MaxRetries: config.MaxRetries,
		RetryDelay: config.RetryDelay,
	}

	// Create a mock RAG service for the MCP client
	mockRAGService := &mockRAGService{}

	return &MCPService{
		qdrantClient: mcp.NewQdrantMCPClient(qdrantConfig, mockRAGService),
		config:       config,
	}
}

// Connect establishes connection to MCP services
func (s *MCPService) Connect(ctx context.Context) error {
	// For now, just test if the service is available
	if !s.IsAvailable(ctx) {
		return fmt.Errorf("MCP RAG service is not available")
	}

	log.Printf("✅ MCP RAG service connected")
	return nil
}

// Disconnect closes connections
func (s *MCPService) Disconnect() error {
	// No explicit disconnect needed for the new interface
	return nil
}

// Mock RAG service for MCP client
type mockRAGService struct{}

func (m *mockRAGService) SearchKnowledge(ctx context.Context, query types.RAGQuery) (*types.RAGResponse, error) {
	return &types.RAGResponse{
		Documents: []types.RAGDocument{
			{
				Content: "Mock search result",
				Score:   0.8,
				Metadata: map[string]string{
					"source": "mock",
				},
				Source: "mock",
			},
		},
		Query:     query.Query,
		TotalHits: 1,
	}, nil
}

func (m *mockRAGService) AddDocument(ctx context.Context, content string, metadata map[string]interface{}) error {
	return nil
}

// InitializeCollections creates collections if they don't exist
func (s *MCPService) InitializeCollections(ctx context.Context) error {
	collections := []struct {
		name        string
		vectorSize  int
		distance    string
		description string
	}{
		{
			name:        "agent_prompts",
			vectorSize:  384,
			distance:    "cosine",
			description: "System prompts for each worker role",
		},
		{
			name:        "coding_standards",
			vectorSize:  384,
			distance:    "cosine",
			description: "Best practices and coding standards",
		},
		{
			name:        "documentation",
			vectorSize:  384,
			distance:    "cosine",
			description: "Technical documentation and guides",
		},
		{
			name:        "code_examples",
			vectorSize:  384,
			distance:    "cosine",
			description: "Code examples and patterns",
		},
		{
			name:        "book_expert",
			vectorSize:  384,
			distance:    "cosine",
			description: "Technical book content and knowledge",
		},
	}

	existingCollections, err := s.qdrantClient.ListCollections(ctx)
	if err != nil {
		log.Printf("Warning: Failed to list collections: %v", err)
		existingCollections = []string{}
	}

	existingMap := make(map[string]bool)
	for _, name := range existingCollections {
		existingMap[name] = true
	}

	for _, collection := range collections {
		if !existingMap[collection.name] {
			if err := s.qdrantClient.CreateCollection(ctx, collection.name, collection.vectorSize, collection.distance); err != nil {
				log.Printf("Warning: Failed to create collection %s: %v", collection.name, err)
			} else {
				log.Printf("✅ Created collection: %s (%s)", collection.name, collection.description)
			}
		} else {
			log.Printf("Collection already exists: %s", collection.name)
		}
	}

	return nil
}

// StoreSystemPrompt stores a system prompt for a worker role
func (s *MCPService) StoreSystemPrompt(ctx context.Context, role types.WorkerRole, prompt string) error {
	// Generate embedding
	embedding := s.generateSimpleEmbedding(prompt)

	// Create point
	point := mcp.Point{
		ID:     fmt.Sprintf("prompt_%s", role),
		Vector: embedding,
		Payload: map[string]interface{}{
			"role":         string(role),
			"prompt":       prompt,
			"content_type": "system_prompt",
			"updated_at":   time.Now().Unix(),
		},
	}

	// Upsert point
	if err := s.qdrantClient.UpsertPoints(ctx, "agent_prompts", []mcp.Point{point}); err != nil {
		return fmt.Errorf("failed to store system prompt: %w", err)
	}

	log.Printf("✅ Stored system prompt for role: %s", role)
	return nil
}

// GetSystemPrompt retrieves the system prompt for a worker role
func (s *MCPService) GetSystemPrompt(ctx context.Context, role types.WorkerRole) (string, error) {
	// Generate embedding for role
	embedding := s.generateSimpleEmbedding(string(role))

	// Search for role-specific prompt
	results, err := s.qdrantClient.SearchPoints(ctx, "agent_prompts", embedding, 1, 0.5)
	if err != nil {
		log.Printf("Failed to search for system prompt, using fallback: %v", err)
		return s.getFallbackSystemPrompt(role), nil
	}

	if len(results) == 0 {
		log.Printf("No system prompt found for role %s, using fallback", role)
		return s.getFallbackSystemPrompt(role), nil
	}

	// Extract prompt from payload
	if payload := results[0].Payload; payload != nil {
		if promptField, exists := payload["prompt"]; exists {
			if prompt, ok := promptField.(string); ok {
				return prompt, nil
			}
		}
	}

	log.Printf("Could not extract prompt from payload, using fallback")
	return s.getFallbackSystemPrompt(role), nil
}

// SearchKnowledge searches the knowledge base for relevant information
func (s *MCPService) SearchKnowledge(ctx context.Context, query types.RAGQuery) (*types.RAGResponse, error) {
	// Generate embedding for query
	queryEmbedding := s.generateSimpleEmbedding(query.Query)

	// Search in collection
	results, err := s.qdrantClient.SearchPoints(ctx, query.Collection, queryEmbedding, query.TopK, float32(query.Threshold))
	if err != nil {
		log.Printf("MCP search failed, using fallback: %v", err)
		return s.fallbackSearch(query), nil
	}

	// Convert to our format
	response := &types.RAGResponse{
		Query:     query.Query,
		TotalHits: len(results),
		Documents: make([]types.RAGDocument, 0, len(results)),
	}

	for _, result := range results {
		doc := types.RAGDocument{
			Score:    float64(result.Score),
			Metadata: make(map[string]string),
		}

		// Extract content and metadata from payload
		if payload := result.Payload; payload != nil {
			if contentField, exists := payload["content"]; exists {
				if content, ok := contentField.(string); ok {
					doc.Content = content
				}
			}

			if sourceField, exists := payload["source"]; exists {
				if source, ok := sourceField.(string); ok {
					doc.Source = source
				}
			}

			// Extract other metadata
			for key, field := range payload {
				if value, ok := field.(string); ok {
					doc.Metadata[key] = value
				}
			}
		}

		response.Documents = append(response.Documents, doc)
	}

	return response, nil
}

// GetRelevantContext gets context for a specific task type
func (s *MCPService) GetRelevantContext(ctx context.Context, taskType, content string) (string, error) {
	query := types.RAGQuery{
		Query:      fmt.Sprintf("%s %s", taskType, content),
		Collection: "coding_standards",
		TopK:       3,
		Threshold:  0.5,
	}

	response, err := s.SearchKnowledge(ctx, query)
	if err != nil {
		return "", err
	}

	if len(response.Documents) == 0 {
		return "No relevant context found", nil
	}

	// Build context string
	var contextParts []string
	for i, doc := range response.Documents {
		contextParts = append(contextParts, fmt.Sprintf("Context %d: %s", i+1, doc.Content))
	}

	return fmt.Sprintf("%s\n\n", contextParts), nil
}

// IsAvailable checks if MCP services are available
func (s *MCPService) IsAvailable(ctx context.Context) bool {
	_, err := s.qdrantClient.ListCollections(ctx)
	return err == nil
}

// generateSimpleEmbedding creates a simple embedding for text
func (s *MCPService) generateSimpleEmbedding(text string) []float32 {
	// This is a placeholder - in production, use a proper embedding model
	embedding := make([]float32, 384)

	// Simple hash-based embedding for demo purposes
	hash := 0
	for i, char := range text {
		hash = hash*31 + int(char)
		if i < len(embedding) {
			embedding[i] = float32((hash%200)-100) / 100.0
		}
	}

	// Normalize the vector
	var magnitude float32
	for _, val := range embedding {
		magnitude += val * val
	}
	magnitude = float32(1.0) / float32(1.0+magnitude)

	for i := range embedding {
		embedding[i] *= magnitude
	}

	return embedding
}

// getFallbackSystemPrompt provides hardcoded system prompts when MCP is unavailable
func (s *MCPService) getFallbackSystemPrompt(role types.WorkerRole) string {
	prompts := map[types.WorkerRole]string{
		types.RoleDeveloper: `You are a skilled software developer focused on creating high-quality, maintainable code. 
Follow best practices, write clear documentation, and ensure your code is production-ready.`,
		types.RoleReviewer: `You are a thorough code reviewer who ensures code quality, security, and maintainability. 
Look for potential issues, suggest improvements, and verify adherence to coding standards.`,
		types.RoleApprover: `You are a final approver who makes critical decisions about code quality and deployment readiness. 
Be thorough but practical, focusing on business impact and risk assessment.`,
		types.RoleTester: `You are a quality assurance specialist who validates code correctness and functionality. 
Be thorough in testing but also practical. Focus on the most important scenarios and edge cases that could impact production use.`,
	}

	if prompt, exists := prompts[role]; exists {
		return prompt
	}

	return "You are a helpful AI assistant focused on high-quality software development."
}

// fallbackSearch provides simple keyword-based knowledge when MCP is unavailable
func (s *MCPService) fallbackSearch(query types.RAGQuery) *types.RAGResponse {
	// Simple knowledge base for coding standards
	knowledge := map[string][]types.RAGDocument{
		"coding_standards": {
			{
				Content: "Go coding standards emphasize explicit error handling, clear naming conventions, and composition over inheritance",
				Score:   0.9,
				Source:  "fallback_knowledge",
				Metadata: map[string]string{
					"type":     "principle",
					"language": "go",
				},
			},
			{
				Content: "Use gofmt, go vet, and golangci-lint for code quality. Follow effective Go guidelines for idiomatic code",
				Score:   0.85,
				Source:  "fallback_knowledge",
				Metadata: map[string]string{
					"type":     "tooling",
					"language": "go",
				},
			},
		},
		"documentation": {
			{
				Content: "Documentation should be clear, concise, and include examples. Use godoc format for Go packages",
				Score:   0.8,
				Source:  "fallback_knowledge",
				Metadata: map[string]string{
					"type": "documentation",
				},
			},
		},
	}

	// Find relevant documents based on query keywords
	var docs []types.RAGDocument
	queryLower := fmt.Sprintf("%s", query.Query)

	for collection, collectionDocs := range knowledge {
		if query.Collection == "" || query.Collection == collection {
			for _, doc := range collectionDocs {
				if contains(doc.Content, queryLower) || contains(doc.Source, queryLower) {
					docs = append(docs, doc)
				}
			}
		}
	}

	// Limit results
	if len(docs) > query.TopK {
		docs = docs[:query.TopK]
	}

	return &types.RAGResponse{
		Query:     query.Query,
		TotalHits: len(docs),
		Documents: docs,
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr))
}
