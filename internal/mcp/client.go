package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// RAGService interface to avoid import cycle
type RAGService interface {
	SearchKnowledge(ctx context.Context, query types.RAGQuery) (*types.RAGResponse, error)
	AddDocument(ctx context.Context, content string, metadata map[string]interface{}) error
}

// MCPClient provides integration between local models and external tools
type MCPClient struct {
	ragService RAGService
	config     *Config
}

// Config holds MCP client configuration
type Config struct {
	QdrantURL      string        `yaml:"qdrant_url"`
	Timeout        time.Duration `yaml:"timeout"`
	MaxRetries     int           `yaml:"max_retries"`
	RetryDelay     time.Duration `yaml:"retry_delay"`
	EnableToolCall bool          `yaml:"enable_tool_call"`
}

// ToolCall represents a tool call request
type ToolCall struct {
	Name       string                 `json:"name"`
	Parameters map[string]interface{} `json:"parameters"`
}

// ToolResponse represents a tool call response
type ToolResponse struct {
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Duration  time.Duration          `json:"duration"`
}

// NewMCPClient creates a new MCP client
func NewMCPClient(config *Config, ragService RAGService) *MCPClient {
	return &MCPClient{
		ragService: ragService,
		config:     config,
	}
}

// SearchKnowledge searches the RAG database for relevant information
func (c *MCPClient) SearchKnowledge(ctx context.Context, query string, limit int) (*ToolResponse, error) {
	start := time.Now()
	
	ragQuery := types.RAGQuery{
		Query: query,
		TopK:  limit,
	}
	
	results, err := c.ragService.SearchKnowledge(ctx, ragQuery)
	if err != nil {
		return &ToolResponse{
			Error:    fmt.Sprintf("Search failed: %v", err),
			Duration: time.Since(start),
		}, err
	}
	
	content, err := json.Marshal(results)
	if err != nil {
		return &ToolResponse{
			Error:    fmt.Sprintf("Failed to marshal results: %v", err),
			Duration: time.Since(start),
		}, err
	}
	
	return &ToolResponse{
		Content:  string(content),
		Duration: time.Since(start),
	}, nil
}

// AddKnowledge adds new content to the RAG database
func (c *MCPClient) AddKnowledge(ctx context.Context, content string, metadata map[string]interface{}) (*ToolResponse, error) {
	start := time.Now()
	
	err := c.ragService.AddDocument(ctx, content, metadata)
	if err != nil {
		return &ToolResponse{
			Error:    fmt.Sprintf("Failed to add knowledge: %v", err),
			Duration: time.Since(start),
		}, err
	}
	
	return &ToolResponse{
		Content:  "Knowledge added successfully",
		Duration: time.Since(start),
	}, nil
}

// GetContext retrieves relevant context for a task
func (c *MCPClient) GetContext(ctx context.Context, task string, contextType string) (*ToolResponse, error) {
	start := time.Now()
	
	// Build context-specific query
	var query string
	switch contextType {
	case "coding_standards":
		query = fmt.Sprintf("coding standards guidelines for: %s", task)
	case "documentation":
		query = fmt.Sprintf("documentation examples for: %s", task)
	case "git_changes":
		query = fmt.Sprintf("recent changes related to: %s", task)
	default:
		query = task
	}
	
	ragQuery := types.RAGQuery{
		Query: query,
		TopK:  3,
	}
	results, err := c.ragService.SearchKnowledge(ctx, ragQuery)
	if err != nil {
		return &ToolResponse{
			Error:    fmt.Sprintf("Context retrieval failed: %v", err),
			Duration: time.Since(start),
		}, err
	}
	
	content, err := json.Marshal(results)
	if err != nil {
		return &ToolResponse{
			Error:    fmt.Sprintf("Failed to marshal context: %v", err),
			Duration: time.Since(start),
		}, err
	}
	
	return &ToolResponse{
		Content:  string(content),
		Duration: time.Since(start),
	}, nil
}

// ExecuteToolCall executes a tool call with retry logic
func (c *MCPClient) ExecuteToolCall(ctx context.Context, toolCall *ToolCall) (*ToolResponse, error) {
	var lastErr error
	
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(c.config.RetryDelay)
		}
		
		response, err := c.executeToolCallOnce(ctx, toolCall)
		if err == nil {
			return response, nil
		}
		
		lastErr = err
	}
	
	return &ToolResponse{
		Error: fmt.Sprintf("Tool call failed after %d attempts: %v", c.config.MaxRetries+1, lastErr),
	}, lastErr
}

// executeToolCallOnce executes a single tool call
func (c *MCPClient) executeToolCallOnce(ctx context.Context, toolCall *ToolCall) (*ToolResponse, error) {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()
	
	switch toolCall.Name {
	case "search_knowledge":
		query, _ := toolCall.Parameters["query"].(string)
		limit, _ := toolCall.Parameters["limit"].(int)
		if limit == 0 {
			limit = 5
		}
		return c.SearchKnowledge(timeoutCtx, query, limit)
		
	case "add_knowledge":
		content, _ := toolCall.Parameters["content"].(string)
		metadata, _ := toolCall.Parameters["metadata"].(map[string]interface{})
		return c.AddKnowledge(timeoutCtx, content, metadata)
		
	case "get_context":
		task, _ := toolCall.Parameters["task"].(string)
		contextType, _ := toolCall.Parameters["context_type"].(string)
		return c.GetContext(timeoutCtx, task, contextType)
		
	default:
		return &ToolResponse{
			Error: fmt.Sprintf("Unknown tool: %s", toolCall.Name),
		}, fmt.Errorf("unknown tool: %s", toolCall.Name)
	}
}

// GetAvailableTools returns the list of available tools
func (c *MCPClient) GetAvailableTools() []string {
	return []string{
		"search_knowledge",
		"add_knowledge", 
		"get_context",
	}
}

// IsToolAvailable checks if a tool is available
func (c *MCPClient) IsToolAvailable(toolName string) bool {
	available := c.GetAvailableTools()
	for _, tool := range available {
		if tool == toolName {
			return true
		}
	}
	return false
}
