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
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
	"github.com/qdrant/go-client/qdrant"
)

// Service provides RAG functionality using qdrant with caching and improved performance
type Service struct {
	client         *qdrant.Client
	qdrantURL      string
	collections    map[string]string
	mqttClient     *mqtt.Client
	pendingReqs    map[string]chan *types.EmbeddingResponse
	embeddingCache *sync.Map
	cacheHits      uint64
	cacheMisses    uint64
	mu             sync.RWMutex
	initialized    bool
}

// NewService creates a new RAG service with caching
func NewService(qdrantBinary, qdrantURL string) (*Service, error) {
	var host string
	var port int

	if qdrantURL == "" {
		host = "localhost"
		port = 6333
	} else {
		cleanURL := qdrantURL
		if strings.HasPrefix(qdrantURL, "http://") {
			cleanURL = strings.TrimPrefix(qdrantURL, "http://")
		} else if strings.HasPrefix(qdrantURL, "https://") {
			cleanURL = strings.TrimPrefix(qdrantURL, "https://")
		}

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

	client, err := qdrant.NewClient(&qdrant.Config{
		Host: host,
		Port: port,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client for %s:%d: %v", host, port, err)
	}

	return &Service{
		client:         client,
		qdrantURL:      qdrantURL,
		collections: map[string]string{
			"agent_prompts":    "System prompts for each worker role",
			"coding_standards": "Best practices and coding standards",
			"documentation":    "Technical documentation and guides",
			"code_examples":    "Code examples and patterns",
			"book_expert":      "Technical book content and knowledge",
		},
		pendingReqs:    make(map[string]chan *types.EmbeddingResponse),
		embeddingCache: &sync.Map{},
	}, nil
}

// generateLocalEmbedding generates embeddings with caching
func (s *Service) generateLocalEmbedding(text string) []float32 {
	if len(text) == 0 {
		log.Printf("ERROR: Empty text provided for embedding generation")
		return nil
	}

	// Check cache first
	textHash := hashString(text)
	if cached, exists := s.embeddingCache.Load(textHash); exists {
		return cached.([]float32)
	}

	// Initialize MQTT client if not already done
	if !s.initialized {
		if err := s.initMQTT(context.Background()); err != nil {
			log.Printf("ERROR: Failed to initialize MQTT client: %v", err)
			return nil
		}
	}

	// Create embedding request
	reqID := fmt.Sprintf("emb-%d", time.Now().UnixNano())
	req := types.EmbeddingRequest{
		Text:      text,
		RequestID: reqID,
	}

	payload, err := json.Marshal(req)
	if err != nil {
		log.Printf("ERROR: Failed to marshal embedding request: %v", err)
		return nil
	}

	responseChan := make(chan *types.EmbeddingResponse, 1)

	s.mu.Lock()
	s.pendingReqs[reqID] = responseChan
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.pendingReqs, reqID)
		s.mu.Unlock()
		close(responseChan)
	}()

	publishCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.mqttClient.Publish(publishCtx, types.EmbeddingRequestTopic, payload); err != nil {
		log.Printf("ERROR: Failed to publish embedding request: %v", err)
		return nil
	}

	select {
	case resp := <-responseChan:
		if resp.Error != "" {
			log.Printf("ERROR: Embedding service error: %s", resp.Error)
			return nil
		}

		// Cache the result
		s.embeddingCache.Store(textHash, resp.Embedding)
		return resp.Embedding

	case <-time.After(30 * time.Second):
		log.Printf("ERROR: Embedding request timeout for request %s", reqID)
		return nil
	}
}

// hashString creates a consistent hash for string values
func hashString(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
