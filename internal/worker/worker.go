package worker

import (
	"context"
	"sync"
	"time"

	"github.com/niko/mqtt-agent-orchestration/pkg/types"
)

// TaskProcessor defines the interface for processing tasks
type TaskProcessor interface {
	ProcessTask(ctx context.Context, task types.Task) (string, error)
}

// Worker represents a single worker instance
type Worker struct {
	id        string
	processor TaskProcessor
	
	// Statistics tracking
	mu           sync.RWMutex
	tasksTotal   int
	tasksError   int
	currentTask  string
	status       string
	lastSeen     time.Time
}

// NewWorker creates a new worker with the specified ID and processor
func NewWorker(id string, processor TaskProcessor) *Worker {
	return &Worker{
		id:        id,
		processor: processor,
		status:    "idle",
		lastSeen:  time.Now(),
	}
}

// ProcessTask processes a single task and returns the result
func (w *Worker) ProcessTask(ctx context.Context, task types.Task) types.TaskResult {
	startTime := time.Now()
	
	// Update worker status
	w.mu.Lock()
	w.status = "busy"
	w.currentTask = task.ID
	w.tasksTotal++
	w.lastSeen = time.Now()
	w.mu.Unlock()
	
	// Ensure we reset status when done
	defer func() {
		w.mu.Lock()
		w.status = "idle"
		w.currentTask = ""
		w.lastSeen = time.Now()
		w.mu.Unlock()
	}()
	
	// Process the task
	result, err := w.processor.ProcessTask(ctx, task)
	
	duration := time.Since(startTime)
	
	// Build result
	taskResult := types.TaskResult{
		TaskID:      task.ID,
		WorkerID:    w.id,
		ProcessedAt: time.Now(),
		Duration:    duration.Milliseconds(),
	}
	
	if err != nil {
		w.mu.Lock()
		w.tasksError++
		w.mu.Unlock()
		
		taskResult.Success = false
		taskResult.Error = err.Error()
	} else {
		taskResult.Success = true
		taskResult.Result = result
	}
	
	return taskResult
}

// GetStatus returns the current status of the worker
func (w *Worker) GetStatus() types.WorkerStatus {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	return types.WorkerStatus{
		ID:          w.id,
		Status:      w.status,
		LastSeen:    w.lastSeen,
		TasksTotal:  w.tasksTotal,
		TasksError:  w.tasksError,
		CurrentTask: w.currentTask,
	}
}

// GetID returns the worker's ID
func (w *Worker) GetID() string {
	return w.id
}

// UpdateLastSeen updates the worker's last seen timestamp
func (w *Worker) UpdateLastSeen() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.lastSeen = time.Now()
}