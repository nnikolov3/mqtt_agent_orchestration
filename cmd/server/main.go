package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/niko/mqtt-agent-orchestration/internal/mqtt"
	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// Configuration constants
const (
	DefaultMQTTHost = "localhost"
	DefaultMQTTPort = 1883
	TaskTopic       = "tasks/new"
	ResultTopic     = "tasks/results"
	StatusTopic     = "workers/status"
)

// TestServer represents a simple test server for the MQTT system
type TestServer struct {
	mqttClient *mqtt.Client
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewTestServer creates a new test server
func NewTestServer(mqttHost string, mqttPort int) *TestServer {
	ctx, cancel := context.WithCancel(context.Background())

	mqttClient := mqtt.NewClientWithID(mqttHost, mqttPort, "test-server")

	return &TestServer{
		mqttClient: mqttClient,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// Start starts the test server
func (s *TestServer) Start() error {
	log.Printf("Starting test server")

	// Connect to MQTT broker
	connectCtx, connectCancel := context.WithTimeout(s.ctx, 10*time.Second)
	defer connectCancel()

	if err := s.mqttClient.Connect(connectCtx); err != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", err)
	}

	log.Printf("Connected to MQTT broker")

	// Subscribe to results
	if err := s.mqttClient.Subscribe(s.ctx, ResultTopic, s.handleResult); err != nil {
		return fmt.Errorf("failed to subscribe to results topic: %w", err)
	}

	// Subscribe to worker status updates
	statusPattern := StatusTopic + "/+"
	if err := s.mqttClient.Subscribe(s.ctx, statusPattern, s.handleWorkerStatus); err != nil {
		return fmt.Errorf("failed to subscribe to status topic: %w", err)
	}

	log.Printf("Test server ready")
	return nil
}

// Stop stops the test server
func (s *TestServer) Stop() {
	log.Printf("Stopping test server")
	s.cancel()
	if s.mqttClient != nil {
		s.mqttClient.Disconnect()
	}
}

// PublishTestTask publishes a test task
func (s *TestServer) PublishTestTask(taskType string, payload map[string]string) error {
	task := types.Task{
		ID:        fmt.Sprintf("task-%d", time.Now().UnixNano()),
		Type:      taskType,
		Payload:   payload,
		CreatedAt: time.Now(),
		Priority:  1,
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	ctx, cancel := context.WithTimeout(s.ctx, 5*time.Second)
	defer cancel()

	log.Printf("Publishing task %s of type %s", task.ID, task.Type)
	return s.mqttClient.Publish(ctx, TaskTopic, data)
}

// handleResult handles incoming task results
func (s *TestServer) handleResult(payload []byte) {
	var result types.TaskResult
	if err := json.Unmarshal(payload, &result); err != nil {
		log.Printf("Failed to unmarshal result: %v", err)
		return
	}

	log.Printf("Received result for task %s from worker %s (success: %v, duration: %dms)",
		result.TaskID, result.WorkerID, result.Success, result.Duration)

	if result.Success {
		log.Printf("Result: %s", result.Result)
	} else {
		log.Printf("Error: %s", result.Error)
	}
}

// handleWorkerStatus handles worker status updates
func (s *TestServer) handleWorkerStatus(payload []byte) {
	var status types.WorkerStatus
	if err := json.Unmarshal(payload, &status); err != nil {
		log.Printf("Failed to unmarshal worker status: %v", err)
		return
	}

	log.Printf("Worker %s status: %s (tasks: %d, errors: %d)",
		status.ID, status.Status, status.TasksTotal, status.TasksError)
}

func main() {
	// Parse command line flags
	var (
		mqttHost = flag.String("mqtt-host", DefaultMQTTHost, "MQTT broker host")
		mqttPort = flag.Int("mqtt-port", DefaultMQTTPort, "MQTT broker port")
		taskType = flag.String("task-type", "echo", "Type of test task to send")
		message  = flag.String("message", "Hello from test server!", "Message for test task")
		numTasks = flag.Int("num-tasks", 3, "Number of test tasks to send")
		interval = flag.Duration("interval", 2*time.Second, "Interval between tasks")
	)
	flag.Parse()

	// Create and start server
	server := NewTestServer(*mqttHost, *mqttPort)

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Wait a moment for connections to establish
	time.Sleep(2 * time.Second)

	// Send test tasks
	log.Printf("Sending %d test tasks of type %s", *numTasks, *taskType)

	for i := 0; i < *numTasks; i++ {
		payload := map[string]string{
			"message": fmt.Sprintf("%s (task %d)", *message, i+1),
		}

		if *taskType == "uppercase" {
			payload = map[string]string{
				"text": fmt.Sprintf("%s (task %d)", *message, i+1),
			}
		}

		if err := server.PublishTestTask(*taskType, payload); err != nil {
			log.Printf("Failed to publish task %d: %v", i+1, err)
		}

		if i < *numTasks-1 {
			time.Sleep(*interval)
		}
	}

	// Keep server running to receive results
	log.Printf("Waiting for results...")
	time.Sleep(10 * time.Second)

	log.Printf("Test completed")
}
