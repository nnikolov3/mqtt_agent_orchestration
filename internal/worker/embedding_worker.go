package worker

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/nnikolov3/mqtt-agent-orchestration/internal/mqtt"
	"github.com/nnikolov3/mqtt-agent-orchestration/pkg/types"
)

// EmbeddingWorker handles embedding requests via MQTT
type EmbeddingWorker struct {
	mqttClient *mqtt.Client
	modelPath  string
}

// NewEmbeddingWorker creates a new embedding worker
func NewEmbeddingWorker(mqttHost string, mqttPort int, modelPath string) *EmbeddingWorker {
	return &EmbeddingWorker{
		mqttClient: mqtt.NewClientWithID(mqttHost, mqttPort, "embedding-worker"),
		modelPath:  modelPath,
	}
}

// Start initializes and starts the embedding worker
func (w *EmbeddingWorker) Start(ctx context.Context) error {
	if err := w.mqttClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect to MQTT: %w", err)
	}

	return w.mqttClient.Subscribe(ctx, types.EmbeddingRequestTopic, w.handleRequest)
}

// Add sanitization function to prevent command injection
func sanitizeInput(text string) string {
	// Use base64 encoding to safely pass text content
	// This prevents shell injection attacks completely
	return base64.StdEncoding.EncodeToString([]byte(text))
}

// Update the handleRequest function to use sanitized input
func (w *EmbeddingWorker) handleRequest(payload []byte) {
	var req types.EmbeddingRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		// Fail loud per design principles - log error and publish error response
		log.Printf("ERROR: Invalid embedding request payload: %v", err)
		w.publishErrorResponse("", fmt.Sprintf("Invalid embedding request: %v", err))
		return
	}

	// Input validation - prevent DoS attacks
	if len(req.Text) > 64*1024 { // 64KB limit
		log.Printf("ERROR: Embedding request text too large: %d bytes", len(req.Text))
		w.publishErrorResponse(req.RequestID, "Text field exceeds maximum allowed size (64KB)")
		return
	}

	// Execute binary from 'bin' directory with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Sanitize input to prevent command injection
	safeText := sanitizeInput(req.Text)

	cmd := exec.CommandContext(ctx, "./bin/llama-embedding",
		"-m", w.modelPath,
		"-p", safeText,
		"--embd-output-format", "json",
		"--embd-normalize", "2",
		"--base64-input") // Add flag to indicate base64 encoded input

	output, err := cmd.Output()
	resp := types.EmbeddingResponse{RequestID: req.RequestID}

	if err != nil {
		// Check if context was cancelled (timeout)
		if ctx.Err() == context.DeadlineExceeded {
			resp.Error = "Embedding generation timed out after 30 seconds"
			log.Printf("ERROR: Embedding timeout for request %s", req.RequestID)
		} else {
			resp.Error = fmt.Sprintf("Embedding binary error: %v", err)
			log.Printf("ERROR: Embedding binary failed for request %s: %v", req.RequestID, err)
		}
	} else {
		// Parse output
		var data struct {
			Data []struct {
				Embedding []float32 `json:"embedding"`
			} `json:"data"`
		}

		if err := json.Unmarshal(output, &data); err != nil {
			resp.Error = fmt.Sprintf("Failed to parse embedding output: %v", err)
			log.Printf("ERROR: Failed to parse embedding output for request %s: %v", req.RequestID, err)
		} else if len(data.Data) > 0 {
			resp.Embedding = data.Data[0].Embedding
			log.Printf("SUCCESS: Generated embedding for request %s (%d dimensions)", req.RequestID, len(resp.Embedding))
		} else {
			resp.Error = "No embedding data in response"
			log.Printf("ERROR: No embedding data returned for request %s", req.RequestID)
		}
	}

	// Publish response with explicit error handling
	if err := w.publishResponse(resp); err != nil {
		log.Printf("ERROR: Failed to publish embedding response for request %s: %v", req.RequestID, err)
	}
}

// publishResponse publishes an embedding response to MQTT
// Following "Explicit error handling" - never ignore errors
func (w *EmbeddingWorker) publishResponse(resp types.EmbeddingResponse) error {
	payloadOut, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return w.mqttClient.Publish(ctx, types.EmbeddingResponseTopic, payloadOut)
}

// publishErrorResponse publishes an error response when we can't parse the original request
// Following "Graceful Degradation" - continue operating even with bad input
func (w *EmbeddingWorker) publishErrorResponse(requestID, errorMsg string) {
	resp := types.EmbeddingResponse{
		RequestID: requestID,
		Error:     errorMsg,
	}

	if err := w.publishResponse(resp); err != nil {
		log.Printf("ERROR: Failed to publish error response: %v", err)
	}
}

// Stop gracefully shuts down the embedding worker
func (w *EmbeddingWorker) Stop(ctx context.Context) error {
	w.mqttClient.Disconnect()
	return nil
}
