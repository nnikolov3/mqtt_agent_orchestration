package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/niko/mqtt-agent-orchestration/internal/mqtt"
	"github.com/niko/mqtt-agent-orchestration/internal/worker"
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// Configuration constants - following "Never hard code values" principle
const (
	DefaultMQTTHost      = "localhost"
	DefaultMQTTPort      = 1883
	DefaultWorkerID      = "worker-1"
	TaskTopic            = "tasks/new"
	ResultTopic          = "tasks/results"
	StatusTopic          = "workers/status"
	StatusUpdateInterval = 30 * time.Second
	TaskTimeout          = 5 * time.Minute
)

// SimpleTaskProcessor implements basic task processing for testing
type SimpleTaskProcessor struct{}

// ProcessTask processes tasks based on their type
func (p *SimpleTaskProcessor) ProcessTask(ctx context.Context, task types.Task) (string, error) {
	switch task.Type {
	case "echo":
		if message, ok := task.Payload["message"]; ok {
			return fmt.Sprintf("Echo: %s", message), nil
		}
		return "Echo: (no message)", nil

	case "uppercase":
		if text, ok := task.Payload["text"]; ok {
			return fmt.Sprintf("UPPERCASE: %s", text), nil
		}
		return "UPPERCASE: (no text)", nil

	case "ai_helper":
		return p.processAIHelperTask(ctx, task)

	case "create_document":
		return p.processCreateDocumentTask(ctx, task)

	default:
		return "", fmt.Errorf("unknown task type: %s", task.Type)
	}
}

// processAIHelperTask processes AI helper tasks by calling external AI helper scripts
func (p *SimpleTaskProcessor) processAIHelperTask(ctx context.Context, task types.Task) (string, error) {
	helperName, ok := task.Payload["helper"]
	if !ok {
		return "", fmt.Errorf("missing 'helper' in task payload")
	}

	prompt, ok := task.Payload["prompt"]
	if !ok {
		return "", fmt.Errorf("missing 'prompt' in task payload")
	}

	// Execute AI helper command
	cmd := exec.CommandContext(ctx, helperName, prompt)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("AI helper %s failed: %w", helperName, err)
	}

	return string(output), nil
}

// processCreateDocumentTask creates documents using AI helpers
func (p *SimpleTaskProcessor) processCreateDocumentTask(ctx context.Context, task types.Task) (string, error) {
	documentType, ok := task.Payload["document_type"]
	if !ok {
		return "", fmt.Errorf("missing 'document_type' in task payload")
	}

	outputFile, ok := task.Payload["output_file"]
	if !ok {
		return "", fmt.Errorf("missing 'output_file' in task payload")
	}

	switch documentType {
	case "go_coding_standards":
		return p.createGoCodingStandards(ctx, outputFile, task.Payload)
	default:
		return "", fmt.Errorf("unknown document type: %s", documentType)
	}
}

// createGoCodingStandards creates Go coding standards document using AI helper
func (p *SimpleTaskProcessor) createGoCodingStandards(ctx context.Context, outputFile string, payload map[string]string) (string, error) {
	// Read the existing bash coding standards as reference
	bashStandardsFile := "/home/niko/.claude/BASH_CODING_STANDARD_CLAUDE.md"

	prompt := fmt.Sprintf(`Create comprehensive Go coding standards document similar to the BASH coding standards at %s. Include:

1. Core Principles section covering Go-specific best practices
2. Package and Module Management (naming, organization, imports)
3. Variable and Constant Declaration (naming conventions, scoping, zero values)
4. Function and Method Design (naming, parameters, returns, receivers)
5. Error Handling Patterns (explicit error handling, wrapping, checking)
6. Struct and Interface Design (composition, embedding, interface segregation)
7. Concurrency Patterns (goroutines, channels, mutexes, context)
8. Testing Standards (table-driven tests, mocking, benchmarks)
9. Code Organization (directory structure, file naming)
10. Performance Guidelines (memory allocation, string handling, profiling)
11. Documentation Standards (godoc, comments, examples)
12. Security Best Practices (input validation, secrets handling)
13. Compliance Checklist similar to bash standards

Follow these Go-specific principles:
- Explicit is better than implicit
- Composition over inheritance  
- Interfaces should be small and focused
- Handle errors explicitly, don't ignore them
- Use gofmt, go vet, golangci-lint
- Write idiomatic Go code
- Prefer clarity over cleverness
- Zero values should be useful

Format as markdown with clear examples of good and bad patterns. Make it comprehensive but practical.`, bashStandardsFile)

	// Use Gemini for comprehensive analysis (best for documentation)
	cmd := exec.CommandContext(ctx, "gemini_code_analyzer", prompt)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to generate Go coding standards: %w", err)
	}

	// Write to output file
	err = os.WriteFile(outputFile, output, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write output file %s: %w", outputFile, err)
	}

	return fmt.Sprintf("Successfully created Go coding standards document: %s (%d bytes)", outputFile, len(output)), nil
}

