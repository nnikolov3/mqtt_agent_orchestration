# Automation Scripts (`scripts/`)

## Overview

The `scripts/` directory contains automation scripts that follow the **"Do more with less"** principle by automating repetitive tasks and ensuring consistent system operations. All scripts implement **"Fail Fast, Fail Loud"** error handling and follow strict BASH scripting standards.

## Architecture Philosophy

Following our **"Excellence through Rigor"** philosophy, all scripts are:
- **Reliable**: Comprehensive error handling and validation
- **Idempotent**: Safe to run multiple times without side effects
- **Observable**: Detailed logging and status reporting
- **Maintainable**: Clear structure and documentation

## Script Categories

### 1. Build and Deployment Scripts

#### `build.sh` - System Build Automation

**Purpose**: Automated build process for all system components.

**Design Principles**:
- **Reproducible Builds**: Consistent builds across environments
- **Dependency Validation**: Verify all required dependencies
- **Quality Gates**: Run tests and linting before build
- **Artifact Management**: Proper artifact organization and versioning

**Key Features**:
- Multi-component build orchestration
- Dependency checking and validation
- Test execution and coverage reporting
- Linting and code quality checks
- Version tagging and artifact management
- Cross-platform build support

**Usage**:
```bash
# Standard build
./scripts/build.sh

# Clean build (remove previous artifacts)
./scripts/build.sh --clean

# Build with specific version
./scripts/build.sh --version v1.2.3

# Build with debug information
./scripts/build.sh --debug
```

**Configuration**:
```bash
# Build configuration (set in environment or script)
export BUILD_VERSION="$(git describe --tags --always --dirty)"
export BUILD_TIME="$(date -u '+%Y-%m-%d_%H:%M:%S')"
export BUILD_COMMIT="$(git rev-parse HEAD)"
export GO_VERSION="1.24"
```

#### `run.sh` - System Execution Script

**Purpose**: Automated system startup and orchestration.

**Design Principles**:
- **Service Orchestration**: Start all components in correct order
- **Health Monitoring**: Verify all services are healthy
- **Graceful Shutdown**: Proper cleanup on termination
- **Configuration Management**: Environment-specific configuration

**Key Features**:
- Dependency service startup (MQTT broker, Qdrant)
- Component health verification
- Log aggregation and monitoring
- Signal handling for graceful shutdown
- Environment-specific configuration

**Usage**:
```bash
# Start all services
./scripts/run.sh

# Start specific components
./scripts/run.sh --components orchestrator,worker

# Start with custom configuration
./scripts/run.sh --config production.yaml

# Start in development mode
./scripts/run.sh --dev
```

### 2. Development and Quality Scripts

#### `lint.sh` - Code Quality Automation

**Purpose**: Automated code quality checks and standards enforcement.

**Design Principles**:
- **Quality Gates**: Enforce coding standards automatically
- **Comprehensive Coverage**: Check all aspects of code quality
- **Fast Feedback**: Quick execution for development workflow
- **Configurable**: Adaptable to different project requirements

**Key Features**:
- Go linting with golangci-lint
- Shell script linting with shellcheck
- YAML/TOML validation
- Security scanning
- Performance analysis
- Documentation validation

**Usage**:
```bash
# Run all linting checks
./scripts/lint.sh

# Run specific checks
./scripts/lint.sh --go-only

# Run with auto-fix where possible
./scripts/lint.sh --fix

# Run with custom configuration
./scripts/lint.sh --config .golangci-custom.yml
```

**Configuration**:
```yaml
# .golangci.yml configuration
run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  enable:
    - gofmt
    - golint
    - govet
    - errcheck
    - staticcheck
    - gosimple
    - ineffassign
    - unused
    - misspell
    - gosec

linters-settings:
  gosec:
    excludes:
      - G404 # Use of weak random number generator
```

#### `fix_bash_standards.sh` - BASH Standards Enforcement

**Purpose**: Enforce BASH scripting standards across all shell scripts.

