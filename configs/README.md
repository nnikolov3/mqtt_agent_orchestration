# Configuration Management (`configs/`)

## Overview

The `configs/` directory contains all configuration files for the MQTT Agent Orchestration System. This directory implements the **"Never hard code values"** principle by centralizing all configuration in external files, making the system environment-agnostic and easily configurable.

## Architecture Philosophy

Following our **"Excellence through Rigor"** philosophy, configuration management is:
- **Centralized**: All configuration in one location
- **Environment Agnostic**: Works across development, staging, production
- **Validated**: Configuration validation at startup
- **Secure**: Sensitive data handled appropriately

## Configuration Files

### 1. `models.yaml` - Local Model Configuration

**Purpose**: Configuration for local GGUF model management and optimization.

**Design Principles**:
- **Resource Management**: GPU memory optimization for RTX 3060
- **Model Selection**: Intelligent model routing based on task requirements
- **Performance Tuning**: Optimized parameters for each model
- **Fallback Strategy**: Graceful degradation when models unavailable

**Configuration Structure**:
```yaml
# Model Configuration for MQTT Agent Orchestration System
# Following "Never hard code values" principle - all model paths and settings centralized

models:
  qwen-omni:
    name: "Qwen2.5-Omni-3B"
    binary_path: "/usr/local/bin/llama-server"
    model_path: "/data/models/Qwen2.5-Omni-3B-Q8_0.gguf"
    type: "text"
    memory_limit: 3072  # 3GB for RTX 3060
    parameters:
      temperature: "0.8"
      max_tokens: "4096"
      context_length: "16384"
    specializations: ["general", "documentation", "code_generation", "text_analysis"]

  qwen-vl:
    name: "Qwen2.5-VL-7B"
    binary_path: "/usr/local/bin/llama-server"
    model_path: "/data/models/Qwen2.5-VL-7B-Abliterated-Caption-it.Q8_0.gguf"
    projector_path: "/data/models/Qwen2.5-VL-7B-Abliterated-Caption-it.mmproj-Q8_0.gguf"
    type: "multimodal"
    memory_limit: 4096  # 4GB
    parameters:
      temperature: "0.7"
      max_tokens: "2048"
      context_length: "8192"
    specializations: ["ui_analysis", "code_with_images", "error_screenshots", "multimodal_tasks"]

  qwen-embedding:
    name: "Qwen3-Embedding-4B"
    binary_path: "/usr/local/bin/llama-server"
    model_path: "./models/Qwen3-Embedding-4B-Q8_0.gguf"  # Project root for RAG access
    type: "embedding"
    memory_limit: 2048  # 2GB
    parameters:
      temperature: "0.1"
      max_tokens: "2560"  # Embedding dimension
      context_length: "8192"
    specializations: ["embeddings", "vector_generation"]

# Manager Configuration
manager:
  max_gpu_memory: 5632  # RTX 3060 5.5GB + 256MB buffer
  nvidia_smi_path: "/usr/bin/nvidia-smi"
  monitor_interval: "30s"
  lru_cache_size: 3  # Maximum concurrent models
  
# Fallback Configuration
fallback:
  enable_external_ai: true
  preferred_apis: ["cerebras", "nvidia", "gemini", "grok", "groq"]
  task_complexity_threshold: "medium"
  max_retries: 3
  retry_delay: "5s"

# Performance Settings
performance:
  enable_batching: true
  max_batch_size: 4
  enable_context_reuse: true
  max_context_age: "300s"
  cache_ttl: "600s"
```

**Usage Examples**:
```bash
# Load model configuration
./bin/client --config configs/models.yaml --list-models

# Use specific model
./bin/client --model qwen-omni --task "Generate documentation"

# Check model status
./bin/client --model-status qwen-vl
```

### 2. `mcp.yaml` - MCP Service Configuration

**Purpose**: Configuration for Model Context Protocol (MCP) services and tools.

