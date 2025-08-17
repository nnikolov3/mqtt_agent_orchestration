# MQTT Agent Orchestration System

## Purpose

Autonomous AI agent orchestration implementing role-based workflow coordination. The system follows the **"Excellence through Rigor"** philosophy with **"Do more with less"** architecture principles, providing explicit behavior, reliable operation, and comprehensive error handling.

## Overview

**Core Design**: Role-based AI workers coordinate over MQTT to process complex workflows
**Architecture**: Modular system with clear boundaries, explicit interfaces, and composable components
**Knowledge Management**: RAG-enhanced processing using local Qdrant vector database
**AI Integration**: Intelligent routing between local GGUF models and external AI services
**Error Handling**: **"Fail Fast, Fail Loud"** with comprehensive diagnostics and recovery

## System Components

### Core Services
- **Orchestrator** (`server`): HTTP API and MQTT-based workflow coordination engine
- **Role Workers** (`role-worker`): Specialized AI agents for development, review, approval, testing
- **RAG Service** (`rag-service`): Knowledge management with semantic search and context retrieval
- **Client Interface** (`client`): Command-line interface for system interaction and monitoring

### Supporting Services
- **Embedding Worker** (`embedding-worker`): Local GGUF model inference for vector generation
- **Operations CLI** (`ops`): System management, build automation, and deployment orchestration

### Architecture Principles Applied
- **Single Responsibility**: Each component has one clear purpose
- **Explicit Interfaces**: All communication through well-defined APIs
- **Graceful Degradation**: System continues with reduced capacity when components fail
- **Resource Efficiency**: Intelligent caching and on-demand resource allocation

## System Prerequisites

### Required Dependencies
- **Go 1.24+**: Primary development language with latest features
- **MQTT Broker**: Mosquitto or compatible (default port: 1883)
- **Vector Database**: Qdrant local binary (searched in PATH, then `bin/qdrant`)

### Optional Components
- **Local Models**: llama.cpp binaries (`llama-server`, `llama-embedding`) for GGUF model inference
- **GPU Acceleration**: NVIDIA GPU with CUDA support for enhanced performance
- **External AI Services**: API keys for Cerebras, NVIDIA, Gemini, Grok, Groq (configured via `configs/ai_helpers.toml`)

### Configuration Philosophy
Following **"Never hard code values"** principle:
- All ports, paths, and endpoints configurable via environment variables
- Service discovery with intelligent fallbacks
- Explicit configuration validation at startup

## Build System

### Automated Build Process
Following **"Test, confirm, validate, lint, analyze, refactor, and improve"** principle:

```bash
# Complete build with validation
./bin/ops build

# Individual component build
go build -o bin/<component> ./cmd/<component>

# Build with comprehensive checks
./bin/ops build --with-tests --with-lint
```

**Build Verification**: All binaries placed in `./bin/` with validation checksums

### Quality Gates Applied
- Syntax validation and compilation checks
- Automated testing execution
- Code quality and security analysis  
- Dependency validation and updates

## System Execution

### Autonomous Startup
Following **"Service orchestration"** with **"Graceful degradation"** principles:

```bash
# Complete system startup with health verification
./bin/ops run --verbose

# Development mode with enhanced debugging
./bin/ops run --dev --debug
```

### Startup Process
1. **Dependency Verification**: MQTT broker availability (default: localhost:1883)
2. **Service Health**: Qdrant HTTP health check (default: localhost:6333) 
3. **Auto-Recovery**: Automatic Qdrant startup if not available
4. **Component Initialization**: Orchestrator and role workers with health monitoring

### Service Architecture
- **Qdrant Communication**: gRPC on localhost:6334 (workers), HTTP on localhost:6333 (health)
- **Local Binary Preference**: Direct binary execution, avoiding container complexity

## HTTP API Interface

### RESTful Endpoints
Following **"Explicit interfaces"** and **"Clear boundaries"** principles:

**Health Monitoring**:
```bash
GET http://localhost:8080/health
# Response: {"status": "healthy", "timestamp": "2024-01-15T10:30:00Z"}
```

**Workflow Submission**:
```bash
POST http://localhost:8080/workflow/submit
Content-Type: application/json
```

**Request Schema**:
```json
{
  "content": "Create a Go REST API with authentication",
  "complexity": "high",
  "metadata": {"project": "demo", "priority": "normal"}
}
```

### Workflow Processing
- **Task Distribution**: Orchestrator publishes to role-specific MQTT topics
- **Result Aggregation**: Collects worker responses with correlation tracking
- **Error Handling**: **"Fail Fast, Fail Loud"** with detailed diagnostics
- **State Management**: Persistent workflow state across system restarts

