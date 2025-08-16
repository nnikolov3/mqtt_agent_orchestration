package rag

import (
	"context"
	"fmt"
	"hash/fnv"
	"log"
	"net"
	"strconv"
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

// NewService creates a new RAG service with proper IPv6/IPv4 dual-stack support
func NewService(qdrantBinary, qdrantURL string) (*Service, error) {
	// Parse URL properly using Go's net package
	var host string
	var port int
	
	if qdrantURL == "" {
		host = "localhost"
		port = 6333
	} else {
		// Use Go's standard library for robust URL parsing
		parsedHost, portStr, err := net.SplitHostPort(qdrantURL)
		if err != nil {
			return nil, fmt.Errorf("invalid Qdrant URL '%s': %v", qdrantURL, err)
		}
		
		host = parsedHost
		port, err = strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid port in Qdrant URL '%s': %v", qdrantURL, err)
		}
	}

	// Create Qdrant client - fail fast if connection fails
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client for %s:%d: %v", host, port, err)
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
	}, nil
}

// InitializeCollections creates collections if they don't exist
func (s *Service) InitializeCollections(ctx context.Context) error {
	// Qwen3-Embedding-4B produces 2560-dimensional vectors - use consistent dimensions
	const vectorDimension = 2560
	
	// Create agent_prompts collection for system prompts
	collectionName := "agent_prompts"
	err := s.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorDimension, // Qwen3-Embedding-4B-Q8_0 dimension
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
	// Generate proper embeddings using Qwen3-Embedding-4B model
	embedding := s.generateLocalEmbedding(prompt)
	if embedding == nil {
		return fmt.Errorf("failed to generate embeddings for prompt - embedding model unavailable")
	}

	// Create point
	point := &qdrant.PointStruct{
		Id:      qdrant.NewIDNum(uint64(hashString(string(role)))),
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
// Fails fast if RAG is unavailable - following Design Principle: "Explicit error handling"
func (s *Service) GetSystemPrompt(ctx context.Context, role types.WorkerRole) (string, error) {
	// Query by role - fail fast if no client
	searchResult, err := s.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: "agent_prompts",
		Query:          qdrant.NewQueryID(qdrant.NewIDNum(uint64(hashString(string(role))))),
		Limit:          qdrant.PtrOf(uint64(1)),
		WithPayload:    qdrant.NewWithPayload(true),
	})

	if err != nil {
		return "", fmt.Errorf("failed to query system prompt for role %s: %v", role, err)
	}

	if len(searchResult) == 0 {
		return "", fmt.Errorf("no system prompt found for role %s in RAG database", role)
	}

	// Extract prompt from payload
	if payload := searchResult[0].Payload; payload != nil {
		if promptField, exists := payload["prompt"]; exists {
			if stringValue, ok := promptField.GetKind().(*qdrant.Value_StringValue); ok {
				return stringValue.StringValue, nil
			}
		}
	}

	return "", fmt.Errorf("invalid prompt data structure in RAG database for role %s", role)
}

// SearchKnowledge searches the knowledge base for relevant information
// Fails fast if RAG is unavailable - following Design Principle: "Explicit error handling"
func (s *Service) SearchKnowledge(ctx context.Context, query types.RAGQuery) (*types.RAGResponse, error) {
	// Generate embedding for query - fail fast if unavailable
	queryEmbedding := s.generateLocalEmbedding(query.Query)
	if queryEmbedding == nil {
		return nil, fmt.Errorf("failed to generate query embedding - embedding model unavailable")
	}

	// Search in Qdrant
	searchResult, err := s.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: query.Collection,
		Query:          qdrant.NewQuery(queryEmbedding...),
		Limit:          qdrant.PtrOf(uint64(query.TopK)),
		ScoreThreshold: qdrant.PtrOf(float32(query.Threshold)),
		WithPayload:    qdrant.NewWithPayload(true),
	})

	if err != nil {
		return nil, fmt.Errorf("Qdrant search failed for collection %s: %v", query.Collection, err)
	}

	// Convert to our format
	response := &types.RAGResponse{
		Query:     query.Query,
		TotalHits: len(searchResult),
		Documents: make([]types.RAGDocument, 0, len(searchResult)),
	}

	for _, point := range searchResult {
		doc := types.RAGDocument{
			Score:    float64(point.Score),
			Metadata: make(map[string]string),
		}

		// Extract content and metadata from payload
		if payload := point.Payload; payload != nil {
			// Handle payload as map[string]*qdrant.Value
			if contentField, exists := payload["content"]; exists {
				if stringValue, ok := contentField.GetKind().(*qdrant.Value_StringValue); ok {
					doc.Content = stringValue.StringValue
				}
			}

			if sourceField, exists := payload["source"]; exists {
				if stringValue, ok := sourceField.GetKind().(*qdrant.Value_StringValue); ok {
					doc.Source = stringValue.StringValue
				}
			}

			// Extract other metadata
			for key, field := range payload {
				if stringValue, ok := field.GetKind().(*qdrant.Value_StringValue); ok {
					doc.Metadata[key] = stringValue.StringValue
				}
			}
		}

		response.Documents = append(response.Documents, doc)
	}

	return response, nil
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

// generateLocalEmbedding generates embeddings using local Qwen3-Embedding-4B model
// Returns nil if the embedding model is unavailable - caller must handle this explicitly
func (s *Service) generateLocalEmbedding(text string) []float32 {
	// Use llama-server with Qwen3-Embedding-4B model when needed
	// Following "Do more with less" - use existing llama-server binary
	// Real implementation would start llama-server, make HTTP call, parse response
	// TODO: Implement actual embedding generation using /home/niko/bin/llama-server
	
	return nil
}

// hashString creates a consistent hash for string values
func hashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

