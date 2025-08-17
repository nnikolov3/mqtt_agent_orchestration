# MQTT Agent Orchestration System

## Overview

Autonomous AI agent orchestration in Go (1.24+). The system coordinates role-based workers over MQTT, optionally augments with RAG via a local Qdrant binary (no Docker required), and can leverage local GGUF models and/or external AI helpers. Design focuses on explicit behavior, reliability, and clear error handling.

## Components

- Orchestrator (HTTP + MQTT): routes workflow stages between workers
- Role workers: developer, reviewer, approver, tester
- CLI client: submit workflows and list workers
- RAG service (CLI): initialize/search Qdrant collections; simple embedding fallback when local model not available
- Embedding worker (optional): generates embeddings on request via MQTT
- Ops CLI: build, run, lint, cleanup, generate MCP compose, training-data helper

## Requirements

- Go 1.24+
- MQTT broker: Mosquitto (listening on 1883)
- Qdrant local binary (in PATH or in `bin/qdrant`)
- Optional:
  - llama.cpp binaries for local GGUF models (e.g., `llama-server`, `llama-embedding`)
  - NVIDIA GPU for acceleration
  - External AI helper tools (if present in PATH or `bin/`)

## Build

```bash
# From repository root
./bin/ops build
# or: go build -o bin/<name> ./cmd/<name>
```

Binaries are placed in `./bin/`.

## Run (local, non-Docker)

```bash
# This will:
# - Verify MQTT on :1883
# - Check Qdrant HTTP health on :6333 (for visibility)
# - If Qdrant is not up, attempt to start local `qdrant` (from PATH or ./bin/qdrant)
# - Start orchestrator and all role workers
./bin/ops run --verbose
```

Notes:
- Workers talk to Qdrant via gRPC `localhost:6334`. Qdrant HTTP health is on `localhost:6333`.
- Qdrant is expected to be a local binary, not Docker.

## Orchestrator HTTP API

- Health: `GET http://localhost:8080/health` → "Orchestrator is healthy"
- Submit workflow: `POST http://localhost:8080/workflow/submit` with JSON body:

```json
{
  "content": "Create a Go REST API with authentication",
  "complexity": "high",
  "metadata": {"project": "demo"}
}
```

The orchestrator publishes workflow tasks to MQTT topics and listens for results.

## Quick Demos

### List workers (via MQTT status)

```bash
./bin/client --list-workers
```

### Submit a workflow (CLI)

```bash
./bin/client --submit-workflow \
  --content "Implement an HTTP middleware that logs latency" \
  --complexity medium
```

Monitor with:

```bash
mosquitto_sub -h localhost -p 1883 -t "tasks/workflow/+" -v
mosquitto_sub -h localhost -p 1883 -t "results/workflow/+" -v
# or tail logs
tail -f logs/*.log
```

### Create a documentation task

```bash
./bin/client --doc-type go_coding_standards --output /tmp/GO_CODING_STANDARDS.md
```

Note: The current implementation demonstrates workflow orchestration and logging. Output files are not guaranteed to be written automatically by the orchestrator; use logs/ and MQTT results to track progress and final content.

## RAG (Qdrant) Usage

Start Qdrant locally (if not auto-started by `ops run`):

```bash
qdrant
# or
./bin/qdrant
```

Initialize collections and try queries:

```bash
# Initialize collections
./bin/rag-service init

# Add a document
./bin/rag-service add-document \
  --collection coding_standards \
  --content "In Go, handle every error explicitly." \
  --metadata '{"language":"go","category":"error_handling"}'

# Search
./bin/rag-service search --query "go error handling" --limit 3

# Get context for a task
./bin/rag-service context myapp development "http server middleware"
```

Ports:
- HTTP health: `http://localhost:6333/health`
- gRPC (used by workers/RAG service): `localhost:6334`

## Embedding Worker (optional)

If you have `llama-embedding` and a compatible GGUF embedding model:

```bash
./bin/embedding-worker --mqtt-host localhost \
  --mqtt-port 1883 \
  --model-path /data/models/Qwen3-Embedding-4B-Q8_0.gguf
```

The RAG service will request embeddings via MQTT topics when available; otherwise it falls back to a simple deterministic embedding.

## Reinforcement Learning (Feedback Collection)

A built-in feedback collector captures task outcomes (success/failure) and quality metrics and prepares them for storage. Example usage in Go:

```go
import (
  "time"
  "github.com/niko/mqtt-agent-orchestration/internal/rl"
  "github.com/niko/mqtt-agent-orchestration/internal/rag"
)

func demoRL() error {
  ragSvc, _ := rag.NewService("qdrant", "localhost:6334")
  collector := rl.NewFeedbackCollector(ragSvc)
  defer collector.Close()

  ctx := rl.TaskContext{
    OriginalPrompt: "Implement HTTP middleware",
    Response:       "// code...",
    TaskType:       "development",
    Complexity:     "medium",
    Mode:           "LOCAL_ONLY",
  }

  metrics := rl.ExecutionMetrics{Duration: 2 * time.Second, SuccessRate: 1.0}
  if err := collector.CollectSuccessFeedback("task-1", "worker-1", "local", "qwen-omni", ctx, metrics); err != nil {
    return err
  }
  return nil
}
```

Current implementation logs feedback and provides conversion utilities to RAG documents. You can extend `storeFeedbackBatch` to persist to Qdrant once you define a target collection.

## Configuration

- Qdrant: local binary; HTTP `:6333`, gRPC `:6334`
- MQTT: `localhost:1883`
- Environment (optional):
  - `LOCAL_MODELS_PATH`, `LLAMA_CLI_PATH`, `LLAMA_SERVER_PATH`
  - `QDRANT_URL` (HTTP health), `QDRANT_GRPC` (gRPC address)
- Config files (optional):
  - `configs/models.yaml` for local models
  - `configs/ai_helpers.toml` for external helper tooling

## Troubleshooting

- MQTT: publish/subscribe a test topic to validate (`mosquitto_pub/sub`)
- Qdrant: `curl http://localhost:6333/health` (ensure the local binary is running)
- Logs: `tail -f logs/*.log`
- Ops run tries to start local `qdrant` automatically when missing

## Development

```bash
# Lint & tests
./bin/ops lint --fix
go test ./... -race

# Build specific binaries
go build -o bin/role-worker ./cmd/role-worker
```

## License

MIT