// WorkerApp represents the main worker application
type WorkerApp struct {
	workerID   string
	mqttClient *mqtt.Client
	worker     *worker.Worker
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWorkerApp creates a new worker application
func NewWorkerApp(workerID, mqttHost string, mqttPort int) *WorkerApp {
	ctx, cancel := context.WithCancel(context.Background())

	mqttClient := mqtt.NewClientWithID(mqttHost, mqttPort, fmt.Sprintf("worker-%s", workerID))
	processor := &SimpleTaskProcessor{}
	w := worker.NewWorker(workerID, processor)

	return &WorkerApp{
		workerID:   workerID,
		mqttClient: mqttClient,
		worker:     w,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start starts the worker application
func (app *WorkerApp) Start() error {
	log.Printf("Starting worker %s", app.workerID)

	// Connect to MQTT broker
	connectCtx, connectCancel := context.WithTimeout(app.ctx, 10*time.Second)
	defer connectCancel()

	if err := app.mqttClient.Connect(connectCtx); err != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", err)
	}

	log.Printf("Connected to MQTT broker")

	// Subscribe to task topic
	if err := app.mqttClient.Subscribe(app.ctx, TaskTopic, app.handleTask); err != nil {
		return fmt.Errorf("failed to subscribe to task topic: %w", err)
	}

	log.Printf("Subscribed to task topic: %s", TaskTopic)

	// Start status update goroutine
	go app.publishStatusPeriodically()

	// Publish initial status
	app.publishStatus()

	log.Printf("Worker %s is ready and waiting for tasks", app.workerID)

	return nil
}

// Stop stops the worker application
func (app *WorkerApp) Stop() {
	log.Printf("Stopping worker %s", app.workerID)

	app.cancel()

	if app.mqttClient != nil {
		app.mqttClient.Disconnect()
	}

	log.Printf("Worker %s stopped", app.workerID)
}

// handleTask processes incoming task messages
func (app *WorkerApp) handleTask(payload []byte) {
	var task types.Task
	if err := json.Unmarshal(payload, &task); err != nil {
		log.Printf("Failed to unmarshal task: %v", err)
		return
	}

	log.Printf("Received task %s of type %s", task.ID, task.Type)

	// Process task with timeout
	taskCtx, taskCancel := context.WithTimeout(app.ctx, TaskTimeout)
	defer taskCancel()

	result := app.worker.ProcessTask(taskCtx, task)

	// Publish result
	if err := app.publishResult(result); err != nil {
		log.Printf("Failed to publish result for task %s: %v", task.ID, err)
	} else {
		log.Printf("Published result for task %s (success: %v)", task.ID, result.Success)
	}
}

// publishResult publishes a task result to the results topic
func (app *WorkerApp) publishResult(result types.TaskResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	ctx, cancel := context.WithTimeout(app.ctx, 5*time.Second)
	defer cancel()

	return app.mqttClient.Publish(ctx, ResultTopic, data)
}

// publishStatus publishes worker status
func (app *WorkerApp) publishStatus() error {
	status := app.worker.GetStatus()

	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	ctx, cancel := context.WithTimeout(app.ctx, 5*time.Second)
	defer cancel()

	topic := fmt.Sprintf("%s/%s", StatusTopic, app.workerID)
	return app.mqttClient.Publish(ctx, topic, data)
}

// publishStatusPeriodically publishes status updates at regular intervals
func (app *WorkerApp) publishStatusPeriodically() {
	ticker := time.NewTicker(StatusUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := app.publishStatus(); err != nil {
				log.Printf("Failed to publish status: %v", err)
			}
		case <-app.ctx.Done():
			return
		}
	}
}

// main function following the design principles
func main() {
	// Parse command line flags - explicit configuration
	var (
		workerID = flag.String("id", DefaultWorkerID, "Worker ID")
		mqttHost = flag.String("mqtt-host", DefaultMQTTHost, "MQTT broker host")
		mqttPort = flag.Int("mqtt-port", DefaultMQTTPort, "MQTT broker port")
		verbose  = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Configure logging
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Create worker application
	app := NewWorkerApp(*workerID, *mqttHost, *mqttPort)

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start worker
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start worker: %v", err)
	}

	// Wait for signal
	<-sigChan

	// Graceful shutdown
	app.Stop()
}
