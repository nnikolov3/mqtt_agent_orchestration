# MQTT Agent Orchestration System

![Go Version](https://img.shields.io/badge/go-1.24+-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)

## Overview

The **MQTT Agent Orchestration System** is a lightweight, efficient Go-based framework for managing autonomous, role-based AI agents that communicate via MQTT. The system integrates with Qdrant vector database for RAG (Retrieval-Augmented Generation) capabilities, enabling intelligent, context-aware task processing.

**Key Design Philosophy:**
- **Token efficiency**: RAG-based context reduces AI API token consumption
- **Speed optimization**: Local knowledge retrieval accelerates development
- **Quality enhancement**: Context-aware agents produce better results  
- **User-level service**: Reusable across multiple projects

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Orchestrator  │    │   MQTT Broker   │    │  Qdrant Vector  │
│   (Workflow     │◀──▶│   (Mosquitto)   │    │     Database    │
│    Engine)      │    │                 │    │  (RAG Context)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                        │                        ▲
         ▼                        ▼                        │
┌─────────────────────────────────────────────────────────┼─────────┐
│                    MQTT Topics                          │         │
│ • tasks/workflow/{stage}     • results/workflow/{stage} │         │
│ • workers/status/{role}/{id} • orchestrator/workflow    │         │
└─────────────────────────────────────────────────────────┼─────────┘
         │                        │                        │
         ▼                        ▼                        │
┌─────────────────┐    ┌─────────────────┐    ┌───────────┼─────────┐
│  Developer      │    │   Reviewer      │    │  Approver │         │
│   Worker        │    │    Worker       │    │   Worker  │         │
│                 │    │                 │    │           │         │
│ • Code gen      │    │ • Code review   │    │ • Quality │         │
│ • RAG context   │    │ • RAG context   │    │ • RAG ctx │         │
└─────────────────┘    └─────────────────┘    └───────────┼─────────┘
                                                          │
┌─────────────────┐                                      │
│    Tester       │                                      │
│    Worker       │                                      │
│                 │                                      │
│ • Testing       │──────────────────────────────────────┘
│ • Validation    │
│ • RAG context   │
└─────────────────┘
```

## Features

### Core Capabilities
- **Role-Based Workers**: Specialized agents (Developer, Reviewer, Approver, Tester)
- **Autonomous Workflows**: Multi-stage document/code generation with automatic retries
- **MQTT Communication**: Lightweight, asynchronous messaging via Eclipse Paho
- **RAG Integration**: Qdrant vector database for context-aware processing
- **Token Optimization**: Reduce AI API costs through intelligent context retrieval
- **Cross-Project Reusability**: User-level service design

### RAG-Enhanced Intelligence
- **System Prompts**: Each agent role has specialized prompts stored in Qdrant
- **Context Retrieval**: Relevant coding standards, examples, and patterns
- **Quality Improvement**: Context-aware responses improve output quality
- **Knowledge Base**: Stores coding standards, best practices, project patterns

### Workflow Automation
- **Autonomous Operation**: End-to-end processing without manual intervention
- **Quality Gates**: Built-in review and approval stages
- **Arbitration Protocol**: Prevents infinite loops (15-transition limit)
- **Error Recovery**: Automatic retries with feedback incorporation

## Quick Start

### Prerequisites

- **Go 1.24+**
- **Mosquitto MQTT Broker** (local or remote)
- **Qdrant Vector Database** (optional but recommended for RAG features)

### Installation

1. **Clone and build:**
```bash
git clone <repository-url>
cd mqtt_agent_orchestration
./scripts/build.sh
```

2. **Start MQTT broker:**
```bash
# On Linux/macOS with systemd
sudo systemctl start mosquitto

# Or run manually
mosquitto -p 1883
```

3. **Start Qdrant (optional):**
```bash
# Using Docker (lightweight option)
docker run -p 6333:6333 qdrant/qdrant

# Or install locally
# See: https://qdrant.tech/documentation/quickstart/
```

4. **Launch the autonomous system:**
```bash
./start_autonomous_system.sh
```

This starts:
- Workflow orchestrator
- Four role-specific workers (developer, reviewer, approver, tester)
- Health monitoring and logging

### Basic Usage

**Create a Go coding standards document:**
```bash
./bin/client --doc-type go_coding_standards --output GO_CODING_STANDARD_CLAUDE.md
```

**List available document types:**
```bash
./bin/client --list
```

**Monitor system:**
```bash
# Watch logs
tail -f logs/*.log

# Check system status
ps aux | grep -E "(orchestrator|role-worker)"
```

## Project Structure

```
.
├── cmd/                    # Application entry points
│   ├── orchestrator/       # Workflow orchestrator
│   ├── role-worker/        # Specialized role workers
│   ├── worker/            # Generic worker implementation
│   ├── client/            # Command-line client
│   └── server/            # Server components
├── internal/              # Private application code
│   ├── mqtt/              # MQTT client implementation
│   ├── rag/               # RAG service integration
│   ├── worker/            # Worker logic and processing
│   ├── orchestrator/      # Workflow management
│   └── config/            # Configuration management
├── pkg/                   # Public API packages
│   └── types/             # Shared type definitions
├── scripts/               # Build and deployment scripts
│   ├── build.sh          # Production build script
│   └── run.sh            # Runtime orchestration
├── bin/                   # Compiled binaries (created by build)
├── logs/                  # Runtime logs (created by system)
└── start_autonomous_system.sh  # System startup script
```

## Configuration

### MQTT Configuration
- **Host**: localhost (default)
- **Port**: 1883 (default)
- **Topics**: Predefined topic structure for workflow coordination

### Qdrant Configuration  
- **URL**: http://localhost:6333 (default)
- **Collection**: Automatically created for agent system prompts
- **Embeddings**: Context vectors for role-specific knowledge

### Worker Roles
- **Developer**: Generates initial code/documents
- **Reviewer**: Reviews and improves content quality
- **Approver**: Makes final approval decisions
- **Tester**: Validates output structure and correctness

## System Design

### Token Efficiency Strategy
1. **RAG Context**: Store coding patterns, standards, examples in Qdrant
2. **Smart Retrieval**: Query relevant context before AI API calls
3. **Prompt Optimization**: Pre-populated context reduces token usage
4. **Caching Strategy**: Reuse similar contexts across tasks

### Quality Enhancement
1. **Role Specialization**: Each worker has domain-specific knowledge
2. **Iterative Improvement**: Feedback loops between stages
3. **Context Awareness**: Decisions based on project-specific patterns
4. **Standards Enforcement**: Automatic adherence to coding standards

### Cross-Project Reusability
1. **User-Level Service**: Deploy once, use in multiple projects
2. **Knowledge Portability**: Export/import knowledge bases
3. **Configurable Workflows**: Adapt to different project types
4. **API Interface**: Programmatic access for integration

## Development

### Building
```bash
# Full build with tests and linting
./scripts/build.sh

# Verbose build
./scripts/build.sh --verbose

# Clean and rebuild
./scripts/build.sh --clean
```

### Testing
```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/mqtt
```

### Code Quality
```bash
# Run linter (if available)
golangci-lint run

# Format code
go fmt ./...

# Vet code
go vet ./...
```

## Troubleshooting

### Common Issues

**MQTT Connection Failed:**
- Verify Mosquitto is running: `systemctl status mosquitto`
- Check port availability: `netstat -ln | grep 1883`
- Verify firewall settings

**Workers Not Starting:**
- Check logs in `logs/` directory
- Verify binaries exist in `bin/` directory  
- Run `./scripts/build.sh` if binaries missing

**RAG Features Not Working:**
- Verify Qdrant is accessible: `curl http://localhost:6333/health`
- Check Qdrant logs for connection issues
- RAG features degrade gracefully if Qdrant unavailable

**System Health Issues:**
- Monitor worker processes: `ps aux | grep role-worker`
- Check system resources: `top` or `htop`
- Review orchestrator logs: `tail -f logs/orchestrator.log`

## License

MIT License - see LICENSE file for details.

## Contributing

1. Follow Go best practices and project coding standards
2. Add tests for new functionality
3. Update documentation for user-facing changes
4. Verify system works end-to-end before submitting

## Roadmap

- [ ] Enhanced RAG knowledge management
- [ ] Multi-project workspace support  
- [ ] Web UI for workflow monitoring
- [ ] Advanced analytics and metrics
- [ ] Plugin architecture for custom agents