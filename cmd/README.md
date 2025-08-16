# Command Line Applications (`cmd/`)

## Overview

The `cmd/` directory contains all executable applications that form the MQTT Agent Orchestration System. Each application follows the **Single Responsibility Principle** and implements **Fail Fast, Fail Loud** error handling as defined in our design principles.

## Architecture Philosophy

Following our **Excellence through Rigor** philosophy, each executable is:
- **Self-contained**: Minimal dependencies, clear entry points
- **Observable**: Comprehensive logging and metrics
- **Resilient**: Graceful degradation and error recovery
- **Testable**: Designed for integration testing

## Applications

### 1. `orchestrator/` - Workflow Orchestration Engine

**Purpose**: Central workflow management and task routing system.

**Design Principles**:
- **Single Responsibility**: Manages task lifecycle and worker coordination
- **Fail Fast, Fail Loud**: Immediate error detection and reporting
- **Graceful Degradation**: Continues operation with reduced worker capacity

**Key Features**:
- MQTT-based task distribution
- Worker health monitoring
- Workflow state management
- Load balancing across workers
- Circuit breaker pattern implementation

**Usage**:
```bash
./bin/orchestrator --mqtt-host localhost --mqtt-port 1883 --verbose
```

### 2. `role-worker/` - Specialized AI Agent Workers

**Purpose**: Role-based AI agents that process tasks according to their specialization.

**Design Principles**:
- **Role-Based Processing**: Each worker has a specific expertise domain
- **Intelligent Routing**: Automatic model selection based on task complexity
- **RAG Integration**: Context-aware processing using knowledge base

**Supported Roles**:
- **Developer**: Code generation and implementation
- **Reviewer**: Code review and quality assessment
- **Approver**: Final validation and approval
- **Tester**: Testing strategy and execution

**Usage**:
```bash
./bin/role-worker --role developer --id dev-1 --mqtt-host localhost
./bin/role-worker --role reviewer --id rev-1 --mqtt-host localhost
```

### 3. `rag-service/` - RAG Knowledge Management

**Purpose**: Retrieval-Augmented Generation service for knowledge base operations.

**Design Principles**:
- **Knowledge Persistence**: Vector-based document storage and retrieval
- **Semantic Search**: Qwen3-Embedding-4B for intelligent matching
- **Fallback Strategy**: Hash-based search when embeddings unavailable

**Key Features**:
- Qdrant vector database integration
- Document indexing and search
- System prompt management
- Context retrieval for tasks
- Collection management

**Usage**:
```bash
./bin/rag-service init                    # Initialize collections
./bin/rag-service add-document --collection coding_standards --content "..."
./bin/rag-service search --query "go best practices" --limit 5
```

### 4. `client/` - System Interaction Interface

**Purpose**: Command-line interface for system interaction and testing.

**Design Principles**:
- **User-Friendly**: Intuitive command structure
- **Comprehensive**: Access to all system features
- **Diagnostic**: Built-in health checking and debugging

**Key Features**:
- Task submission and monitoring
- System health checks
- Model management
- API testing
- Performance benchmarking

**Usage**:
```bash
./bin/client --task "Create Go function" --role developer
./bin/client --health-check
./bin/client --list-models
./bin/client --benchmark-mqtt --messages 1000
```

### 5. `worker/` - Generic Worker Implementation

**Purpose**: Base worker implementation for custom role extensions.

**Design Principles**:
- **Extensible**: Framework for custom worker types
- **Reusable**: Common worker functionality
- **Configurable**: Flexible task processing

**Usage**:
```bash
./bin/worker --config worker-config.yaml --mqtt-host localhost
```

### 6. `server/` - HTTP API Server

**Purpose**: RESTful API server for external system integration.

**Design Principles**:
- **RESTful Design**: Standard HTTP API patterns
- **JSON Communication**: Structured data exchange
- **Authentication**: Secure access control

**Key Features**:
- Task management endpoints
- Worker status monitoring
- Model management API
- Health check endpoints

**Usage**:
```bash
./bin/server --port 8080 --mqtt-host localhost
```

## Development Standards

### Error Handling
- **Explicit Error Checking**: All errors are checked and handled
- **Structured Logging**: JSON-formatted logs with correlation IDs
- **Graceful Shutdown**: Proper cleanup on termination signals

### Configuration Management
- **Environment Variables**: Sensitive configuration via env vars
- **Configuration Files**: YAML/TOML for complex settings
- **Validation**: Configuration validation at startup

### Testing Strategy
- **Integration Tests**: End-to-end workflow testing
- **Unit Tests**: Individual component testing
- **Performance Tests**: Load and stress testing

## Deployment Considerations

### Resource Requirements
- **Memory**: 512MB-2GB per worker (depending on model loading)
- **CPU**: 2-4 cores per worker
- **GPU**: Optional, for local model acceleration
- **Network**: MQTT broker connectivity

### Monitoring
- **Health Checks**: Built-in health endpoints
- **Metrics**: Prometheus-compatible metrics
- **Logging**: Structured logging for observability
- **Alerting**: Configurable alert thresholds

### Security
- **Authentication**: MQTT authentication and authorization
- **Encryption**: TLS for MQTT communication
- **Input Validation**: All inputs validated and sanitized
- **Audit Logging**: Security-relevant event logging

## Troubleshooting

### Common Issues
1. **MQTT Connection Failures**: Check broker status and network connectivity
2. **Model Loading Errors**: Verify GPU memory and model file availability
3. **RAG Service Issues**: Check Qdrant connectivity and collection status
4. **Worker Coordination**: Monitor orchestrator logs for workflow issues

### Debug Commands
```bash
# Check MQTT connectivity
mosquitto_pub -h localhost -p 1883 -t test -m "hello"

# Verify Qdrant status
curl http://localhost:6333/health

# Monitor system logs
tail -f logs/*.log

# Check GPU memory
nvidia-smi
```

## Performance Optimization

### Worker Scaling
- **Horizontal Scaling**: Add workers for increased throughput
- **Load Balancing**: Orchestrator distributes tasks evenly
- **Resource Monitoring**: Track CPU, memory, and GPU usage

### Caching Strategy
- **Model Caching**: LRU cache for frequently used models
- **RAG Caching**: Query result caching for repeated searches
- **Connection Pooling**: Reused MQTT and HTTP connections

## Future Enhancements

### Planned Features
- **Dynamic Worker Scaling**: Automatic worker provisioning
- **Advanced Routing**: ML-based task routing optimization
- **Multi-Modal Support**: Enhanced image and document processing
- **Distributed Deployment**: Multi-node cluster support

### Integration Opportunities
- **Kubernetes**: Container orchestration integration
- **Service Mesh**: Advanced networking and security
- **Observability Stack**: Prometheus, Grafana, Jaeger integration
- **CI/CD Integration**: Automated testing and deployment

---

**Production Ready**: All applications are designed for production deployment with comprehensive error handling, monitoring, and scalability features. Each executable can operate independently while contributing to the overall system functionality.
