package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/niko/mqtt-agent-orchestration/internal/mqtt"
	"github.com/niko/mqtt-agent-orchestration/internal/rl"
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
	"github.com/qdrant/go-client/qdrant"
)

// Service provides RAG functionality using qdrant
type Service struct {
	client      *qdrant.Client
	qdrantURL   string
	collections map[string]string                        // collection name -> description
	mqttClient  *mqtt.Client                             // Long-lived MQTT client for embedding requests
	pendingReqs map[string]chan *types.EmbeddingResponse // Request correlation map
	mu          sync.RWMutex                             // Thread safety for pending requests
	initialized bool                                     // Track if MQTT client is initialized
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
		// Handle both URL formats: "http://localhost:6333" and "localhost:6333"
		cleanURL := qdrantURL
		if strings.HasPrefix(qdrantURL, "http://") {
			cleanURL = strings.TrimPrefix(qdrantURL, "http://")
		} else if strings.HasPrefix(qdrantURL, "https://") {
			cleanURL = strings.TrimPrefix(qdrantURL, "https://")
		}

		// Use Go's standard library for robust URL parsing
		parsedHost, portStr, err := net.SplitHostPort(cleanURL)
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
		pendingReqs: make(map[string]chan *types.EmbeddingResponse),
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

// Updated StoreSystemPrompt with RL feedback
func (s *Service) StoreSystemPrompt(ctx context.Context, role types.WorkerRole, prompt string) error {
	start := time.Now()

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

	duration := time.Since(start)

	collector := rl.NewFeedbackCollector()

	taskContext := rl.TaskContext{
		OriginalPrompt: prompt,
		TaskType:       "store_system_prompt",
		Complexity:     "low",
		Mode:           "LOCAL",
	}

	metrics := rl.ExecutionMetrics{
		Duration:    duration,
		SuccessRate: 0.0, // Will be set based on success
		ErrorCount:  0,
	}

	taskID := fmt.Sprintf("rag-store-%s-%d", role, time.Now().UnixNano())
	workerID := "rag-service"
	provider := "local"
	model := "qdrant"

	if err != nil {
		metrics.SuccessRate = 0.0
		metrics.ErrorCount = 1
		taskContext.Response = fmt.Sprintf("Failed: %v", err)

		fbErr := collector.CollectFailureFeedback(
			taskID,
			workerID,
			provider,
			model,
			taskContext,
			err.Error(),
			metrics,
		)
		if fbErr != nil {
			log.Printf("Warning: Failed to collect failure feedback: %v", fbErr)
		}
		return fmt.Errorf("failed to store system prompt: %w", err)
	}

	metrics.SuccessRate = 1.0
	metrics.ErrorCount = 0
	taskContext.Response = "Stored successfully"

	fbErr := collector.CollectSuccessFeedback(
		taskID,
		workerID,
		provider,
		model,
		taskContext,
		metrics,
	)
	if fbErr != nil {
		log.Printf("Warning: Failed to collect success feedback: %v", fbErr)
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

// initMQTT initializes the long-lived MQTT client for embedding requests
// Following "Performance and Efficiency" - reuse connections instead of creating new ones
func (s *Service) initMQTT(ctx context.Context) error {
	if s.initialized {
		return nil
	}

	// Create MQTT client with unique ID for this service
	s.mqttClient = mqtt.NewClientWithID("localhost", 1883, "rag-service")

	// Connect to MQTT broker with timeout
	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.mqttClient.Connect(connectCtx); err != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", err)
	}

	// Subscribe to embedding responses once
	if err := s.mqttClient.Subscribe(ctx, types.EmbeddingResponseTopic, s.handleEmbeddingResponse); err != nil {
		return fmt.Errorf("failed to subscribe to embedding responses: %w", err)
	}

	s.initialized = true
	log.Printf("MQTT client initialized for embedding requests")
	return nil
}

// handleEmbeddingResponse routes embedding responses to waiting goroutines
// Following "Thread Safety" - uses mutex for safe concurrent access
func (s *Service) handleEmbeddingResponse(payload []byte) {
	var resp types.EmbeddingResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		log.Printf("ERROR: Failed to unmarshal embedding response: %v", err)
		return
	}

	// Thread-safe lookup of response channel
	s.mu.RLock()
	responseChan, exists := s.pendingReqs[resp.RequestID]
	s.mu.RUnlock()

	if exists {
		// Send response with timeout to prevent blocking
		select {
		case responseChan <- &resp:
			// Response delivered successfully
		case <-time.After(100 * time.Millisecond):
			log.Printf("WARNING: Response channel timeout for request %s", resp.RequestID)
		}
	} else {
		log.Printf("WARNING: Received response for unknown request %s", resp.RequestID)
	}
}

// generateLocalEmbedding generates embeddings using long-lived MQTT client
// Following "Performance and Efficiency" - reuses connection and implements proper correlation
func (s *Service) generateLocalEmbedding(text string) []float32 {
	// Initialize MQTT client if not already done
	if !s.initialized {
		if err := s.initMQTT(context.Background()); err != nil {
			log.Printf("ERROR: Failed to initialize MQTT client: %v", err)
			return nil
		}
	}

	// Create embedding request with unique ID
	reqID := fmt.Sprintf("emb-%d", time.Now().UnixNano())
	req := types.EmbeddingRequest{
		Text:      text,
		RequestID: reqID,
	}

	// Marshal request
	payload, err := json.Marshal(req)
	if err != nil {
		log.Printf("ERROR: Failed to marshal embedding request: %v", err)
		return nil
	}

	// Create response channel and register request
	responseChan := make(chan *types.EmbeddingResponse, 1)

	s.mu.Lock()
	s.pendingReqs[reqID] = responseChan
	s.mu.Unlock()

	// Cleanup function to remove request from map
	defer func() {
		s.mu.Lock()
		delete(s.pendingReqs, reqID)
		s.mu.Unlock()
		close(responseChan)
	}()

	// Publish embedding request with timeout
	publishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.mqttClient.Publish(publishCtx, types.EmbeddingRequestTopic, payload); err != nil {
		log.Printf("ERROR: Failed to publish embedding request: %v", err)
		return nil
	}

	// Wait for response with timeout
	select {
	case resp := <-responseChan:
		if resp.Error != "" {
			log.Printf("ERROR: Embedding service error: %s", resp.Error)
			return nil
		}

		if len(resp.Embedding) != 2560 {
			log.Printf("WARNING: Expected 2560-dim embedding, got %d dimensions", len(resp.Embedding))
		}

		return resp.Embedding

	case <-time.After(30 * time.Second):
		log.Printf("ERROR: Embedding request timeout for request %s", reqID)
		return nil
	}
}

// Close gracefully shuts down the RAG service
// Following "Graceful Degradation" - clean shutdown of resources
func (s *Service) Close() error {
	if s.mqttClient != nil {
		s.mqttClient.Disconnect()
		log.Printf("MQTT client disconnected")
	}

	// Clear any pending requests
	s.mu.Lock()
	for reqID, ch := range s.pendingReqs {
		close(ch)
		delete(s.pendingReqs, reqID)
	}
	s.mu.Unlock()

	log.Printf("RAG service shutdown complete")
	return nil
}

// hashString creates a consistent hash for string values
func hashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
