package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/niko/mqtt-agent-orchestration/internal/localmodels"
	"github.com/niko/mqtt-agent-orchestration/internal/mqtt"
	"github.com/niko/mqtt-agent-orchestration/internal/rag"
	"github.com/niko/mqtt-agent-orchestration/internal/worker"
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// Configuration constants
const (
	DefaultMQTTHost      = "localhost"
	DefaultMQTTPort      = 1883
	DefaultQdrantURL     = "http://localhost:6333"
	StatusUpdateInterval = 30 * time.Second
	TaskTimeout          = 10 * time.Minute
)

// RoleWorkerApp represents a role-specific worker application
type RoleWorkerApp struct {
	workerID   string
	role       types.WorkerRole
	mqttClient *mqtt.Client
	processor  *worker.RoleBasedProcessor
	ragService *rag.Service
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewRoleWorkerApp creates a new role-specific worker
func NewRoleWorkerApp(workerID string, role types.WorkerRole, mqttHost string, mqttPort int, qdrantURL string) (*RoleWorkerApp, error) {
	ctx, cancel := context.WithCancel(context.Background())

	clientID := fmt.Sprintf("%s-%s", role, workerID)
	mqttClient := mqtt.NewClientWithID(mqttHost, mqttPort, clientID)

	// Create RAG service - fail fast if unavailable
	ragService, err := rag.NewService("qdrant", qdrantURL)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create RAG service: %v", err)
	}

	// Load model configurations
	modelConfigs := make(map[string]localmodels.ModelConfig)
	// For now, create a simple config - in production, load from configs/models.yaml
	modelConfigs["qwen-text"] = localmodels.ModelConfig{
		Name:        "Qwen2.5-Omni-3B",
		BinaryPath:  "/home/niko/bin/llama-server",
		ModelPath:   "/data/models/Qwen2.5-Omni-3B-Q8_0.gguf",
		Type:        localmodels.ModelTypeText,
		MemoryLimit: 5500,
	}

	// Create local models manager
	modelManager, err := localmodels.NewManager(localmodels.ModelManagerConfig{
		MaxGPUMemory:    5632, // 5.5GB for RTX 3060
		NvidiaSMIPath:   "/usr/bin/nvidia-smi",
		MonitorInterval: 30 * time.Second,
		Models:          modelConfigs,
	})
	if err != nil {
		log.Printf("Warning: Failed to create model manager: %v", err)
		modelManager = nil
	}

	// Create content analyzer
	contentAnalyzer := worker.NewContentAnalyzer(modelConfigs)

	// Create role-based processor
	processor := worker.NewRoleBasedProcessor(role, ragService, modelManager, contentAnalyzer)

	return &RoleWorkerApp{
		workerID:   workerID,
		role:       role,
		mqttClient: mqttClient,
		processor:  processor,
		ragService: ragService,
		ctx:        ctx,
		cancel:     cancel,
	}, nil
}

// Start starts the role worker
func (app *RoleWorkerApp) Start() error {
	log.Printf("Starting %s worker %s", app.role, app.workerID)

	// Connect to MQTT
	connectCtx, connectCancel := context.WithTimeout(app.ctx, 10*time.Second)
	defer connectCancel()

	if err := app.mqttClient.Connect(connectCtx); err != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", err)
	}

	log.Printf("Connected to MQTT broker")

	// Subscribe to role-specific task topic
	taskTopic := fmt.Sprintf("tasks/workflow/%s", app.getStageForRole())
	if err := app.mqttClient.Subscribe(app.ctx, taskTopic, app.handleTask); err != nil {
		return fmt.Errorf("failed to subscribe to %s: %w", taskTopic, err)
	}

	log.Printf("Subscribed to task topic: %s", taskTopic)

	// Start status updates
	go app.publishStatusPeriodically()

	// Check RAG availability
	if app.ragService.IsAvailable(app.ctx) {
		log.Printf("RAG service (qdrant) is available")
	} else {
		log.Printf("RAG service (qdrant) is not available, using fallback knowledge")
	}

	log.Printf("%s worker %s is ready", app.role, app.workerID)
	return nil
}

// Stop stops the worker
func (app *RoleWorkerApp) Stop() {
	log.Printf("Stopping %s worker %s", app.role, app.workerID)
	app.cancel()
	if app.mqttClient != nil {
		app.mqttClient.Disconnect()
	}
}

