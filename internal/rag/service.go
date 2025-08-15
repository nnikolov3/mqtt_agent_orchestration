package rag

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/niko/mqtt-agent-orchestration/pkg/types"
	"github.com/qdrant/go-client/qdrant"
)

// Service provides RAG functionality using qdrant
type Service struct {
	client      *qdrant.Client
	qdrantURL   string
	collections map[string]string // collection name -> description
}

// NewService creates a new RAG service
func NewService(qdrantBinary, qdrantURL string) *Service {
	// Parse URL to get host and port
	var host string
	var port int
	if qdrantURL == "" {
		host = "localhost"
		port = 6333
	} else {
		// Simple URL parsing for localhost:port format
		if strings.Contains(qdrantURL, "localhost:6333") || strings.Contains(qdrantURL, "127.0.0.1:6333") {
			host = "localhost"
			port = 6333
		} else {
			host = "localhost"
			port = 6333
			log.Printf("Warning: Using default Qdrant connection (localhost:6333)")
		}
	}
	
	// Create Qdrant client
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		log.Printf("Warning: Failed to create Qdrant client: %v", err)
		client = nil
	}
	
	return &Service{
		client:    client,
		qdrantURL: qdrantURL,
		collections: map[string]string{
			"agent_prompts":    "System prompts for each worker role",
			"coding_standards": "Best practices and coding standards",
			"documentation":    "Technical documentation and guides", 
			"code_examples":    "Code examples and patterns",
			"book_expert":      "Technical book content and knowledge",
		},
	}
}

// InitializeCollections creates collections if they don't exist
func (s *Service) InitializeCollections(ctx context.Context) error {
	if s.client == nil {
		log.Printf("Qdrant client not available, skipping collection initialization")
		return nil
	}
	
	// Create agent_prompts collection for system prompts
	collectionName := "agent_prompts"
	_, err := s.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     384, // sentence-transformers/all-MiniLM-L6-v2 dimension
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		log.Printf("Collection %s may already exist: %v", collectionName, err)
	}
	
	return nil
}

// StoreSystemPrompt stores a system prompt for a worker role
func (s *Service) StoreSystemPrompt(ctx context.Context, role types.WorkerRole, prompt string) error {
	if s.client == nil {
		log.Printf("Qdrant client not available, cannot store system prompt")
		return fmt.Errorf("qdrant client not available")
	}
	
	// For now, use a simple embedding (in production, use proper embedding service)
	embedding := s.generateSimpleEmbedding(prompt)
	
	// Create point
	point := &qdrant.PointStruct{
		Id:      qdrant.NewIDStr(string(role)),
		Vectors: qdrant.NewVectors(embedding...),
		Payload: qdrant.NewValueMap(map[string]any{
			"role":         string(role),
			"prompt":       prompt,
			"content_type": "system_prompt",
			"updated_at":   fmt.Sprintf("%d", ctx.Value("timestamp")),
		}),
	}
	
	_, err := s.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: "agent_prompts",
		Points:         []*qdrant.PointStruct{point},
	})
	
	if err != nil {
		return fmt.Errorf("failed to store system prompt: %w", err)
	}
	
	log.Printf("Stored system prompt for role: %s", role)
	return nil
}

// GetSystemPrompt retrieves the system prompt for a worker role
func (s *Service) GetSystemPrompt(ctx context.Context, role types.WorkerRole) (string, error) {
	if s.client == nil {
		// Fallback to hardcoded prompts
		return s.getFallbackSystemPrompt(role), nil
	}
	
	// Query by role
	searchResult, err := s.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: "agent_prompts",
		Query:          qdrant.NewQueryID(qdrant.NewIDStr(string(role))),
		Limit:          qdrant.PtrOf(uint64(1)),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	
	if err != nil {
		log.Printf("Failed to query system prompt, using fallback: %v", err)
		return s.getFallbackSystemPrompt(role), nil
	}
	
	if len(searchResult.Points) == 0 {
		log.Printf("No system prompt found for role %s, using fallback", role)
		return s.getFallbackSystemPrompt(role), nil
	}
	
	// Extract prompt from payload
	if payload := searchResult.Points[0].Payload; payload != nil {
		if promptValue, ok := payload.GetKind().(*qdrant.Value_StructValue); ok {
			if promptField, exists := promptValue.StructValue.Fields["prompt"]; exists {
				if stringValue, ok := promptField.GetKind().(*qdrant.Value_StringValue); ok {
					return stringValue.StringValue, nil
				}
			}
		}
	}
	
	log.Printf("Could not extract prompt from payload, using fallback")
	return s.getFallbackSystemPrompt(role), nil
}

