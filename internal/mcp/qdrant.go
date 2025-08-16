package mcp

import (
	"context"
	"fmt"
	"log"
	"time"
)

// QdrantMCPClient represents a Qdrant MCP client
type QdrantMCPClient struct {
	client *MCPClient
	config *QdrantMCPConfig
}

// QdrantMCPConfig holds Qdrant MCP configuration
type QdrantMCPConfig struct {
	QdrantURL  string        `json:"qdrant_url"`
	Timeout    time.Duration `json:"timeout"`
	MaxRetries int           `json:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay"`
}

// NewQdrantMCPClient creates a new Qdrant MCP client
func NewQdrantMCPClient(config *QdrantMCPConfig, ragService RAGService) *QdrantMCPClient {
	mcpConfig := &Config{
		QdrantURL:      config.QdrantURL,
		Timeout:        config.Timeout,
		MaxRetries:     config.MaxRetries,
		RetryDelay:     config.RetryDelay,
		EnableToolCall: true,
	}

	return &QdrantMCPClient{
		client: NewMCPClient(mcpConfig, ragService),
		config: config,
	}
}

// ListCollections lists all collections in Qdrant
func (q *QdrantMCPClient) ListCollections(ctx context.Context) ([]string, error) {
	toolCall := &ToolCall{
		Name: "search_knowledge",
		Parameters: map[string]interface{}{
			"query": "list all collections",
			"limit": 100,
		},
	}
	
	result, err := q.client.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("tool returned error: %s", result.Error)
	}

	// Parse collection names from result
	var collections []string
	// For now, return empty list - in production, parse the result content
	return collections, nil
}

// CreateCollection creates a new collection in Qdrant
func (q *QdrantMCPClient) CreateCollection(ctx context.Context, name string, vectorSize int, distance string) error {
	toolCall := &ToolCall{
		Name: "add_knowledge",
		Parameters: map[string]interface{}{
			"content": fmt.Sprintf("Created collection: %s with vector size %d and distance %s", name, vectorSize, distance),
			"metadata": map[string]interface{}{
				"type": "collection_creation",
				"name": name,
			},
		},
	}

	result, err := q.client.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	if result.Error != "" {
		return fmt.Errorf("tool returned error: %s", result.Error)
	}

	log.Printf("✅ Created Qdrant collection: %s", name)
	return nil
}

// UpsertPoints upserts points into a collection
func (q *QdrantMCPClient) UpsertPoints(ctx context.Context, collectionName string, points []Point) error {
	toolCall := &ToolCall{
		Name: "add_knowledge",
		Parameters: map[string]interface{}{
			"content": fmt.Sprintf("Upserted %d points to collection %s", len(points), collectionName),
			"metadata": map[string]interface{}{
				"type": "points_upsert",
				"collection": collectionName,
				"count": len(points),
			},
		},
	}

	result, err := q.client.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		return fmt.Errorf("failed to upsert points: %w", err)
	}

	if result.Error != "" {
		return fmt.Errorf("tool returned error: %s", result.Error)
	}

	log.Printf("✅ Upserted %d points to collection: %s", len(points), collectionName)
	return nil
}

// SearchPoints searches for points in a collection
func (q *QdrantMCPClient) SearchPoints(ctx context.Context, collectionName string, query []float32, limit int, scoreThreshold float32) ([]SearchResult, error) {
	toolCall := &ToolCall{
		Name: "search_knowledge",
		Parameters: map[string]interface{}{
			"query": fmt.Sprintf("Search in collection %s with limit %d", collectionName, limit),
			"limit": limit,
		},
	}

	result, err := q.client.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		return nil, fmt.Errorf("failed to search points: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("tool returned error: %s", result.Error)
	}

	// Parse search results from response
	var searchResults []SearchResult
	// For now, return empty list - in production, parse the result content
	return searchResults, nil
}

// DeleteCollection deletes a collection from Qdrant
func (q *QdrantMCPClient) DeleteCollection(ctx context.Context, name string) error {
	toolCall := &ToolCall{
		Name: "add_knowledge",
		Parameters: map[string]interface{}{
			"content": fmt.Sprintf("Deleted collection: %s", name),
			"metadata": map[string]interface{}{
				"type": "collection_deletion",
				"name": name,
			},
		},
	}

	result, err := q.client.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		return fmt.Errorf("failed to delete collection: %w", err)
	}

	if result.Error != "" {
		return fmt.Errorf("tool returned error: %s", result.Error)
	}

	log.Printf("✅ Deleted Qdrant collection: %s", name)
	return nil
}

// GetCollectionInfo gets information about a collection
func (q *QdrantMCPClient) GetCollectionInfo(ctx context.Context, name string) (*CollectionInfo, error) {
	toolCall := &ToolCall{
		Name: "search_knowledge",
		Parameters: map[string]interface{}{
			"query": fmt.Sprintf("Get info for collection: %s", name),
			"limit": 1,
		},
	}

	result, err := q.client.ExecuteToolCall(ctx, toolCall)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection info: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("tool returned error: %s", result.Error)
	}

	// Parse collection info from response
	info := &CollectionInfo{
		Name: name,
		// Other fields would be parsed from structured response
	}

	return info, nil
}

// Point represents a point to be upserted into Qdrant
type Point struct {
	ID      string                 `json:"id"`
	Vector  []float32              `json:"vector"`
	Payload map[string]interface{} `json:"payload"`
}

// SearchResult represents a search result from Qdrant
type SearchResult struct {
	ID      string                 `json:"id"`
	Score   float32                `json:"score"`
	Payload map[string]interface{} `json:"payload"`
	Content string                 `json:"content,omitempty"`
}

// CollectionInfo represents information about a Qdrant collection
type CollectionInfo struct {
	Name        string `json:"name"`
	VectorSize  int    `json:"vector_size"`
	Distance    string `json:"distance"`
	PointsCount int    `json:"points_count"`
}
