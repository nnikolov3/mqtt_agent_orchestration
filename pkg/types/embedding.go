package types

// EmbeddingRequest represents a request for text embedding
type EmbeddingRequest struct {
	Text      string `json:"text"`
	RequestID string `json:"request_id"`
}

// EmbeddingResponse represents the response from embedding generation
type EmbeddingResponse struct {
	RequestID string    `json:"request_id"`
	Embedding []float32 `json:"embedding"`
	Error     string    `json:"error,omitempty"`
}