// SearchKnowledge searches the knowledge base for relevant information
func (s *Service) SearchKnowledge(ctx context.Context, query types.RAGQuery) (*types.RAGResponse, error) {
	if s.client == nil {
		return s.fallbackSearch(query), nil
	}
	
	// Generate embedding for query
	queryEmbedding := s.generateSimpleEmbedding(query.Query)
	
	// Search in Qdrant
	searchResult, err := s.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: query.Collection,
		Query:          qdrant.NewQuery(queryEmbedding...),
		Limit:          qdrant.PtrOf(uint64(query.TopK)),
		ScoreThreshold: qdrant.PtrOf(float32(query.Threshold)),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	
	if err != nil {
		log.Printf("Qdrant search failed, using fallback: %v", err)
		return s.fallbackSearch(query), nil
	}
	
	// Convert to our format
	response := &types.RAGResponse{
		Query:     query.Query,
		TotalHits: len(searchResult.Points),
		Documents: make([]types.RAGDocument, 0, len(searchResult.Points)),
	}
	
	for _, point := range searchResult.Points {
		doc := types.RAGDocument{
			Score:    float64(point.Score),
			Metadata: make(map[string]string),
		}
		
		// Extract content and metadata from payload
		if payload := point.Payload; payload != nil {
			if structValue, ok := payload.GetKind().(*qdrant.Value_StructValue); ok {
				fields := structValue.StructValue.Fields
				
				if contentField, exists := fields["content"]; exists {
					if stringValue, ok := contentField.GetKind().(*qdrant.Value_StringValue); ok {
						doc.Content = stringValue.StringValue
					}
				}
				
				if sourceField, exists := fields["source"]; exists {
					if stringValue, ok := sourceField.GetKind().(*qdrant.Value_StringValue); ok {
						doc.Source = stringValue.StringValue
					}
				}
				
				// Extract other metadata
				for key, field := range fields {
					if stringValue, ok := field.GetKind().(*qdrant.Value_StringValue); ok {
						doc.Metadata[key] = stringValue.StringValue
					}
				}
			}
		}
		
		response.Documents = append(response.Documents, doc)
	}
	
	return response, nil
}

