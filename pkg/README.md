# Public Packages (`pkg/`)

## Overview

The `pkg/` directory contains public packages that can be imported by external applications. These packages follow the **"Single Responsibility"** principle and provide reusable functionality for the MQTT Agent Orchestration System ecosystem.

## Architecture Philosophy

Following our **"Excellence through Rigor"** philosophy, public packages are:
- **Well-Documented**: Comprehensive documentation for all exported elements
- **Stable APIs**: Versioned interfaces with backward compatibility
- **Tested**: High test coverage with integration test examples
- **Reusable**: Designed for use by external applications

## Public Packages

### 1. `types/` - Core Type Definitions

**Purpose**: Shared type definitions and data structures used across the system.

**Design Principles**:
- **Type Safety**: Strong typing to prevent runtime errors
- **Immutability**: Immutable data structures where appropriate
- **Serialization**: JSON marshaling/unmarshaling support
- **Validation**: Built-in validation for data integrity

**Key Types**:

#### Workflow Types
```go
// WorkflowTask represents a task in the orchestration workflow
type WorkflowTask struct {
    ID          string                 `json:"id"`
    Content     string                 `json:"content"`
    Role        WorkerRole             `json:"role"`
    Stage       WorkflowStage          `json:"stage"`
    Priority    TaskPriority           `json:"priority"`
    Complexity  TaskComplexity         `json:"complexity"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

// WorkerRole defines the role of a worker in the system
type WorkerRole string

const (
    RoleDeveloper WorkerRole = "developer"
    RoleReviewer  WorkerRole = "reviewer"
    RoleApprover  WorkerRole = "approver"
    RoleTester    WorkerRole = "tester"
)

// WorkflowStage represents the current stage in the workflow
type WorkflowStage string

const (
    StageDevelopment WorkflowStage = "development"
    StageReview      WorkflowStage = "review"
    StageApproval    WorkflowStage = "approval"
    StageTesting     WorkflowStage = "testing"
    StageComplete    WorkflowStage = "complete"
)
```

#### RAG Types
```go
// RAGQuery represents a search query for the RAG system
type RAGQuery struct {
    Query      string  `json:"query"`
    Collection string  `json:"collection,omitempty"`
    TopK       int     `json:"top_k"`
    Threshold  float64 `json:"threshold"`
}

// RAGResponse represents the response from a RAG search
type RAGResponse struct {
    Query      string       `json:"query"`
    TotalHits  int          `json:"total_hits"`
    Documents  []RAGDocument `json:"documents"`
    SearchTime time.Duration `json:"search_time"`
}

// RAGDocument represents a document in the RAG system
type RAGDocument struct {
    ID       string                 `json:"id"`
    Content  string                 `json:"content"`
    Source   string                 `json:"source"`
    Score    float64                `json:"score"`
    Metadata map[string]interface{} `json:"metadata,omitempty"`
}
```

#### Worker Types
```go
// WorkerStatus represents the status of a worker
type WorkerStatus struct {
    ID       string    `json:"id"`
    Status   string    `json:"status"`
    LastSeen time.Time `json:"last_seen"`
}

// ExtendedWorkerStatus includes role-specific information
type ExtendedWorkerStatus struct {
    WorkerStatus
    Role         WorkerRole   `json:"role"`
    Capabilities []string     `json:"capabilities"`
    Load         float64      `json:"load"`
    Memory       MemoryUsage  `json:"memory"`
}

// MemoryUsage represents memory usage information
type MemoryUsage struct {
    Total     uint64  `json:"total"`
    Used      uint64  `json:"used"`
    Available uint64  `json:"available"`
    Usage     float64 `json:"usage_percent"`
}
```

**Usage Example**:
```go
package main

import (
    "github.com/niko/mqtt-agent-orchestration/pkg/types"
)

func main() {
    // Create a new workflow task
    task := types.WorkflowTask{
        ID:         "task-123",
        Content:    "Create a Go HTTP server",
        Role:       types.RoleDeveloper,
        Stage:      types.StageDevelopment,
        Priority:   types.PriorityHigh,
        Complexity: types.ComplexityMedium,
        CreatedAt:  time.Now(),
    }

    // Create a RAG query
    query := types.RAGQuery{
        Query:      "Go HTTP server best practices",
        Collection: "coding_standards",
        TopK:       5,
        Threshold:  0.7,
    }
}
```

### 2. `userservice/` - User Service Integration

**Purpose**: User service integration for authentication, authorization, and user management.

**Design Principles**:
- **Service Abstraction**: Clean interface for user service operations
- **Error Handling**: Comprehensive error handling and logging
- **Caching**: Intelligent caching for performance optimization
- **Security**: Secure handling of user credentials and tokens

**Key Features**:
- User authentication and authorization
- Role-based access control (RBAC)
- User profile management
- Session management
- Audit logging

**Core Components**:

#### RAG Manager
```go
// RAGManager provides RAG functionality for user services
type RAGManager struct {
    service *rag.Service
    cache   *cache.Cache
}