// handleTask processes incoming workflow tasks
func (app *RoleWorkerApp) handleTask(payload []byte) {
	var workflowTask types.WorkflowTask
	if err := json.Unmarshal(payload, &workflowTask); err != nil {
		log.Printf("Failed to unmarshal workflow task: %v", err)
		return
	}

	// Check if this task is for our role
	if workflowTask.RequiredRole != app.role {
		log.Printf("Ignoring task %s - requires role %s, we are %s",
			workflowTask.ID, workflowTask.RequiredRole, app.role)
		return
	}

	log.Printf("Processing workflow task %s (stage: %s, workflow: %s)",
		workflowTask.ID, workflowTask.Stage, workflowTask.WorkflowID)

	// Process task with timeout
	taskCtx, taskCancel := context.WithTimeout(app.ctx, TaskTimeout)
	defer taskCancel()

	// Process workflow task with role-based processor
	result, err := app.processor.ProcessWorkflowTask(taskCtx, &workflowTask)

	// Create workflow result
	workflowResult := types.WorkflowResult{
		TaskResult: types.TaskResult{
			TaskID:      workflowTask.ID,
			WorkerID:    app.workerID,
			ProcessedAt: time.Now(),
			Duration:    0, // Will be calculated
		},
		WorkflowID: workflowTask.WorkflowID,
		Stage:      workflowTask.Stage,
		WorkerRole: app.role,
	}

	if err != nil {
		workflowResult.Success = false
		workflowResult.Error = err.Error()
		log.Printf("Task %s failed: %v", workflowTask.ID, err)
	} else {
		workflowResult.Success = true
		workflowResult.Result = result

		// Check for approval/rejection patterns
		resultUpper := strings.ToUpper(result)
		if strings.Contains(resultUpper, "APPROVED") {
			workflowResult.Approved = true
		} else if strings.Contains(resultUpper, "REJECTED") {
			workflowResult.RequiresRetry = true
			// Extract feedback
			lines := strings.Split(result, "\n")
			for _, line := range lines {
				if strings.Contains(strings.ToUpper(line), "REJECTED:") {
					workflowResult.ReviewFeedback = strings.TrimSpace(strings.TrimPrefix(line, "REJECTED:"))
					break
				}
			}
		}

		log.Printf("Task %s completed successfully", workflowTask.ID)
	}

	// Publish result
	if err := app.publishResult(workflowResult); err != nil {
		log.Printf("Failed to publish result for task %s: %v", workflowTask.ID, err)
	}
}

// publishResult publishes workflow result
func (app *RoleWorkerApp) publishResult(result types.WorkflowResult) error {
	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	ctx, cancel := context.WithTimeout(app.ctx, 5*time.Second)
	defer cancel()

	topic := fmt.Sprintf("results/workflow/%s", result.Stage)
	return app.mqttClient.Publish(ctx, topic, data)
}

// publishStatusPeriodically publishes status updates
func (app *RoleWorkerApp) publishStatusPeriodically() {
	ticker := time.NewTicker(StatusUpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			status := types.ExtendedWorkerStatus{
				WorkerStatus: types.WorkerStatus{
					ID:       app.workerID,
					Status:   "idle",
					LastSeen: time.Now(),
				},
				Role:         app.role,
				Capabilities: worker.GetCapabilitiesForRole(app.role),
			}

			data, err := json.Marshal(status)
			if err != nil {
				log.Printf("Failed to marshal status: %v", err)
				continue
			}

			ctx, cancel := context.WithTimeout(app.ctx, 5*time.Second)
			topic := fmt.Sprintf("workers/status/%s/%s", app.role, app.workerID)
			if err := app.mqttClient.Publish(ctx, topic, data); err != nil {
				log.Printf("Failed to publish status: %v", err)
			}
			cancel()

		case <-app.ctx.Done():
			return
		}
	}
}

// getStageForRole maps roles to workflow stages
func (app *RoleWorkerApp) getStageForRole() types.WorkflowStage {
	switch app.role {
	case types.RoleDeveloper:
		return types.StageDevelopment
	case types.RoleReviewer:
		return types.StageReview
	case types.RoleApprover:
		return types.StageApproval
	case types.RoleTester:
		return types.StageTesting
	default:
		return types.StageDevelopment
	}
}

func main() {
	// Parse command line flags
	var (
		workerID  = flag.String("id", "worker-1", "Worker ID")
		role      = flag.String("role", "developer", "Worker role (developer, reviewer, approver, tester)")
		mqttHost  = flag.String("mqtt-host", DefaultMQTTHost, "MQTT broker host")
		mqttPort  = flag.Int("mqtt-port", DefaultMQTTPort, "MQTT broker port")
		qdrantURL = flag.String("qdrant-url", DefaultQdrantURL, "Qdrant URL for RAG")
		verbose   = flag.Bool("verbose", false, "Enable verbose logging")
	)
	flag.Parse()

	// Configure logging
	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Parse role
	var workerRole types.WorkerRole
	switch *role {
	case "developer":
		workerRole = types.RoleDeveloper
	case "reviewer":
		workerRole = types.RoleReviewer
	case "approver":
		workerRole = types.RoleApprover
	case "tester":
		workerRole = types.RoleTester
	default:
		log.Fatalf("Invalid role: %s. Must be one of: developer, reviewer, approver, tester", *role)
	}

	// Create worker application
	app, err := NewRoleWorkerApp(*workerID, workerRole, *mqttHost, *mqttPort, *qdrantURL)
	if err != nil {
		log.Fatalf("Failed to create worker application: %v", err)
	}

	// Set up signal handling
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