**Design Principles**:
- **Strict Mode**: Always use `set -euo pipefail`
- **Variable Safety**: Proper variable declaration and quoting
- **Error Handling**: Comprehensive error capture and reporting
- **Portability**: Ensure scripts work across different environments

**Key Features**:
- Automatic strict mode enforcement
- Variable declaration standardization
- Error handling pattern application
- Portability improvements
- Documentation standardization

**Usage**:
```bash
# Fix all shell scripts
./scripts/fix_bash_standards.sh

# Fix specific script
./scripts/fix_bash_standards.sh scripts/build.sh

# Preview changes without applying
./scripts/fix_bash_standards.sh --dry-run

# Fix with custom standards
./scripts/fix_bash_standards.sh --config bash-standards.yaml
```

**Standards Applied**:
```bash
#!/bin/bash
# Always use strict mode
set -euo pipefail

# Declare variables at top
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly VERSION="1.0.0"

# Error handling function
error_exit() {
    echo "ERROR: $1" >&2
    exit 1
}

# Main function
main() {
    # Implementation
}

# Script entry point
main "$@"
```

### 3. System Management Scripts

#### `start_autonomous_system.sh` - Autonomous System Startup

**Purpose**: Complete autonomous system startup with all components.

**Design Principles**:
- **Autonomous Operation**: Minimal human intervention required
- **Self-Healing**: Automatic recovery from failures
- **Monitoring**: Comprehensive system monitoring
- **Scalability**: Support for multiple instances

**Key Features**:
- Complete system orchestration
- Health monitoring and alerting
- Automatic recovery mechanisms
- Resource monitoring and optimization
- Log aggregation and analysis

**Usage**:
```bash
# Start autonomous system
./scripts/start_autonomous_system.sh

# Start with monitoring
./scripts/start_autonomous_system.sh --monitor

# Start in cluster mode
./scripts/start_autonomous_system.sh --cluster

# Start with custom configuration
./scripts/start_autonomous_system.sh --config cluster.yaml
```

#### `install_qdrant_mcp.sh` - Qdrant MCP Installation

**Purpose**: Automated installation and configuration of Qdrant MCP server.

**Design Principles**:
- **Automated Setup**: Minimal manual intervention
- **Configuration Management**: Proper configuration setup
- **Service Integration**: Integration with system services
- **Health Verification**: Post-installation health checks

**Key Features**:
- Qdrant server installation
- MCP server setup and configuration
- Service integration and startup
- Health verification and testing
- Configuration backup and restore

**Usage**:
```bash
# Install Qdrant MCP
./scripts/install_qdrant_mcp.sh

# Install with custom configuration
./scripts/install_qdrant_mcp.sh --config qdrant-config.yaml

# Install specific version
./scripts/install_qdrant_mcp.sh --version 1.0.0

# Install with development setup
./scripts/install_qdrant_mcp.sh --dev
```

### 4. Maintenance and Cleanup Scripts

#### `cleanup_project.sh` - Project Cleanup Automation

**Purpose**: Automated project cleanup and maintenance.

**Design Principles**:
- **Safe Cleanup**: Never delete important data
- **Selective Cleaning**: Target specific areas for cleanup
- **Backup Creation**: Create backups before cleanup
- **Verification**: Verify cleanup operations

**Key Features**:
- Temporary file cleanup
- Build artifact cleanup
- Log file rotation and cleanup
- Cache cleanup and optimization
- Database maintenance

**Usage**:
```bash
# Standard cleanup
./scripts/cleanup_project.sh

# Aggressive cleanup (remove more files)
./scripts/cleanup_project.sh --aggressive

# Cleanup specific areas
./scripts/cleanup_project.sh --logs --cache

# Cleanup with backup
./scripts/cleanup_project.sh --backup
```

## Script Standards

### BASH Scripting Standards

All scripts follow strict BASH standards:

