# Data Storage (`storage/`)

## Overview

The `storage/` directory contains the persistent data storage for the MQTT Agent Orchestration System. This directory implements the **"Never hard code values"** principle by providing a centralized, organized structure for all system data, following the global data structure defined in `docs/DATA_STRUCTURE.md`.

## Architecture Philosophy

Following our **"Excellence through Rigor"** philosophy, data storage is:
- **Organized**: Clear structure for different types of data
- **Persistent**: Reliable storage across system restarts
- **Scalable**: Designed to handle growth in data volume
- **Secure**: Appropriate access controls and data protection

## Directory Structure

```
storage/
├── aliases/                    # System aliases and shortcuts
│   └── data.json              # Alias definitions and mappings
├── collections/               # Qdrant vector database collections
│   ├── ai_conversations/      # AI conversation history
│   ├── ai_prompts/           # System prompts for each role
│   ├── api_documentation/    # API documentation and examples
│   ├── best_practices/       # Coding and system best practices
│   ├── claude_rag/           # Claude-specific RAG data
│   ├── code_examples/        # Code examples and patterns
│   ├── coding_standards/     # Language-specific coding standards
│   ├── error_patterns/       # Error patterns and solutions
│   ├── git_history/          # Git repository history and metadata
│   ├── model_performance/    # Model performance metrics and data
│   ├── project_documentation/ # Project documentation and guides
│   ├── project_knowledge/    # Project-specific knowledge base
│   ├── successful_patterns/  # Successful implementation patterns
│   └── system_prompts/       # System prompt templates
└── raft_state.json           # Raft consensus state (if applicable)
```

## Qdrant Collections

### Collection Architecture

Each Qdrant collection follows a consistent structure optimized for the Qwen3-Embedding-4B model (2560-dimensional vectors):

```yaml
# Collection Configuration
vector_dimension: 2560        # Qwen3-Embedding-4B dimension
distance_metric: "cosine"     # Cosine similarity for semantic search
index_type: "hnsw"           # Hierarchical Navigable Small World
payload_indexing: true       # Enable payload filtering
```

### Collection Details

#### 1. `ai_conversations/` - Conversation History

**Purpose**: Store AI conversation history for context and learning.

**Data Structure**:
```json
{
  "id": "conv_123",
  "vectors": [0.1, 0.2, ...],  // 2560-dimensional embedding
  "payload": {
    "conversation_id": "conv_123",
    "user_id": "user_456",
    "role": "developer",
    "content": "How do I implement a REST API?",
    "response": "Here's how to implement a REST API...",
    "timestamp": "2024-01-15T10:30:00Z",
    "metadata": {
      "task_type": "code_generation",
      "complexity": "medium",
      "language": "go"
    }
  }
}
```

**Usage**:
```bash
# Search conversation history
./bin/rag-service search \
  --collection ai_conversations \
  --query "REST API implementation" \
  --limit 5

# Add conversation to history
./bin/rag-service add-conversation \
  --user-id user_456 \
  --role developer \
  --content "How do I implement a REST API?" \
  --response "Here's how to implement a REST API..."
```

#### 2. `ai_prompts/` - System Prompts

**Purpose**: Store system prompts for each worker role.

**Data Structure**:
```json
{
  "id": "prompt_dev_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "role": "developer",
    "prompt": "You are an expert Go developer focused on clean, idiomatic code...",
    "version": "1.0.0",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z",
    "metadata": {
      "specialization": "backend_development",
      "experience_level": "senior",
      "focus_areas": ["api_design", "performance", "security"]
    }
  }
}
```

**Usage**:
```bash
# Store system prompt for role
./bin/rag-service store-prompt \
  --role developer \
  --prompt "You are an expert Go developer..."

# Retrieve system prompt
./bin/rag-service get-prompt \
  --role developer \
  --specialization backend_development
```

#### 3. `api_documentation/` - API Documentation

**Purpose**: Store API documentation and examples for reference.

**Data Structure**:
```json
{
  "id": "api_rest_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "title": "REST API Design Best Practices",
    "content": "REST APIs should follow these principles...",
    "api_type": "rest",
    "language": "go",
    "version": "1.0.0",
    "tags": ["api", "rest", "best_practices"],
    "examples": [
      {
        "method": "GET",
        "endpoint": "/api/v1/users",
        "code": "func GetUsers(w http.ResponseWriter, r *http.Request) {...}"
      }
    ],
    "metadata": {
      "difficulty": "intermediate",
      "category": "backend",
      "last_updated": "2024-01-15T10:30:00Z"
    }
  }
}
```