## System Demonstrations

### Worker Discovery and Status
Following **"Observable"** principle with comprehensive status monitoring:

```bash
# List available workers with health status
./bin/client --list-workers

# Expected output: Worker roles, capabilities, current load
```

### Workflow Execution Examples
Following **"Break work into smallest possible tasks"** principle:

```bash
# Submit development task with explicit parameters
./bin/client --submit-workflow \
  --content "Implement HTTP middleware for request latency logging" \
  --complexity medium \
  --role developer

# Submit with metadata for context
./bin/client --submit-workflow \
  --content "Review authentication implementation for security" \
  --complexity high \
  --role reviewer \
  --metadata '{"project": "api-service", "deadline": "2024-01-20"}'
```

### Real-time Monitoring
Following **"Observable systems"** principle:

```bash
# Monitor task distribution
mosquitto_sub -h localhost -p 1883 -t "tasks/workflow/+" -v

# Monitor results aggregation  
mosquitto_sub -h localhost -p 1883 -t "results/workflow/+" -v

# Structured log monitoring
tail -f logs/*.log | jq '.'
```

### Documentation Generation
Following **"Self-documenting code"** principle:

```bash
# Generate coding standards documentation
./bin/client --doc-type go_coding_standards --output /tmp/standards.md

# Generate system architecture documentation
./bin/client --doc-type architecture --output /tmp/architecture.md
```

### System Behavior
- **Workflow Orchestration**: Demonstrates complete task lifecycle management
- **Progress Tracking**: Monitor via structured logs and MQTT message flows  
- **Result Persistence**: Check `logs/` directory and MQTT topics for complete audit trail

## Knowledge Management (RAG)

### Vector Database Operations
Following **"Knowledge persistence"** and **"Semantic search"** principles:

**Service Initialization**:
```bash
# Auto-started during system startup, or manual startup:
qdrant                    # If in PATH
./bin/qdrant             # Local binary execution
```

**Collection Management**:
```bash
# Initialize all knowledge collections with proper schema
./bin/rag-service init

# Initialize specific collection with validation
./bin/rag-service init --collection coding_standards --validate
```

### Document Storage and Retrieval
Following **"Explicit metadata"** and **"Semantic matching"** principles:

```bash
# Store knowledge with comprehensive metadata
./bin/rag-service add-document \
  --collection coding_standards \
  --content "In Go, handle every error explicitly using if err != nil pattern" \
  --metadata '{"language":"go","category":"error_handling","confidence":0.95}'

# Semantic search with intelligent ranking
./bin/rag-service search \
  --query "go error handling best practices" \
  --collection coding_standards \
  --limit 5 \
  --threshold 0.7

# Context-aware task assistance  
./bin/rag-service context \
  --project myapp \
  --phase development \
  --task "implement HTTP middleware with error handling"
```

### Service Architecture
Following **"Clear boundaries"** principle:
- **Health Monitoring**: HTTP endpoint at `localhost:6333/health`
- **Worker Communication**: gRPC interface at `localhost:6334`
- **Embedding Generation**: Qwen3-Embedding-4B for 2560-dimensional vectors
- **Fallback Strategy**: Hash-based search when embeddings unavailable

## Local Model Integration

### Embedding Worker (Optional Enhancement)
Following **"Resource efficiency"** and **"Graceful degradation"** principles:

**GGUF Model Deployment**:
```bash
# Deploy embedding worker with local model
./bin/embedding-worker \
  --mqtt-host localhost \
  --mqtt-port 1883 \
  --model-path /data/models/Qwen3-Embedding-4B-Q8_0.gguf \
  --gpu-acceleration \
  --memory-limit 2048MB

# Configure with comprehensive monitoring  
./bin/embedding-worker \
  --config configs/embedding-worker.yaml \
  --health-check-interval 30s \
  --metrics-port 9090
```

### Service Integration
- **Primary Mode**: RAG service requests embeddings via MQTT with correlation tracking
- **Fallback Strategy**: Deterministic hash-based embedding when local model unavailable
- **Performance Monitoring**: Real-time GPU utilization and response time metrics
- **Resource Management**: Automatic memory cleanup and model lifecycle management

## Adaptive Learning System

### Feedback Collection and Analysis
Following **"Continuous improvement"** and **"Validate, analyze, refactor"** principles:

**Feedback Architecture**:
```go
// Task outcome analysis with comprehensive metrics
import (
  "time"
  "github.com/niko/mqtt-agent-orchestration/internal/rl"
  "github.com/niko/mqtt-agent-orchestration/internal/rag"
)

func implementFeedbackCollection() error {
  // Initialize services with explicit configuration
  ragService, err := rag.NewService("qdrant", "localhost:6334")
  if err != nil {
    return fmt.Errorf("RAG service initialization failed: %w", err)
  }
  
  collector := rl.NewFeedbackCollector(ragService)
  defer collector.Close()

  // Comprehensive task context with explicit parameters
  taskContext := rl.TaskContext{
    OriginalPrompt: "Implement HTTP middleware with request logging",
    Response:       "// Generated implementation with error handling",
    TaskType:       "development",
    Complexity:     "medium",
    ProcessingMode: "LOCAL_MODEL_PREFERRED",
    WorkerRole:     "developer",
  }

  // Detailed execution metrics for analysis
  metrics := rl.ExecutionMetrics{
    ProcessingDuration: 2 * time.Second,
    SuccessRate:       1.0,
    QualityScore:      0.92,
    ResourceUsage:     0.75,
    ErrorCount:        0,
  }

  // Store feedback with full context for learning
  err = collector.CollectSuccessFeedback(
    "task-001", "worker-dev-1", "local-qwen-omni", 
    "qwen-omni-3b", taskContext, metrics
  )
  
  return err
}
```

### Learning Integration
- **Knowledge Persistence**: Feedback converted to RAG documents for future reference
- **Pattern Recognition**: Successful approaches stored for similar task types
- **Performance Optimization**: Model selection based on historical success rates
- **Quality Improvement**: Continuous refinement of worker responses based on feedback

## System Configuration

### Service Endpoints
Following **"Never hard code values"** principle with configurable defaults:

- **Qdrant Vector Database**:
  - HTTP Health Check: `localhost:6333` (configurable via `QDRANT_URL`)
  - gRPC Communication: `localhost:6334` (configurable via `QDRANT_GRPC`)
- **MQTT Message Broker**:
  - Connection: `localhost:1883` (configurable via `MQTT_HOST`, `MQTT_PORT`)

### Environment Configuration
**Required Environment Variables**:
```bash
export LOCAL_MODELS_PATH="/data/models"           # Local GGUF model storage
export PROJECT_ROOT="$(pwd)"                     # Project root directory
export QDRANT_STORAGE_PATH="/data/qdrant"        # Vector database storage
```

**Optional Environment Variables**:
```bash
export LLAMA_CLI_PATH="/usr/local/bin"           # llama.cpp binary location
export LLAMA_SERVER_PATH="/usr/local/bin"        # llama-server binary location
export QDRANT_URL="http://localhost:6333"        # Qdrant HTTP endpoint
export QDRANT_GRPC="localhost:6334"             # Qdrant gRPC endpoint
```

### Configuration Files
Following **"Explicit configuration"** and **"Single source of truth"** principles:

- **`configs/models.yaml`**: Local GGUF model configuration and GPU memory management
- **`configs/mcp.yaml`**: Model Context Protocol service definitions
- **`configs/ai_helpers.toml`**: External AI service API keys and routing

## Diagnostics and Troubleshooting

### System Health Verification
Following **"Fail fast, fail loud"** with comprehensive diagnostics:

```bash
# Verify MQTT broker connectivity
mosquitto_pub -h localhost -p 1883 -t "health/test" -m "ping"
mosquitto_sub -h localhost -p 1883 -t "health/test" -C 1

# Verify Qdrant vector database health
curl http://localhost:6333/health
# Expected: {"status":"ok"}

# Check service dependencies
./bin/client --health-check --verbose

# Monitor structured logs with correlation tracking
tail -f logs/*.log | jq '.'
```

### Automatic Recovery Features
- **Service Discovery**: Automatic Qdrant startup when not detected
- **Connection Recovery**: MQTT reconnection with exponential backoff
- **Resource Monitoring**: GPU memory and model lifecycle management
- **Graceful Degradation**: Fallback to external AI services when local models unavailable

## Development Workflow

### Quality Assurance Process
Following **"Test, confirm, validate, lint, analyze, refactor, and improve"** principle:

```bash
# Comprehensive code quality checks
./bin/ops lint --fix --comprehensive

# Test execution with race detection
go test ./... -race -cover -timeout 30s

# Build with validation and optimization
./bin/ops build --optimize --validate

# Specific component development
go build -ldflags="-X main.version=$(git describe --tags)" \
  -o bin/role-worker ./cmd/role-worker
```

### Development Standards
- **Error Handling**: Every error explicitly checked and handled
- **Testing**: Unit tests with >90% coverage, integration tests for workflows
- **Documentation**: Self-documenting code with comprehensive godoc comments
- **Performance**: Benchmarking and profiling for critical paths

## Project License

MIT License - Supporting open collaboration and knowledge sharing