```bash
#!/bin/bash
# Script: script_name.sh
# Purpose: Brief description of script purpose
# Author: Author name
# Date: Creation date
# Version: 1.0.0

# Strict mode - fail fast, fail loud
set -euo pipefail

# Script configuration
readonly SCRIPT_NAME="${0##*/}"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Color codes for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

# Error handling
error_exit() {
    log_error "$1"
    exit 1
}

# Cleanup function
cleanup() {
    log_info "Cleaning up..."
    # Cleanup operations
}

# Signal handling
trap cleanup EXIT
trap 'error_exit "Interrupted by user"' INT TERM

# Main function
main() {
    log_info "Starting ${SCRIPT_NAME}..."
    
    # Validate prerequisites
    validate_prerequisites
    
    # Execute main logic
    execute_main_logic
    
    log_success "${SCRIPT_NAME} completed successfully"
}

# Validate prerequisites
validate_prerequisites() {
    log_info "Validating prerequisites..."
    
    # Check required commands
    command -v go >/dev/null 2>&1 || error_exit "Go is required but not installed"
    command -v git >/dev/null 2>&1 || error_exit "Git is required but not installed"
    
    # Check required files
    [[ -f "${PROJECT_ROOT}/go.mod" ]] || error_exit "go.mod not found in project root"
    
    log_success "Prerequisites validated"
}

# Execute main logic
execute_main_logic() {
    log_info "Executing main logic..."
    
    # Main implementation
    # ...
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi
```

### Error Handling Patterns

```bash
# Function with error handling
safe_command() {
    local cmd="$1"
    local description="$2"
    
    log_info "Executing: ${description}"
    
    if ! eval "$cmd"; then
        error_exit "Failed to execute: ${description}"
    fi
    
    log_success "Completed: ${description}"
}

# Retry pattern
retry_command() {
    local cmd="$1"
    local max_attempts="${2:-3}"
    local delay="${3:-5}"
    
    for attempt in $(seq 1 "$max_attempts"); do
        log_info "Attempt $attempt/$max_attempts"
        
        if eval "$cmd"; then
            log_success "Command succeeded on attempt $attempt"
            return 0
        fi
        
        if [[ $attempt -lt $max_attempts ]]; then
            log_warning "Command failed, retrying in ${delay}s..."
            sleep "$delay"
        fi
    done
    
    error_exit "Command failed after $max_attempts attempts"
}
```

### Configuration Management

```bash
# Load configuration
load_config() {
    local config_file="$1"
    
    if [[ ! -f "$config_file" ]]; then
        error_exit "Configuration file not found: $config_file"
    fi
    
    # Source configuration file
    source "$config_file"
    
    # Validate required variables
    validate_config_variables
}

# Validate configuration variables
validate_config_variables() {
    local required_vars=("BUILD_VERSION" "PROJECT_ROOT" "GO_VERSION")
    
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var:-}" ]]; then
            error_exit "Required configuration variable not set: $var"
        fi
    done
}
```

## Testing and Validation

### Script Testing

```bash
# Test script functionality
test_script() {
    local script="$1"
    
    log_info "Testing script: $script"
    
    # Run with test mode
    if ! bash -n "$script"; then
        error_exit "Syntax error in $script"
    fi
    
    # Run with dry-run mode if supported
    if grep -q "dry-run" "$script"; then
        if ! bash "$script" --dry-run; then
            error_exit "Dry-run failed for $script"
        fi
    fi
    
    log_success "Script test passed: $script"
}
```

### Integration Testing

```bash
# Test script integration
test_integration() {
    log_info "Running integration tests..."
    
    # Test build process
    test_script "scripts/build.sh"
    
    # Test linting
    test_script "scripts/lint.sh"
    
    # Test cleanup
    test_script "scripts/cleanup_project.sh"
    
    log_success "Integration tests passed"
}
```

## Monitoring and Logging

### Log Management

