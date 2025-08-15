package types

import (
	"time"
)

// Task represents a work item to be processed by a worker
type Task struct {
	ID        string            `json:"id"`
	Type      string            `json:"type"`
	Payload   map[string]string `json:"payload"`
	CreatedAt time.Time         `json:"created_at"`
	Priority  int               `json:"priority"`
}

// TaskResult represents the result of processing a task
type TaskResult struct {
	TaskID      string    `json:"task_id"`
	WorkerID    string    `json:"worker_id"`
	Success     bool      `json:"success"`
	Result      string    `json:"result"`
	Error       string    `json:"error,omitempty"`
	ProcessedAt time.Time `json:"processed_at"`
	Duration    int64     `json:"duration_ms"`
}

// WorkerStatus represents the current status of a worker
type WorkerStatus struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"` // "idle", "busy", "error"
	LastSeen    time.Time `json:"last_seen"`
	TasksTotal  int       `json:"tasks_total"`
	TasksError  int       `json:"tasks_error"`
	CurrentTask string    `json:"current_task,omitempty"`
}