**Design Principles**:
- **Service Discovery**: Automatic discovery of available MCP servers
- **Connection Management**: Efficient connection pooling and management
- **Security**: Secure communication with MCP services
- **Fallback**: Graceful degradation when services unavailable

**Configuration Structure**:
```yaml
# MCP Configuration for MQTT Agent Orchestration System
# Following "Never hard code values" principle - all MCP settings centralized

# Qdrant MCP Server Configuration
qdrant_mcp:
  server_path: "/usr/local/bin/mcp-server-qdrant"
  qdrant_url: "http://localhost:6333"
  timeout: "30s"
  env:
    QDRANT_URL: "http://localhost:6333"
    QDRANT_STORAGE_PATH: "/data/qdrant"
    QDRANT_API_KEY: ""
    LOG_LEVEL: "info"

# Local Models MCP Server Configuration
local_models_mcp:
  server_path: "/usr/local/bin/mcp-server-local-models"
  timeout: "60s"
  env:
    MODELS_PATH: "/data/models"
    GPU_MEMORY_LIMIT: "5632"
    LOG_LEVEL: "info"

# File System MCP Server Configuration
filesystem_mcp:
  server_path: "/usr/local/bin/mcp-server-filesystem"
  timeout: "30s"
  env:
    ROOT_PATH: "/data"
    LOG_LEVEL: "info"

# Git MCP Server Configuration
git_mcp:
  server_path: "/usr/local/bin/mcp-server-git"
  timeout: "30s"
  env:
    GIT_ROOT: "/data/projects"
    LOG_LEVEL: "info"

# MCP Client Configuration
client:
  timeout: "30s"
  max_retries: 3
  retry_delay: "5s"
  max_connections: 5
  connection_timeout: "10s"
  log_level: "info"
  enable_debug: false

# MCP Service Selection
services:
  rag:
    enabled: true
    type: "qdrant_mcp"
    priority: 1
  
  local_models:
    enabled: true
    type: "local_models_mcp"
    priority: 2
  
  filesystem:
    enabled: true
    type: "filesystem_mcp"
    priority: 3
  
  git:
    enabled: true
    type: "git_mcp"
    priority: 4

# Fallback Configuration
fallback:
  enable_direct_api: true
  enable_local_knowledge: true
  degradation_timeout: "5s"
  max_fallback_retries: 2

# Performance Configuration
performance:
  enable_pooling: true
  enable_caching: true
  cache_ttl: "300s"
  enable_batching: true
  batch_size: 10
  batch_timeout: "100ms"

# Security Configuration
security:
  enable_tls: false
  enable_auth: false
  auth_token: ""
  enable_rate_limiting: true
  requests_per_second: 100

# Monitoring Configuration
monitoring:
  enable_metrics: true
  enable_health_checks: true
  health_check_interval: "30s"
  enable_logging: true
  log_format: "json"
  enable_tracing: false
  trace_sampling_rate: 0.1
```

**Usage Examples**:
```bash
# Initialize MCP services
./bin/rag-service init --config configs/mcp.yaml

# Test MCP connectivity
./bin/client --test-mcp --config configs/mcp.yaml

# List available MCP tools
./bin/client --list-tools --config configs/mcp.yaml
```

## Configuration Management

### Environment-Specific Configuration

**Development Configuration**:
```yaml
# configs/development.yaml
environment: "development"
debug: true
log_level: "debug"

models:
  qwen-omni:
    memory_limit: 2048  # Lower memory for development
    parameters:
      temperature: "0.9"  # Higher creativity for development

mcp:
  client:
    timeout: "60s"  # Longer timeout for development
    enable_debug: true
```

**Production Configuration**:
```yaml
# configs/production.yaml
environment: "production"
debug: false
log_level: "info"

models:
  qwen-omni:
    memory_limit: 3072  # Full memory allocation
    parameters:
      temperature: "0.7"  # Conservative settings

mcp:
  client:
    timeout: "30s"  # Strict timeout for production
    enable_debug: false
  security:
    enable_tls: true
    enable_auth: true
```