**Usage**:
```bash
# Search API documentation
./bin/rag-service search \
  --collection api_documentation \
  --query "REST API authentication" \
  --limit 3

# Add API documentation
./bin/rag-service add-document \
  --collection api_documentation \
  --title "JWT Authentication" \
  --content "JWT authentication implementation..." \
  --tags "authentication,jwt,security"
```

#### 4. `best_practices/` - Best Practices

**Purpose**: Store coding and system best practices.

**Data Structure**:
```json
{
  "id": "bp_go_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "title": "Go Error Handling Best Practices",
    "content": "In Go, always check errors explicitly...",
    "category": "error_handling",
    "language": "go",
    "difficulty": "beginner",
    "tags": ["go", "error_handling", "best_practices"],
    "examples": [
      {
        "description": "Proper error checking",
        "code": "if err != nil { return err }",
        "explanation": "Always check errors immediately after function calls"
      }
    ],
    "metadata": {
      "author": "Go Team",
      "source": "official_documentation",
      "rating": 4.8
    }
  }
}
```

**Usage**:
```bash
# Search best practices
./bin/rag-service search \
  --collection best_practices \
  --query "Go error handling" \
  --limit 5

# Add best practice
./bin/rag-service add-best-practice \
  --title "Go Concurrency Patterns" \
  --content "Use channels for communication..." \
  --category concurrency \
  --language go
```

#### 5. `claude_rag/` - Claude-Specific RAG Data

**Purpose**: Store Claude-specific RAG data and knowledge.

**Data Structure**:
```json
{
  "id": "claude_rag_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "content": "Claude-specific knowledge and patterns...",
    "source": "claude_conversations",
    "context": "system_optimization",
    "confidence": 0.95,
    "metadata": {
      "claude_version": "3.5-sonnet",
      "conversation_id": "conv_123",
      "timestamp": "2024-01-15T10:30:00Z"
    }
  }
}
```

#### 6. `code_examples/` - Code Examples

**Purpose**: Store code examples and patterns for reference.

**Data Structure**:
```json
{
  "id": "example_go_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "title": "HTTP Server with Middleware",
    "description": "Complete example of HTTP server with middleware",
    "language": "go",
    "code": "package main\n\nimport (\n  \"net/http\"\n  \"log\"\n)\n\nfunc main() {\n  // Implementation...\n}",
    "tags": ["http", "middleware", "server"],
    "difficulty": "intermediate",
    "metadata": {
      "author": "system",
      "tested": true,
      "performance_notes": "Optimized for production use"
    }
  }
}
```

**Usage**:
```bash
# Search code examples
./bin/rag-service search \
  --collection code_examples \
  --query "HTTP server middleware" \
  --language go \
  --limit 3

# Add code example
./bin/rag-service add-example \
  --title "Database Connection Pool" \
  --language go \
  --code "package main\n\n// Implementation..." \
  --tags "database,connection_pool"
```

#### 7. `coding_standards/` - Coding Standards

**Purpose**: Store language-specific coding standards and conventions.

**Data Structure**:
```json
{
  "id": "std_go_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "language": "go",
    "standard": "Go Code Review Comments",
    "rule": "Package names should be short and clear",
    "description": "Package names should be short, clear, and avoid underscores",
    "examples": [
      {
        "good": "package user",
        "bad": "package user_management"
      }
    ],
    "severity": "warning",
    "category": "naming",
    "metadata": {
      "source": "golang.org/doc/effective_go.html",
      "enforced": true
    }
  }
}
```

#### 8. `error_patterns/` - Error Patterns

**Purpose**: Store common error patterns and their solutions.

**Data Structure**:
```json
{
  "id": "error_go_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "error_pattern": "undefined: function_name",
    "language": "go",
    "description": "Function is not defined or not imported",
    "causes": [
      "Function name is misspelled",
      "Function is not imported",
      "Function is in different package"
    ],
    "solutions": [
      "Check function name spelling",
      "Add import statement",
      "Use correct package name"
    ],
    "examples": [
      {
        "error": "undefined: fmt.Println",
        "solution": "import \"fmt\""
      }
    ],
    "metadata": {
      "frequency": "high",
      "difficulty": "beginner"
    }
  }
}
```

#### 9. `git_history/` - Git Repository History

**Purpose**: Store Git repository history and metadata.