// GetRelevantContext retrieves relevant context for a user task
func (rm *RAGManager) GetRelevantContext(ctx context.Context, userID string, taskType string, content string) (string, error) {
    // Implementation with user-specific context retrieval
}

// StoreUserKnowledge stores user-specific knowledge
func (rm *RAGManager) StoreUserKnowledge(ctx context.Context, userID string, knowledge KnowledgeItem) error {
    // Implementation for storing user knowledge
}
```

**Usage Example**:
```go
package main

import (
    "github.com/niko/mqtt-agent-orchestration/pkg/userservice"
)

func main() {
    // Initialize RAG manager
    ragManager := userservice.NewRAGManager(ragService, cache)

    // Get relevant context for user task
    context, err := ragManager.GetRelevantContext(ctx, "user-123", "development", "Create API endpoint")
    if err != nil {
        log.Printf("Failed to get context: %v", err)
        return
    }

    // Store user knowledge
    knowledge := userservice.KnowledgeItem{
        Content:  "User prefers RESTful APIs",
        Category: "preferences",
        Tags:     []string{"api", "rest"},
    }
    
    err = ragManager.StoreUserKnowledge(ctx, "user-123", knowledge)
    if err != nil {
        log.Printf("Failed to store knowledge: %v", err)
    }
}
```

## Package Design Guidelines

### API Design Principles

1. **Consistency**: Consistent naming and structure across packages
2. **Simplicity**: Simple, intuitive interfaces
3. **Composability**: Packages can be composed together
4. **Extensibility**: Easy to extend without breaking changes

### Error Handling

```go
// All packages use consistent error handling
var (
    ErrNotFound     = errors.New("resource not found")
    ErrUnauthorized = errors.New("unauthorized access")
    ErrInvalidInput = errors.New("invalid input")
)

// Errors include context and can be wrapped
func (s *Service) GetResource(id string) (*Resource, error) {
    if id == "" {
        return nil, fmt.Errorf("%w: id cannot be empty", ErrInvalidInput)
    }
    
    resource, err := s.storage.Get(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get resource %s: %w", id, err)
    }
    
    return resource, nil
}
```

### Configuration

```go
// Packages support configuration through structs
type Config struct {
    Timeout     time.Duration `yaml:"timeout"`
    MaxRetries  int           `yaml:"max_retries"`
    CacheSize   int           `yaml:"cache_size"`
    EnableDebug bool          `yaml:"enable_debug"`
}

