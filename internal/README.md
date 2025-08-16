# Internal System Architecture (`internal/`)

## Overview

The `internal/` directory contains the core system architecture of the MQTT Agent Orchestration System. This directory implements the **"Excellence through Rigor"** philosophy with every component designed for maximum reliability, performance, and maintainability.

## Architecture Philosophy

Following our **"Do more with less"** principle, the internal system is:
- **Modular**: Each package has a single, well-defined responsibility
- **Resilient**: Implements **"Fail Fast, Fail Loud"** error handling
- **Observable**: Comprehensive logging and metrics throughout
- **Testable**: Designed for comprehensive unit and integration testing

## Core Components

### 1. `ai/` - AI Service Integration

**Purpose**: Intelligent model routing and external AI service integration.

**Design Principles**:
- **Intelligent Routing**: Automatic model selection based on task complexity
- **Graceful Degradation**: Fallback mechanisms when services are unavailable
- **Cost Optimization**: Route to appropriate services based on task requirements

**Key Features**:
- External AI API integration (Cerebras, NVIDIA, Gemini, Grok, Groq)
- Model selection based on task complexity and content type
- Automatic retry logic with exponential backoff
- Response caching and optimization
- Rate limiting and cost management

**Configuration**:
```yaml
# Uses ~/.claude/claude_helpers.toml for API configuration
# Automatic provider selection based on task requirements
```

### 2. `config/` - Configuration Management

**Purpose**: Centralized configuration management following **"Never hard code values"** principle.

**Design Principles**:
- **Single Source of Truth**: All configuration centralized
- **Environment Agnostic**: Works across development, staging, production
- **Validation**: Configuration validation at startup

**Key Features**:
- YAML/TOML configuration parsing
- Environment variable integration
- Configuration validation and defaults
- Hot-reload capability for dynamic configuration

### 3. `localmodels/` - Local Model Management

**Purpose**: Local GGUF model management with LRU cache optimization.

**Design Principles**:
- **GPU Memory Management**: Intelligent memory allocation and deallocation
- **Performance Optimization**: LRU cache for frequently used models
- **Resource Efficiency**: Load models on-demand, unload when idle

**Supported Models**:
- **Qwen2.5-Omni-3B**: General text tasks (3GB memory)
- **Qwen2.5-VL-7B**: Multimodal tasks (4GB memory)
- **LLaVA-Llama-3-8B**: Multimodal tasks (4GB memory)
- **MiMo-VL-7B**: Multimodal tasks (4GB memory)
- **Qwen3-Embedding-4B**: Vector embeddings (2GB memory)

**Key Features**:
- Automatic GPU memory monitoring
- LRU cache management (max 3 concurrent models)
- Model loading/unloading optimization
- Fallback to external AI when local models unavailable

### 4. `mcp/` - Model Context Protocol Integration

**Purpose**: Standardized tool access through MCP for external services.

**Design Principles**:
- **Standardized Access**: Consistent interface for external tools
- **Service Discovery**: Automatic discovery of available MCP servers
- **Graceful Degradation**: Fallback when MCP services unavailable

**Supported MCP Servers**:
- **Qdrant MCP**: Vector database operations
- **File System MCP**: File operations and project management
- **Git MCP**: Version control operations
- **Local Models MCP**: Model management and inference

**Key Features**:
- Connection pooling and management
- Request caching and batching
- Health monitoring and failover
- Security and authentication support

### 5. `mqtt/` - MQTT Communication Layer

**Purpose**: Asynchronous message-based communication between system components.

**Design Principles**:
- **Reliable Delivery**: QoS=1 for guaranteed message delivery
- **Scalable Architecture**: Pub/sub pattern for easy scaling
- **Fault Tolerance**: Connection recovery and message persistence

**Key Features**:
- Persistent connections with keep-alive
- Automatic reconnection on failure
- Message queuing and delivery confirmation
- Topic-based message routing
- Security with TLS and authentication

### 6. `rag/` - Retrieval-Augmented Generation

**Purpose**: Knowledge management and context retrieval for AI tasks.

**Design Principles**:
- **Semantic Search**: Intelligent content matching using embeddings
- **Fallback Strategy**: Hash-based search when embeddings unavailable
- **Knowledge Persistence**: Vector-based document storage

**Key Features**:
- Qdrant vector database integration
- Qwen3-Embedding-4B for 2560-dimensional vectors
- Document indexing and semantic search
- System prompt management per role
- Context retrieval for task processing

### 7. `worker/` - Role-Based Worker System

**Purpose**: Specialized AI agents for different workflow stages.

**Design Principles**:
- **Role Specialization**: Each worker has specific expertise
- **Content Analysis**: Intelligent task routing based on content
- **Model Optimization**: Automatic model selection per task

**Supported Roles**:
- **Developer**: Code generation and implementation
- **Reviewer**: Code review and quality assessment
- **Approver**: Final validation and approval
- **Tester**: Testing strategy and execution

**Key Features**:
- Content type detection (multimodal, code, documentation, general)
- Complexity assessment (low, medium, high)
- Automatic model routing based on task requirements
- RAG context integration for enhanced responses

## System Integration

### Workflow Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   MQTT Client   │◀──▶│   Orchestrator  │◀──▶│   Role Workers  │
│   (External)    │    │   (Workflow)    │    │   (Processing)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                        │                        │
         ▼                        ▼                        ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   AI Services   │    │   RAG Service   │    │   Local Models  │
│   (External)    │    │   (Knowledge)   │    │   (Inference)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### Data Flow

