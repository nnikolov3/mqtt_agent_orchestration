package mcp

import (
	"encoding/json"
)

// Request represents an MCP request
type Request struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      string      `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// Response represents an MCP response
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error represents an MCP error
type Error struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// InitializeParams represents initialization parameters
type InitializeParams struct {
	ProtocolVersion string              `json:"protocolVersion"`
	Capabilities    ClientCapabilities  `json:"capabilities"`
	ClientInfo      *ClientInfo         `json:"clientInfo,omitempty"`
}

// ClientCapabilities represents client capabilities
type ClientCapabilities struct {
	Tools     ToolCapabilities     `json:"tools,omitempty"`
	Resources ResourceCapabilities `json:"resources,omitempty"`
}

// ToolCapabilities represents tool capabilities
type ToolCapabilities struct {
	Enabled bool `json:"enabled"`
}

// ResourceCapabilities represents resource capabilities
type ResourceCapabilities struct {
	Enabled bool `json:"enabled"`
}

// ClientInfo represents client information
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// ToolCallParams represents tool call parameters
type ToolCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult represents a tool execution result
type ToolResult struct {
	Content []ToolResultContent `json:"content"`
	IsError bool                `json:"isError"`
}

// ToolResultContent represents content in a tool result
type ToolResultContent struct {
	Type    string                 `json:"type"`
	Text    string                 `json:"text,omitempty"`
	Image   *ImageContent          `json:"image,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// ImageContent represents image content
type ImageContent struct {
	URI       string `json:"uri"`
	AltText   string `json:"altText,omitempty"`
	MimeType  string `json:"mimeType,omitempty"`
}

// Resource represents an MCP resource
type Resource struct {
	URI         string                 `json:"uri"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	MimeType    string                 `json:"mimeType"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
}

// ResourceReadParams represents resource read parameters
type ResourceReadParams struct {
	URI string `json:"uri"`
}

// ResourceContent represents resource content
type ResourceContent struct {
	URI       string                 `json:"uri"`
	MimeType  string                 `json:"mimeType"`
	Text      string                 `json:"text,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	ExpiresAt *string                `json:"expiresAt,omitempty"`
}