```bash
# Setup logging
setup_logging() {
    local log_dir="${PROJECT_ROOT}/logs"
    local log_file="${log_dir}/scripts.log"
    
    mkdir -p "$log_dir"
    
    # Redirect output to log file
    exec 1> >(tee -a "$log_file")
    exec 2> >(tee -a "$log_file" >&2)
    
    log_info "Logging to: $log_file"
}

# Log rotation
rotate_logs() {
    local log_dir="${PROJECT_ROOT}/logs"
    local max_logs=10
    
    find "$log_dir" -name "*.log" -mtime +7 -delete
    
    # Keep only recent logs
    find "$log_dir" -name "*.log" -exec ls -t {} + | tail -n +$((max_logs + 1)) | xargs -r rm
}
```

### Health Monitoring

```bash
# Check script health
check_script_health() {
    local script="$1"
    
    # Check if script exists and is executable
    if [[ ! -x "$script" ]]; then
        log_error "Script not executable: $script"
        return 1
    fi
    
    # Check script syntax
    if ! bash -n "$script"; then
        log_error "Syntax error in script: $script"
        return 1
    fi
    
    # Check for required functions
    local required_functions=("main" "error_exit")
    for func in "${required_functions[@]}"; do
        if ! grep -q "function $func\|$func()" "$script"; then
            log_warning "Missing function in $script: $func"
        fi
    done
    
    log_success "Script health check passed: $script"
    return 0
}
```

## Security Considerations

### Input Validation

```bash
# Validate script inputs
validate_inputs() {
    local args=("$@")
    
    for arg in "${args[@]}"; do
        # Check for dangerous patterns
        if [[ "$arg" =~ [;&|`$] ]]; then
            error_exit "Dangerous characters in input: $arg"
        fi
        
        # Check for path traversal
        if [[ "$arg" =~ \.\. ]]; then
            error_exit "Path traversal detected: $arg"
        fi
    done
}

# Sanitize file paths
sanitize_path() {
    local path="$1"
    
    # Remove dangerous characters
    path="${path//[;&|`$]/}"
    
    # Ensure path is within project
    path="$(realpath "$path")"
    if [[ ! "$path" =~ ^"$PROJECT_ROOT" ]]; then
        error_exit "Path outside project directory: $path"
    fi
    
    echo "$path"
}
```

### Permission Management

```bash
# Set secure permissions
set_secure_permissions() {
    local file="$1"
    
    # Set restrictive permissions
    chmod 600 "$file"
    
    # Ensure proper ownership
    chown "$(whoami):$(whoami)" "$file"
    
    log_info "Set secure permissions for: $file"
}
```

## Troubleshooting

### Common Issues

1. **Permission Denied**: Check script permissions and ownership
2. **Command Not Found**: Verify required dependencies are installed
3. **Configuration Errors**: Check configuration file syntax and values
4. **Network Issues**: Verify network connectivity for external dependencies

### Debug Mode

```bash
# Enable debug mode
if [[ "${DEBUG:-false}" == "true" ]]; then
    set -x
    PS4='+(${BASH_SOURCE}:${LINENO}): ${FUNCNAME[0]:+${FUNCNAME[0]}(): }'
fi
```

### Recovery Procedures

```bash
# Recovery function
recover_from_failure() {
    local error_code="$1"
    local context="$2"
    
    log_error "Recovery needed: $context (error: $error_code)"
    
    case "$error_code" in
        1) # General error
            cleanup
            ;;
        2) # Configuration error
            restore_backup_config
            ;;
        3) # Network error
            retry_network_operations
            ;;
        *) # Unknown error
            log_error "Unknown error code: $error_code"
            ;;
    esac
}
```

## Future Enhancements

### Planned Features

- **Automated Testing**: Comprehensive test automation for all scripts
- **Configuration Management**: Advanced configuration management system
- **Monitoring Integration**: Integration with monitoring systems
- **CI/CD Integration**: Automated deployment and testing

### Extension Points

- **Plugin System**: Extensible script functionality
- **Custom Validators**: User-defined validation rules
- **Multi-Platform Support**: Cross-platform script compatibility
- **Cloud Integration**: Cloud-specific automation features

---

**Production Ready**: All scripts are designed for production use with comprehensive error handling, logging, and security features. They provide reliable automation for the MQTT Agent Orchestration System.
