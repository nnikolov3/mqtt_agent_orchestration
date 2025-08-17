# Architecture Overview

## System Type

The MQTT Agent Orchestration System is a **backend-only system** that provides AI agent orchestration through:

- **HTTP RESTful API**: For external system integration  
- **Command-Line Interface (CLI)**: For system interaction and management
- **MQTT Message Bus**: For internal component communication

## No Web UI Components

This system **does not include** any web UI, frontend, or browser-based components. All interaction is through:

1. **HTTP API Endpoints** (Port 8080)
   - `/health` - System health monitoring
   - `/workflow/submit` - Task submission
   - `/workflow/status` - Workflow status checking

2. **Command-Line Tools**
   - `./bin/client` - Main interaction tool
   - `./bin/ops` - Operations management
   - `./bin/rag-service` - RAG management

3. **MQTT Topics** for real-time monitoring
   - `tasks/workflow/+` - Task distribution
   - `results/workflow/+` - Result aggregation

## System Components

### Core Services (Backend Only)
- **Orchestrator**: Workflow coordination engine
- **Role Workers**: Specialized AI processing agents
- **RAG Service**: Knowledge management system
- **Embedding Worker**: Local model inference

### Integration Points
- **External AI APIs**: Cerebras, NVIDIA, Gemini, Grok, Groq
- **Vector Database**: Qdrant for knowledge storage
- **MQTT Broker**: Message-based communication

## Design Philosophy

Following the **"Do more with less"** principle, the system focuses on:
- **API-First Design**: All functionality exposed through APIs
- **CLI Automation**: Comprehensive command-line tools
- **Headless Operation**: Designed for server environments
- **Integration Ready**: Easy integration with existing systems

## Future UI Considerations

While the current system is backend-only, it's designed to support future UI development:
- **RESTful API**: Ready for frontend consumption
- **WebSocket Support**: Can be added for real-time updates
- **OpenAPI Specification**: Can be generated for API documentation
- **Event-Driven Architecture**: Supports reactive UI patterns

---

For detailed component documentation, see the README files in each directory.