**Data Structure**:
```json
{
  "id": "git_commit_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "commit_hash": "abc123def456",
    "author": "developer@example.com",
    "message": "Add HTTP server implementation",
    "files_changed": ["server.go", "main.go"],
    "changes": [
      {
        "file": "server.go",
        "type": "added",
        "lines": "+50 -0"
      }
    ],
    "timestamp": "2024-01-15T10:30:00Z",
    "branch": "main",
    "metadata": {
      "review_status": "approved",
      "test_coverage": 0.85
    }
  }
}
```

#### 10. `model_performance/` - Model Performance Data

**Purpose**: Store model performance metrics and optimization data.

**Data Structure**:
```json
{
  "id": "perf_qwen_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "model_name": "qwen-omni-3b",
    "task_type": "code_generation",
    "metrics": {
      "accuracy": 0.92,
      "latency_ms": 1500,
      "memory_usage_mb": 3072,
      "throughput": 10.5
    },
    "parameters": {
      "temperature": 0.7,
      "max_tokens": 2048,
      "context_length": 8192
    },
    "timestamp": "2024-01-15T10:30:00Z",
    "metadata": {
      "hardware": "RTX 3060",
      "optimization_level": "Q8_0"
    }
  }
}
```

#### 11. `project_documentation/` - Project Documentation

**Purpose**: Store project-specific documentation and guides.

**Data Structure**:
```json
{
  "id": "doc_arch_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "title": "System Architecture Overview",
    "content": "The MQTT Agent Orchestration System consists of...",
    "category": "architecture",
    "version": "1.0.0",
    "tags": ["architecture", "overview", "system_design"],
    "sections": [
      {
        "title": "Components",
        "content": "The system has the following components..."
      }
    ],
    "metadata": {
      "author": "system_architect",
      "last_updated": "2024-01-15T10:30:00Z",
      "review_status": "approved"
    }
  }
}
```

#### 12. `project_knowledge/` - Project Knowledge Base

**Purpose**: Store project-specific knowledge and insights.

**Data Structure**:
```json
{
  "id": "knowledge_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "topic": "MQTT Performance Optimization",
    "content": "MQTT performance can be optimized by...",
    "category": "performance",
    "insights": [
      "Use QoS 1 for guaranteed delivery",
      "Implement connection pooling",
      "Monitor message throughput"
    ],
    "references": [
      "MQTT Specification v5.0",
      "Performance Testing Results"
    ],
    "metadata": {
      "confidence": 0.95,
      "source": "performance_analysis",
      "last_updated": "2024-01-15T10:30:00Z"
    }
  }
}
```

#### 13. `successful_patterns/` - Successful Patterns

**Purpose**: Store successful implementation patterns and solutions.

**Data Structure**:
```json
{
  "id": "pattern_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "pattern_name": "Circuit Breaker Pattern",
    "description": "Implementation of circuit breaker for external API calls",
    "category": "resilience",
    "language": "go",
    "implementation": "type CircuitBreaker struct {...}",
    "usage_examples": [
      {
        "scenario": "External API calls",
        "code": "breaker := NewCircuitBreaker()"
      }
    ],
    "success_metrics": {
      "reliability": 0.99,
      "performance": "improved",
      "adoption_rate": 0.85
    },
    "metadata": {
      "author": "senior_developer",
      "review_count": 5,
      "rating": 4.9
    }
  }
}
```

#### 14. `system_prompts/` - System Prompt Templates

**Purpose**: Store system prompt templates for different scenarios.

**Data Structure**:
```json
{
  "id": "prompt_template_001",
  "vectors": [0.1, 0.2, ...],
  "payload": {
    "template_name": "Code Review Template",
    "description": "Template for code review system prompts",
    "template": "You are an expert code reviewer. Review the following code for...",
    "variables": [
      "language",
      "complexity",
      "focus_areas"
    ],
    "usage_context": "code_review",
    "effectiveness_score": 0.92,
    "metadata": {
      "created_by": "system",
      "usage_count": 150,
      "last_updated": "2024-01-15T10:30:00Z"
    }
  }
}
```

## Data Management

### Collection Operations

**Initialize Collections**:
```bash
# Initialize all collections
./bin/rag-service init

# Initialize specific collection
./bin/rag-service init --collection coding_standards

# Initialize with custom configuration
./bin/rag-service init --config custom-config.yaml
```

**Search Operations**:
```bash
# Search across all collections
./bin/rag-service search --query "error handling" --limit 10

# Search specific collection
./bin/rag-service search \
  --collection coding_standards \
  --query "Go error handling" \
  --limit 5

# Search with filters
./bin/rag-service search \
  --collection code_examples \
  --query "HTTP server" \
  --filter '{"language": "go", "difficulty": "intermediate"}' \
  --limit 3
```

