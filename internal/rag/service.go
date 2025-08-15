package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// Service provides RAG functionality using qdrant
type Service struct {
	qdrantBinary string
	qdrantURL    string
	collections  map[string]string // collection name -> description
}

// NewService creates a new RAG service
func NewService(qdrantBinary, qdrantURL string) *Service {
	return &Service{
		qdrantBinary: qdrantBinary,
		qdrantURL:    qdrantURL,
		collections: map[string]string{
			"coding_standards": "Best practices and coding standards",
			"documentation":    "Technical documentation and guides", 
			"code_examples":    "Code examples and patterns",
			"book_expert":      "Technical book content and knowledge",
		},
	}
}

// SearchKnowledge searches the knowledge base for relevant information
func (s *Service) SearchKnowledge(ctx context.Context, query types.RAGQuery) (*types.RAGResponse, error) {
	// For now, use a simple text-based search approach
	// In production, this would use proper vector embeddings
	
	searchCmd := fmt.Sprintf(`curl -s "%s/collections/%s/points/search" \
		-H "Content-Type: application/json" \
		-d '{
			"vector": null,
			"filter": null,
			"limit": %d,
			"with_payload": true,
			"with_vector": false
		}'`, s.qdrantURL, query.Collection, query.TopK)
	
	cmd := exec.CommandContext(ctx, "bash", "-c", searchCmd)
	output, err := cmd.Output()
	if err != nil {
		// Fallback to simple knowledge base
		return s.fallbackSearch(query), nil
	}
	
	// Parse qdrant response (simplified)
	var qdrantResp struct {
		Result []struct {
			Payload map[string]interface{} `json:"payload"`
			Score   float64               `json:"score"`
		} `json:"result"`
	}
	
	if err := json.Unmarshal(output, &qdrantResp); err != nil {
		return s.fallbackSearch(query), nil
	}
	
	// Convert to our format
	response := &types.RAGResponse{
		Query:     query.Query,
		TotalHits: len(qdrantResp.Result),
		Documents: make([]types.RAGDocument, 0, len(qdrantResp.Result)),
	}
	
	for _, result := range qdrantResp.Result {
		doc := types.RAGDocument{
			Score:    result.Score,
			Metadata: make(map[string]string),
		}
		
		// Extract content and metadata
		if content, ok := result.Payload["content"].(string); ok {
			doc.Content = content
		}
		if source, ok := result.Payload["source"].(string); ok {
			doc.Source = source
		}
		
		// Convert metadata
		for k, v := range result.Payload {
			if str, ok := v.(string); ok {
				doc.Metadata[k] = str
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
	cmd := exec.CommandContext(ctx, "curl", "-s", fmt.Sprintf("%s/health", s.qdrantURL))
	err := cmd.Run()
	return err == nil
}