### Configuration Validation

**Schema Validation**:
```yaml
# configs/schema.yaml
$schema: "http://json-schema.org/draft-07/schema#"
type: "object"
properties:
  models:
    type: "object"
    patternProperties:
      "^[a-zA-Z0-9_-]+$":
        type: "object"
        required: ["name", "type", "model_path"]
        properties:
          name:
            type: "string"
          type:
            enum: ["text", "multimodal", "embedding"]
          model_path:
            type: "string"
            pattern: "^[^<>:\"|?*]+$"
          memory_limit:
            type: "integer"
            minimum: 1024
            maximum: 8192
  manager:
    type: "object"
    required: ["max_gpu_memory"]
    properties:
      max_gpu_memory:
        type: "integer"
        minimum: 1024
        maximum: 32768
required: ["models", "manager"]
```

**Validation Script**:
```bash
#!/bin/bash
# scripts/validate_config.sh

validate_config() {
    local config_file="$1"
    local schema_file="$2"
    
    log_info "Validating configuration: $config_file"
    
    # Convert YAML to JSON for validation
    if ! yq eval -o=json "$config_file" | jq . > /dev/null; then
        error_exit "Invalid YAML syntax in $config_file"
    fi
    
    # Validate against schema
    if [[ -f "$schema_file" ]]; then
        if ! yq eval -o=json "$config_file" | jq -s ".[0]" | \
           python3 -m jsonschema -i - "$schema_file"; then
            error_exit "Configuration validation failed for $config_file"
        fi
    fi
    
    log_success "Configuration validated: $config_file"
}
```

### Configuration Loading

**Programmatic Loading**:
```go
// Load configuration from file
func LoadConfig(configPath string) (*Config, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %w", err)
    }
    
    var config Config
    if err := yaml.Unmarshal(data, &config); err != nil {
        return nil, fmt.Errorf("failed to parse config file: %w", err)
    }
    
    // Validate configuration
    if err := config.Validate(); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }
    
    return &config, nil
}

// Validate configuration
func (c *Config) Validate() error {
    if c.Manager.MaxGPUMemory <= 0 {
        return errors.New("max_gpu_memory must be positive")
    }
    
    for name, model := range c.Models {
        if err := model.Validate(); err != nil {
            return fmt.Errorf("model %s validation failed: %w", name, err)
        }
    }
    
    return nil
}
```

## Security Considerations

### Sensitive Data Handling

**Environment Variables**:
```bash
# Load sensitive data from environment
export QDRANT_API_KEY="your-api-key"
export CEREBRAS_API_KEY="your-cerebras-key"
export NVIDIA_API_KEY="your-nvidia-key"
export GEMINI_API_KEY="your-gemini-key"
export GROK_API_KEY="your-grok-key"
export GROQ_API_KEY="your-groq-key"
```

**Configuration Template**:
```yaml
# configs/template.yaml
qdrant_mcp:
  env:
    QDRANT_API_KEY: "${QDRANT_API_KEY}"  # Load from environment

models:
  qwen-omni:
    api_key: "${CEREBRAS_API_KEY}"  # Load from environment
```

**Secure Configuration Loading**:
```go
// Load configuration with environment variable substitution
func LoadConfigWithEnv(configPath string) (*Config, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, err
    }
    
    // Substitute environment variables
    expanded := os.Expand(string(data), func(key string) string {
        return os.Getenv(key)
    })
    
    var config Config
    if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

### Configuration Encryption

**Encrypted Configuration**:
```bash
# Encrypt sensitive configuration
gpg --encrypt --recipient user@example.com configs/secrets.yaml