1. **Task Submission**: External client submits task via MQTT
2. **Orchestration**: Orchestrator routes task to appropriate worker
3. **Content Analysis**: Worker analyzes task content and complexity
4. **Model Selection**: Worker selects optimal model (local or external)
5. **Context Retrieval**: RAG service provides relevant knowledge context
6. **Task Processing**: Worker processes task with selected model
7. **Result Delivery**: Result returned via MQTT to orchestrator

## Performance Optimization

### Resource Management

- **GPU Memory**: LRU cache management for local models
- **Connection Pooling**: Reused MQTT and HTTP connections
- **Request Batching**: Grouped requests for efficiency
- **Caching Strategy**: Multi-level caching for responses

### Scalability Features

- **Horizontal Scaling**: Add workers for increased throughput
- **Load Balancing**: Orchestrator distributes tasks evenly
- **Service Discovery**: Automatic discovery of available services
- **Graceful Degradation**: System continues with reduced capacity

## Error Handling

### Fail Fast, Fail Loud Implementation

- **Immediate Error Detection**: All errors caught and logged immediately
- **Clear Diagnostics**: Detailed error messages with context
- **Graceful Recovery**: Automatic retry and fallback mechanisms
- **Circuit Breakers**: Prevent cascading failures

### Error Categories

1. **Connection Errors**: MQTT, HTTP, database connection failures
2. **Model Errors**: Local model loading and inference failures
3. **Service Errors**: External AI service failures
4. **Resource Errors**: Memory, disk, or CPU resource exhaustion

## Monitoring and Observability

### Health Monitoring

- **Component Health**: Individual component status checking
- **System Health**: Overall system health aggregation
- **Dependency Health**: External service dependency monitoring
- **Resource Monitoring**: CPU, memory, disk, network usage

### Metrics Collection

- **Performance Metrics**: Response times, throughput, error rates
- **Resource Metrics**: Memory usage, GPU utilization, disk I/O
- **Business Metrics**: Task completion rates, model usage patterns
- **Custom Metrics**: Domain-specific measurements

### Logging Strategy

- **Structured Logging**: JSON-formatted logs with consistent fields
- **Log Levels**: DEBUG, INFO, WARN, ERROR with appropriate usage
- **Correlation IDs**: Request tracking across system boundaries
- **Log Aggregation**: Centralized log collection and analysis

## Security Considerations

### Authentication and Authorization

- **MQTT Authentication**: Username/password or certificate-based auth
- **API Key Management**: Secure storage and rotation of API keys
- **Access Control**: Role-based access to system components
- **Audit Logging**: Comprehensive security event logging

### Data Protection

- **Input Validation**: All inputs validated and sanitized
- **Encryption**: TLS for all external communications
- **Secure Storage**: Encrypted storage of sensitive data
- **Data Retention**: Appropriate data lifecycle management

## Development Standards

### Code Quality

- **Go Idioms**: Follow Go conventions and best practices
- **Error Handling**: Explicit error checking, never ignore errors
- **Interface Design**: Small, focused interfaces with clear contracts
- **Documentation**: Comprehensive godoc comments for all exported elements

### Testing Strategy

- **Unit Tests**: Individual component testing with high coverage
- **Integration Tests**: Component interaction testing
- **Performance Tests**: Load and stress testing
- **Security Tests**: Vulnerability and penetration testing

## Configuration Management

### Environment Variables

```bash
# Core system configuration
export LOCAL_MODELS_PATH="/data/models"
export PROJECT_ROOT="$(pwd)"
export QDRANT_STORAGE_PATH="/data/qdrant"
export RAG_DATA_PATH="/data/rag"

# API configuration (from ~/.claude/claude_helpers.toml)
export CEREBRAS_API_KEY="your-key"
export NVIDIA_API_KEY="your-key"
export GEMINI_API_KEY="your-key"
export GROK_API_KEY="your-key"
export GROQ_API_KEY="your-key"
```

### Configuration Files

- **`configs/models.yaml`**: Local model configuration
- **`configs/mcp.yaml`**: MCP service configuration
- **`~/.claude/claude_helpers.toml`**: External AI service configuration

## Troubleshooting

### Common Issues

1. **MQTT Connection Failures**: Check broker status and network connectivity
2. **Model Loading Errors**: Verify GPU memory and model file availability
3. **RAG Service Issues**: Check Qdrant connectivity and collection status
4. **API Key Problems**: Verify environment variables and API key validity

### Debug Commands

```bash
# Check MQTT connectivity
mosquitto_pub -h localhost -p 1883 -t test -m "hello"
mosquitto_sub -h localhost -p 1883 -t test

# Verify Qdrant status
curl http://localhost:6333/health

# Check GPU memory
nvidia-smi

# Monitor system logs
tail -f logs/*.log
```

## Future Enhancements

### Planned Features

- **Dynamic Worker Scaling**: Automatic worker provisioning based on load
- **Advanced Model Routing**: ML-based model selection optimization
- **Multi-Modal Support**: Enhanced image and document processing
- **Distributed Deployment**: Multi-node cluster support

### Integration Opportunities

- **Kubernetes**: Container orchestration integration
- **Service Mesh**: Advanced networking and security
- **Observability Stack**: Prometheus, Grafana, Jaeger integration
- **CI/CD Integration**: Automated testing and deployment

---

**Production Ready**: The internal system is designed for production deployment with comprehensive error handling, monitoring, and scalability features. Each component can operate independently while contributing to the overall system functionality.