// fallbackSearch provides simple keyword-based knowledge when qdrant is unavailable
func (s *Service) fallbackSearch(query types.RAGQuery) *types.RAGResponse {
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
	queryLower := strings.ToLower(query.Query)
	
	for collection, collectionDocs := range knowledge {
		if query.Collection == "" || query.Collection == collection {
			for _, doc := range collectionDocs {
				if strings.Contains(strings.ToLower(doc.Content), queryLower) ||
				   strings.Contains(strings.ToLower(doc.Source), queryLower) {
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
		Documents: docs,
		TotalHits: len(docs),
	}
}

// GetRelevantContext gets context for a specific task type
func (s *Service) GetRelevantContext(ctx context.Context, taskType, content string) (string, error) {
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
	
	return strings.Join(contextParts, "\n\n"), nil
}

// IsAvailable checks if qdrant service is available
func (s *Service) IsAvailable(ctx context.Context) bool {
	if s.client == nil {
		return false
	}
	
	// Try to list collections as a health check
	_, err := s.client.ListCollections(ctx)
	return err == nil
}

// generateSimpleEmbedding creates a simple embedding for text (placeholder for real embedding service)
func (s *Service) generateSimpleEmbedding(text string) []float32 {
	// This is a placeholder - in production, use a proper embedding model
	// such as sentence-transformers, OpenAI embeddings, or local models
	
	embedding := make([]float32, 384) // Matching sentence-transformers dimension
	
	// Simple hash-based embedding for demo purposes
	hash := 0
	for i, char := range text {
		hash = hash*31 + int(char)
		if i < len(embedding) {
			embedding[i] = float32((hash % 200) - 100) / 100.0 // Normalize to [-1, 1]
		}
	}
	
	// Normalize the vector
	var magnitude float32
	for _, val := range embedding {
		magnitude += val * val
	}
	magnitude = float32(1.0) / float32(1.0+magnitude) // Simple normalization
	
	for i := range embedding {
		embedding[i] *= magnitude
	}
	
	return embedding
}

// getFallbackSystemPrompt provides hardcoded system prompts when Qdrant is unavailable
func (s *Service) getFallbackSystemPrompt(role types.WorkerRole) string {
	prompts := map[types.WorkerRole]string{
		types.RoleDeveloper: `You are an expert Go developer tasked with creating high-quality code and documentation.

Core Responsibilities:
- Generate clean, idiomatic Go code following best practices
- Write comprehensive documentation with clear examples
- Follow coding standards and design principles
- Ensure proper error handling and testing considerations

Key Principles:
- Explicit is better than implicit
- Composition over inheritance
- Handle errors explicitly, never ignore them
- Use gofmt, go vet, and golangci-lint standards
- Write self-documenting code with clear naming
- Follow SOLID principles and clean architecture

Output Format:
Provide complete, working implementations with:
1. Clear package and import declarations
2. Proper struct and interface definitions
3. Well-documented functions with godoc comments
4. Appropriate error handling
5. Example usage when applicable

Focus on creating production-ready code that is maintainable, testable, and follows Go idioms.`,

		types.RoleReviewer: `You are a senior code reviewer with expertise in Go development and software engineering best practices.

Core Responsibilities:
- Review code for quality, correctness, and adherence to standards
- Identify potential bugs, security issues, and performance problems
- Ensure code follows established patterns and conventions
- Provide constructive feedback for improvements

Review Criteria:
- Code correctness and logic
- Error handling and edge cases
- Performance and memory efficiency
- Security considerations
- Maintainability and readability
- Test coverage and testability
- Documentation quality

Output Format:
Provide structured feedback with:
1. Overall assessment (APPROVED/NEEDS_CHANGES)
2. Specific issues found with line references
3. Suggested improvements
4. Best practice recommendations

Always be constructive and educational in your feedback. Focus on making the code better while explaining the reasoning behind suggestions.`,

		types.RoleApprover: `You are a technical lead responsible for final approval of code and documentation before release.

Core Responsibilities:
- Make final go/no-go decisions on implementations
- Ensure alignment with project goals and requirements
- Verify quality standards are met
- Check for consistency across the codebase

Approval Criteria:
- Meets functional requirements
- Follows coding standards and best practices
- Has adequate error handling and testing
- Is properly documented
- Aligns with architectural decisions
- Ready for production use

Output Format:
Provide clear decision with:
1. Decision: APPROVED or REJECTED
2. Rationale for the decision
3. Any final requirements before approval (if rejected)
4. Acknowledgment of quality and completeness (if approved)

Be decisive but fair. Only approve work that truly meets production standards.`,

		types.RoleTester: `You are a quality assurance engineer focused on testing and validation of code and systems.

Core Responsibilities:
- Verify implementations work correctly
- Test edge cases and error conditions
- Validate documentation accuracy
- Ensure code is testable and maintainable

Testing Focus Areas:
- Functional correctness
- Error handling and edge cases
- Performance characteristics
- Security vulnerabilities
- Integration points
- Documentation accuracy

Output Format:
Provide test results with:
1. Test Status: PASSED or FAILED
2. Test scenarios covered
3. Issues found (if any)
4. Recommendations for improvement
5. Verification of documentation accuracy

Be thorough in testing but also practical. Focus on the most important scenarios and edge cases that could impact production use.`,
	}
	
	if prompt, exists := prompts[role]; exists {
		return prompt
	}
	
	return "You are a helpful AI assistant focused on high-quality software development."
}