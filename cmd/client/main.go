package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/niko/mqtt-agent-orchestration/internal/mqtt"
)

// Configuration constants
const (
	DefaultMQTTHost = "localhost"
	DefaultMQTTPort = 1883
)

// WorkflowClient provides a standalone interface to trigger workflows
type WorkflowClient struct {
	mqttClient *mqtt.Client
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWorkflowClient creates a new workflow client
func NewWorkflowClient(mqttHost string, mqttPort int) *WorkflowClient {
	ctx, cancel := context.WithCancel(context.Background())
	mqttClient := mqtt.NewClientWithID(mqttHost, mqttPort, "workflow-client")

	return &WorkflowClient{
		mqttClient: mqttClient,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start connects the client to MQTT
func (c *WorkflowClient) Start() error {
	connectCtx, connectCancel := context.WithTimeout(c.ctx, 10*time.Second)
	defer connectCancel()

	return c.mqttClient.Connect(connectCtx)
}

// Stop disconnects the client
func (c *WorkflowClient) Stop() {
	c.cancel()
	if c.mqttClient != nil {
		c.mqttClient.Disconnect()
	}
}

// CreateDocument triggers an autonomous document creation workflow
func (c *WorkflowClient) CreateDocument(docType, outputFile string) error {
	request := map[string]interface{}{
		"type": "create_document",
		"payload": map[string]string{
			"document_type": docType,
			"output_file":   outputFile,
		},
	}

	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(c.ctx, 5*time.Second)
	defer cancel()

	log.Printf("Triggering %s document creation workflow", docType)
	log.Printf("Output will be written to: %s", outputFile)
	log.Printf("This will go through: Developer → Reviewer → Approver → Tester → Final Output")

	return c.mqttClient.Publish(ctx, "orchestrator/workflow", data)
}

// ListAvailableDocuments returns available document types
func (c *WorkflowClient) ListAvailableDocuments() []string {
	return []string{
		"go_coding_standards",
		"python_coding_standards",
		"bash_coding_standards",
		"project_documentation",
		"api_documentation",
		// Add more document types here as they are implemented
	}
}

// ListAvailableModels returns available local models
func (c *WorkflowClient) ListAvailableModels() ([]string, error) {
	// This would query the model manager for available models
	// For now, return a static list based on configuration
	return []string{
		"qwen-omni",
		"qwen-vl",
		"llava",
		"mimo",
		"qwen-embedding-4b",
	}, nil
}

func main() {
	// Parse command line flags
	var (
		mqttHost    = flag.String("mqtt-host", DefaultMQTTHost, "MQTT broker host")
		mqttPort    = flag.Int("mqtt-port", DefaultMQTTPort, "MQTT broker port")
		docType     = flag.String("doc-type", "", "Document type to create")
		outputFile  = flag.String("output", "", "Output file path")
		list        = flag.Bool("list", false, "List available document types")
		listModels  = flag.Bool("list-models", false, "List available local models")
		preferLocal = flag.Bool("prefer-local", false, "Prefer local models over external AI helpers")
		modelType   = flag.String("model-type", "", "Specify model type for task")
		verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Configure logging
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Create client
	client := NewWorkflowClient(*mqttHost, *mqttPort)
	defer client.Stop()

	// Connect
	if err := client.Start(); err != nil {
		log.Fatalf("Failed to connect to MQTT broker: %v", err)
	}

	// Handle list commands
	if *list {
		fmt.Println("Available document types:")
		for _, docType := range client.ListAvailableDocuments() {
			fmt.Printf("  - %s\n", docType)
		}
		return
	}

	if *listModels {
		models, err := client.ListAvailableModels()
		if err != nil {
			log.Fatalf("Failed to list models: %v", err)
		}
		fmt.Println("Available local models:")
		for _, model := range models {
			fmt.Printf("  - %s\n", model)
		}
		return
	}

	// Validate required flags
	if *docType == "" {
		log.Fatal("Please specify --doc-type. Use --list to see available types.")
	}

	if *outputFile == "" {
		log.Fatal("Please specify --output file path")
	}

	// Check if document type is supported
	supported := false
	for _, availableType := range client.ListAvailableDocuments() {
		if availableType == *docType {
			supported = true
			break
		}
	}

	if !supported {
		log.Fatalf("Document type '%s' is not supported. Use --list to see available types.", *docType)
	}

	// Trigger workflow with model preferences
	if *preferLocal {
		log.Printf("Preferring local models over external AI helpers")
	}
	if *modelType != "" {
		log.Printf("Using specified model type: %s", *modelType)
	}

	if err := client.CreateDocument(*docType, *outputFile); err != nil {
		log.Fatalf("Failed to trigger workflow: %v", err)
	}

	log.Printf("Workflow triggered successfully!")
	log.Printf("")
	log.Printf("The autonomous workflow will now:")
	log.Printf("1. Developer worker: Create initial %s document", *docType)
	log.Printf("2. Reviewer worker: Review and improve the document")
	log.Printf("3. Approver worker: Perform final approval check")
	log.Printf("4. Tester worker: Validate document structure and content")
	log.Printf("5. Orchestrator: Write final approved version to %s", *outputFile)
	log.Printf("")
	log.Printf("Monitor the orchestrator and worker logs to see progress.")
	log.Printf("The workflow will handle retries automatically if any stage fails.")
}