# Decrypt configuration at runtime
gpg --decrypt configs/secrets.yaml.gpg | yq eval -o=json
```

## Configuration Best Practices

### Naming Conventions

- **File Names**: Use descriptive, lowercase names with hyphens
- **Environment Suffixes**: `configs/models.yaml`, `configs/models.prod.yaml`
- **Version Control**: Include version in configuration structure
- **Documentation**: Include comments explaining configuration options

### Configuration Organization

```
configs/
├── models.yaml              # Model configuration
├── mcp.yaml                 # MCP service configuration
├── development.yaml         # Development environment overrides
├── production.yaml          # Production environment overrides
├── schema.yaml              # Configuration schema
├── template.yaml            # Configuration template
└── secrets.yaml.gpg         # Encrypted secrets
```

### Configuration Validation

**Startup Validation**:
```go
// Validate configuration at startup
func (app *Application) validateConfiguration() error {
    // Validate model configuration
    for name, model := range app.config.Models {
        if err := app.validateModelConfig(name, model); err != nil {
            return fmt.Errorf("model %s: %w", name, err)
        }
    }
    
    // Validate MCP configuration
    if err := app.validateMCPConfig(); err != nil {
        return fmt.Errorf("MCP configuration: %w", err)
    }
    
    // Validate paths and permissions
    if err := app.validatePaths(); err != nil {
        return fmt.Errorf("path validation: %w", err)
    }
    
    return nil
}
```

## Monitoring and Observability

### Configuration Metrics

```go
// Track configuration usage
type ConfigMetrics struct {
    LoadTime     time.Duration
    ValidationTime time.Duration
    Errors       int
    Warnings     int
}

func (app *Application) trackConfigMetrics(metrics ConfigMetrics) {
    app.metrics.ConfigLoadTime.Observe(metrics.LoadTime.Seconds())
    app.metrics.ConfigValidationTime.Observe(metrics.ValidationTime.Seconds())
    app.metrics.ConfigErrors.Inc()
}
```

### Configuration Health Checks

```go
// Health check for configuration
func (app *Application) configHealthCheck() error {
    // Check if configuration is loaded
    if app.config == nil {
        return errors.New("configuration not loaded")
    }
    
    // Check if required models are available
    for name, model := range app.config.Models {
        if !app.isModelAvailable(name) {
            return fmt.Errorf("model %s not available", name)
        }
    }
    
    // Check MCP service connectivity
    if err := app.checkMCPConnectivity(); err != nil {
        return fmt.Errorf("MCP connectivity: %w", err)
    }
    
    return nil
}
```

## Troubleshooting

### Common Configuration Issues

1. **Invalid YAML Syntax**: Use YAML validators to check syntax
2. **Missing Environment Variables**: Verify all required environment variables are set
3. **Path Issues**: Ensure all file paths are correct and accessible
4. **Permission Issues**: Check file permissions for configuration files

### Debug Configuration

```bash
# Enable configuration debugging
export CONFIG_DEBUG=true

# Validate configuration
./bin/client --validate-config configs/models.yaml

# Show effective configuration
./bin/client --show-config --config configs/models.yaml

# Test configuration loading
./bin/client --test-config configs/models.yaml
```

### Configuration Recovery

```bash
# Backup configuration
cp configs/models.yaml configs/models.yaml.backup

# Restore from backup
cp configs/models.yaml.backup configs/models.yaml

# Reset to defaults
./scripts/reset_config.sh --config-type models
```

## Future Enhancements

### Planned Features

- **Configuration Hot Reload**: Reload configuration without restart
- **Configuration Versioning**: Version control for configuration changes
- **Configuration Templates**: Pre-built templates for common scenarios
- **Configuration Migration**: Automated migration between versions

### Extension Points

- **Custom Validators**: User-defined configuration validation rules
- **Configuration Providers**: Support for external configuration sources
- **Configuration Encryption**: Enhanced encryption for sensitive data
- **Configuration Analytics**: Analysis of configuration usage patterns

---

**Production Ready**: Configuration management is designed for production use with comprehensive validation, security features, and monitoring capabilities. It provides a robust foundation for managing system configuration across different environments.
