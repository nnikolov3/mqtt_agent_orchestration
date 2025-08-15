#!/usr/bin/env bash

# Design: Build script for MQTT Agent Orchestration System
# Purpose: Build Go binaries and Docker images following coding standards
# Usage: ./scripts/build.sh [--verbose] [--clean]

set -euo pipefail

# ALL global variables declared at top
declare -r PROJECT_NAME_GLOBAL="mqtt-orchestrator"
declare -r VERSION_GLOBAL="${VERSION:-$(git describe --tags --always --dirty || echo 'v0.1.0-dev')}"
declare -r BUILD_TIME_GLOBAL="$(date -u '+%Y-%m-%d_%H:%M:%S')"
declare -r GIT_COMMIT_GLOBAL="$(git rev-parse HEAD || echo 'unknown')"
declare -r BUILD_DIR_GLOBAL="bin"
declare VERBOSE_GLOBAL=false
declare CLEAN_GLOBAL=false

# Check dependencies
function check_dependencies() {
    local dep=""
    local -a required_deps=("go" "git")
    local check_result=""
    local command_exit=""
    
    for dep in "${required_deps[@]}"; do
        if command -v "$dep"; then
            log_info "Found dependency: $dep"
        else
            log_error "Required dependency not found: $dep"
            exit 1
        fi
    done
}

# Logging functions
function log_info() {
    local message="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $message"
}

function log_error() {
    local message="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $message" >&2
}

function log_warn() {
    local message="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] WARN: $message" >&2
}

# Parse command line arguments
function parse_args() {
    local arg=""
    
    while [[ $# -gt 0 ]]; do
        arg="$1"
        case "$arg" in
            --verbose)
                VERBOSE_GLOBAL=true
                shift
                ;;
            --clean)
                CLEAN_GLOBAL=true
                shift
                ;;
            --help)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown option: $arg"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Show usage information
function show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build script for MQTT Agent Orchestration System

OPTIONS:
    --verbose    Enable verbose output
    --clean      Clean build artifacts before building
    --help       Show this help message

EXAMPLES:
    $0                  Build with default settings
    $0 --verbose        Build with verbose output
    $0 --clean          Clean and rebuild
EOF
}

# Clean build artifacts
function clean_build() {
    local rm_result=""
    local rm_exit=""
    
    log_info "Cleaning build artifacts..."
    
    if [[ -d "$BUILD_DIR_GLOBAL" ]]; then
        rm_result=$(rm -rf "$BUILD_DIR_GLOBAL")
        rm_exit=$?
        
        if [[ $rm_exit -ne 0 ]]; then
            log_error "Failed to clean build directory: $rm_result"
            exit 1
        fi
    fi
    
    log_info "Build artifacts cleaned"
}

# Create build directory
function create_build_dir() {
    local mkdir_result=""
    local mkdir_exit=""
    
    mkdir_result=$(mkdir -p "$BUILD_DIR_GLOBAL")
    mkdir_exit=$?
    
    if [[ $mkdir_exit -ne 0 ]]; then
        log_error "Failed to create build directory: $mkdir_result"
        exit 1
    fi
}

# Build Go binary
function build_go_binary() {
    local binary_name="$1"
    local source_path="$2"
    local ldflags=""
    local go_build_result=""
    local go_build_exit=""
    
    ldflags="-X main.Version=$VERSION_GLOBAL -X main.BuildTime=$BUILD_TIME_GLOBAL -X main.GitCommit=$GIT_COMMIT_GLOBAL"
    
    log_info "Building Go binary: $binary_name"
    
    if [[ "$VERBOSE_GLOBAL" == true ]]; then
        log_info "Source: $source_path"
        log_info "Output: $BUILD_DIR_GLOBAL/$binary_name"
        log_info "Version: $VERSION_GLOBAL"
    fi
    
    go_build_result=$(go build -ldflags "$ldflags" -o "$BUILD_DIR_GLOBAL/$binary_name" "$source_path")
    go_build_exit=$?
    
    if [[ $go_build_exit -ne 0 ]]; then
        log_error "Failed to build $binary_name: $go_build_result"
        exit 1
    fi
    
    log_info "Successfully built $binary_name"
}

# Run tests
function run_tests() {
    local test_result=""
    local test_exit=""
    
    log_info "Running Go tests..."
    
    test_result=$(go test ./...)
    test_exit=$?
    
    if [[ $test_exit -ne 0 ]]; then
        log_error "Tests failed: $test_result"
        exit 1
    fi
    
    log_info "All tests passed"
}

# Run linting
function run_linting() {
    local lint_result=""
    local lint_exit=""
    local lint_check_result=""
    local command_exit=""
    
    if command -v golangci-lint; then
        log_info "Running golangci-lint..."
        
        lint_result=$(golangci-lint run)
        lint_exit=$?
        
        if [[ $lint_exit -ne 0 ]]; then
            log_error "Linting failed: $lint_result"
            exit 1
        fi
        
        log_info "Linting passed"
    else
        log_warn "golangci-lint not found, skipping linting"
    fi
}

# Main build function
function main() {
    log_info "Starting build process for $PROJECT_NAME_GLOBAL"
    log_info "Version: $VERSION_GLOBAL"
    log_info "Build Time: $BUILD_TIME_GLOBAL"
    log_info "Git Commit: $GIT_COMMIT_GLOBAL"
    
    check_dependencies
    
    if [[ "$CLEAN_GLOBAL" == true ]]; then
        clean_build
    fi
    
    create_build_dir
    
    # Run tests and linting
    run_tests
    run_linting
    
    # Build Go binaries
    build_go_binary "worker" "./cmd/worker"
    build_go_binary "server" "./cmd/server"
    build_go_binary "orchestrator" "./cmd/orchestrator"
    build_go_binary "role-worker" "./cmd/role-worker"
    build_go_binary "client" "./cmd/client"
    
    log_info "Build completed successfully"
    log_info "Binaries available in: $BUILD_DIR_GLOBAL/"
}

# Parse arguments and run main
parse_args "$@"
main