**Data Management**:
```bash
# Add document to collection
./bin/rag-service add-document \
  --collection best_practices \
  --content "Always check errors in Go" \
  --metadata '{"language": "go", "category": "error_handling"}'

# Update document
./bin/rag-service update-document \
  --collection api_documentation \
  --id doc_123 \
  --content "Updated API documentation..."

# Delete document
./bin/rag-service delete-document \
  --collection code_examples \
  --id example_123
```

### Backup and Recovery

**Backup Strategy**:
```bash
# Create backup of all collections
./bin/rag-service backup --output backup_$(date +%Y%m%d).tar.gz

# Backup specific collection
./bin/rag-service backup \
  --collection coding_standards \
  --output coding_standards_backup.tar.gz

# Restore from backup
./bin/rag-service restore --input backup_20240115.tar.gz
```

**Data Migration**:
```bash
# Export collection data
./bin/rag-service export \
  --collection best_practices \
  --format json \
  --output best_practices.json

# Import collection data
./bin/rag-service import \
  --collection best_practices \
  --format json \
  --input best_practices.json
```

## Performance Optimization

### Indexing Strategy

**Vector Indexing**:
```yaml
# HNSW index configuration for optimal performance
index_type: "hnsw"
parameters:
  m: 16                    # Number of connections per layer
  ef_construct: 100        # Search accuracy during construction
  ef_search: 50           # Search accuracy during queries
  full_scan_threshold: 10000
```

**Payload Indexing**:
```yaml
# Enable payload indexing for efficient filtering
payload_indexing:
  language: true
  category: true
  difficulty: true
  tags: true
```

### Query Optimization

**Efficient Queries**:
```bash
# Use specific filters to reduce search space
./bin/rag-service search \
  --collection code_examples \
  --query "HTTP server" \
  --filter '{"language": "go"}' \
  --limit 5

# Use approximate search for large collections
./bin/rag-service search \
  --collection ai_conversations \
  --query "error handling" \
  --approximate \
  --limit 10
```

## Monitoring and Maintenance

### Health Monitoring

**Collection Health**:
```bash
# Check collection health
./bin/rag-service health --collection coding_standards

# Monitor collection size
./bin/rag-service stats --collection coding_standards

# Check vector quality
./bin/rag-service quality --collection coding_standards
```

**Performance Monitoring**:
```bash
# Monitor query performance
./bin/rag-service monitor --collection coding_standards

# Track search accuracy
./bin/rag-service accuracy --collection coding_standards

# Monitor memory usage
./bin/rag-service memory --collection coding_standards
```

### Maintenance Operations

**Data Cleanup**:
```bash
# Remove duplicate documents
./bin/rag-service deduplicate --collection code_examples

# Clean up old data
./bin/rag-service cleanup \
  --collection ai_conversations \
  --older-than 30d

# Optimize collection
./bin/rag-service optimize --collection coding_standards
```

## Security Considerations

### Access Control

**Collection Permissions**:
```yaml
# Collection access control
permissions:
  ai_conversations:
    read: ["authenticated_users"]
    write: ["system", "authenticated_users"]
    admin: ["system_admin"]
  
  coding_standards:
    read: ["authenticated_users"]
    write: ["senior_developers", "system"]
    admin: ["system_admin"]
```

### Data Protection

**Encryption**:
```yaml
# Enable encryption for sensitive collections
encryption:
  enabled: true
  algorithm: "AES-256-GCM"
  key_rotation: "30d"
```

## Troubleshooting

### Common Issues

1. **Collection Not Found**: Verify collection exists and is initialized
2. **Search Performance**: Check index configuration and query optimization
3. **Memory Issues**: Monitor collection size and optimize storage
4. **Data Corruption**: Use backup and recovery procedures

### Debug Commands

```bash
# Debug collection issues
./bin/rag-service debug --collection coding_standards

# Check collection integrity
./bin/rag-service integrity --collection coding_standards

# Validate data format
./bin/rag-service validate --collection coding_standards
```

## Future Enhancements

### Planned Features

- **Real-time Indexing**: Automatic indexing of new data
- **Advanced Analytics**: Deep insights into collection usage
- **Multi-language Support**: Support for multiple languages
- **Federated Search**: Search across multiple collections

### Extension Points

- **Custom Collections**: User-defined collection types
- **Advanced Filtering**: Complex query filters
- **Data Versioning**: Version control for collection data
- **API Integration**: REST API for collection management

---

**Production Ready**: The storage system is designed for production use with comprehensive backup, monitoring, and security features. It provides a robust foundation for managing the system's knowledge base and data persistence.