// Default configuration
func DefaultConfig() *Config {
    return &Config{
        Timeout:     30 * time.Second,
        MaxRetries:  3,
        CacheSize:   1000,
        EnableDebug: false,
    }
}
```

### Testing

```go
// All packages include comprehensive tests
func TestWorkflowTask_Validation(t *testing.T) {
    tests := []struct {
        name    string
        task    types.WorkflowTask
        wantErr bool
    }{
        {
            name: "valid task",
            task: types.WorkflowTask{
                ID:      "test-123",
                Content: "test content",
                Role:    types.RoleDeveloper,
            },
            wantErr: false,
        },
        {
            name: "missing ID",
            task: types.WorkflowTask{
                Content: "test content",
                Role:    types.RoleDeveloper,
            },
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := tt.task.Validate()
            if (err != nil) != tt.wantErr {
                t.Errorf("WorkflowTask.Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Versioning and Compatibility

### Semantic Versioning

Packages follow semantic versioning (SemVer):
- **Major**: Breaking changes
- **Minor**: New features, backward compatible
- **Patch**: Bug fixes, backward compatible

### Backward Compatibility

- **Interface Stability**: Public interfaces remain stable within major versions
- **Deprecation Policy**: Deprecated features are marked and documented
- **Migration Guides**: Clear migration guides for breaking changes

### Dependency Management

```go
// go.mod example
module github.com/niko/mqtt-agent-orchestration

go 1.24

require (
    github.com/qdrant/go-client v1.0.0
    github.com/eclipse/paho.mqtt.golang v1.4.0
    gopkg.in/yaml.v3 v3.0.1
)
```

## Documentation Standards

### Package Documentation

```go
// Package types provides core type definitions for the MQTT Agent Orchestration System.
//
// This package defines the fundamental data structures used throughout the system,
// including workflow tasks, worker status, and RAG operations. All types are designed
// for JSON serialization and include validation methods.
//
// Example:
//
//	task := types.WorkflowTask{
//	    ID:      "task-123",
//	    Content: "Create a Go HTTP server",
//	    Role:    types.RoleDeveloper,
//	}
package types
```

### Function Documentation

```go
// NewWorkflowTask creates a new workflow task with the given parameters.
//
// The function validates the input parameters and returns an error if any
// required fields are missing or invalid. The task is initialized with
// default values for optional fields.
//
// Parameters:
//   - id: Unique identifier for the task
//   - content: Task content or description
//   - role: Worker role responsible for the task
//
// Returns:
//   - *WorkflowTask: The created task
//   - error: Validation or creation error
//
// Example:
//
//	task, err := types.NewWorkflowTask("task-123", "Create API", types.RoleDeveloper)
//	if err != nil {
//	    log.Fatal(err)
//	}
func NewWorkflowTask(id, content string, role WorkerRole) (*WorkflowTask, error) {
    // Implementation
}
```

## Performance Considerations

### Memory Management

- **Efficient Structs**: Optimized struct layouts for memory efficiency
- **Object Pooling**: Reuse objects where appropriate
- **Garbage Collection**: Minimize allocations in hot paths

### Caching Strategy

```go
// Packages implement intelligent caching
type Cache struct {
    data    map[string]interface{}
    maxSize int
    mutex   sync.RWMutex
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mutex.RLock()
    defer c.mutex.RUnlock()
    
    value, exists := c.data[key]
    return value, exists
}
```

## Security Considerations

### Input Validation

```go
// All packages validate input data
func (t *WorkflowTask) Validate() error {
    if t.ID == "" {
        return errors.New("task ID cannot be empty")
    }
    
    if t.Content == "" {
        return errors.New("task content cannot be empty")
    }
    
    if !isValidRole(t.Role) {
        return fmt.Errorf("invalid role: %s", t.Role)
    }
    
    return nil
}
```

### Secure Serialization

```go
// Packages use secure serialization methods
func (t *WorkflowTask) MarshalJSON() ([]byte, error) {
    // Sanitize sensitive data before serialization
    sanitized := t.sanitizeForSerialization()
    return json.Marshal(sanitized)
}
```

## Integration Examples

### External Application Integration

```go
package main

import (
    "github.com/niko/mqtt-agent-orchestration/pkg/types"
    "github.com/niko/mqtt-agent-orchestration/pkg/userservice"
)

func main() {
    // Create workflow task
    task := types.WorkflowTask{
        ID:      "external-task-123",
        Content: "Process external request",
        Role:    types.RoleDeveloper,
    }
    
    // Initialize user service
    userService := userservice.NewService(config)
    
    // Process task with user context
    result, err := userService.ProcessTask(ctx, "user-456", task)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Task result: %s\n", result)
}
```

### Microservice Integration

```go
// HTTP handler using pkg types
func handleTask(w http.ResponseWriter, r *http.Request) {
    var task types.WorkflowTask
    if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    if err := task.Validate(); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    // Process task...
    w.WriteHeader(http.StatusOK)
}
```

## Troubleshooting

### Common Issues

1. **Import Errors**: Ensure correct module path and version
2. **Type Mismatches**: Check type compatibility across versions
3. **Validation Errors**: Verify input data meets requirements
4. **Performance Issues**: Monitor memory usage and cache efficiency

### Debug Tools

```go
// Enable debug logging
import "github.com/niko/mqtt-agent-orchestration/pkg/types"

// Set debug mode
types.SetDebugMode(true)

// Validate with detailed errors
if err := task.ValidateWithDetails(); err != nil {
    log.Printf("Validation failed: %+v", err)
}
```

## Future Enhancements

### Planned Features

- **Generic Types**: Enhanced generic type support for Go 1.24+
- **Streaming**: Streaming support for large data sets
- **Metrics**: Built-in metrics collection
- **Tracing**: Distributed tracing support

### Extension Points

- **Custom Validators**: Plugin system for custom validation
- **Serialization Formats**: Support for additional formats (Protocol Buffers, MessagePack)
- **Caching Backends**: Pluggable caching backends
- **Authentication Providers**: Extensible authentication system

---

**Production Ready**: Public packages are designed for production use with comprehensive testing, documentation, and backward compatibility guarantees. They provide a stable foundation for building applications that integrate with the MQTT Agent Orchestration System.
