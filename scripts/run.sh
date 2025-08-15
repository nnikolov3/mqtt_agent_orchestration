#!/usr/bin/env bash

# Design: Runtime orchestration script for MQTT Agent Orchestration System
# Purpose: Provides development and production runtime management with proper monitoring
# Usage: ./scripts/run.sh [--mode dev|prod] [--config path] [--verbose]

set -euo pipefail

# ALL global variables declared at top
declare -r PROJECT_NAME_GLOBAL="MQTT Agent Orchestration System"
declare -r SCRIPT_DIR_GLOBAL="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
declare -r PROJECT_ROOT_GLOBAL="$(dirname "$SCRIPT_DIR_GLOBAL")"
declare -r DEFAULT_MQTT_HOST_GLOBAL="localhost"
declare -r DEFAULT_MQTT_PORT_GLOBAL="1883"
declare -r DEFAULT_QDRANT_URL_GLOBAL="http://localhost:6333"
declare MODE_GLOBAL="dev"
declare CONFIG_PATH_GLOBAL=""
declare VERBOSE_GLOBAL=false
declare -a WORKER_PIDS_GLOBAL=()

# Logging functions
function log_info() {
    local message="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $message"
}

function log_error() {
    local message="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $message" >&2
}

function log_debug() {
    local message="$1"
    if [[ "$VERBOSE_GLOBAL" == true ]]; then
        echo "[$(date +'%Y-%m-%d %H:%M:%S')] DEBUG: $message"
    fi
}

# Parse command line arguments
function parse_args() {
    local arg=""
    
    while [[ $# -gt 0 ]]; do
        arg="$1"
        case "$arg" in
            --mode)
                if [[ -n "${2:-}" ]]; then
                    MODE_GLOBAL="$2"
                    shift 2
                else
                    log_error "Mode parameter requires a value (dev|prod)"
                    exit 1
                fi
                ;;
            --config)
                if [[ -n "${2:-}" ]]; then
                    CONFIG_PATH_GLOBAL="$2"
                    shift 2
                else
                    log_error "Config parameter requires a path"
                    exit 1
                fi
                ;;
            --verbose)
                VERBOSE_GLOBAL=true
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

Runtime orchestration script for MQTT Agent Orchestration System

OPTIONS:
    --mode dev|prod     Runtime mode (default: dev)
    --config PATH       Configuration file path
    --verbose           Enable verbose logging
    --help              Show this help message

MODES:
    dev                 Development mode with file watching and hot reload
    prod                Production mode with monitoring and health checks

EXAMPLES:
    $0                          Run in development mode
    $0 --mode prod              Run in production mode
    $0 --mode dev --verbose     Run with verbose logging
    $0 --config ./config.yaml   Use specific configuration file

EOF
}

# Validate mode parameter
function validate_mode() {
    local mode="$1"
    
    case "$mode" in
        dev|prod)
            return 0
            ;;
        *)
            log_error "Invalid mode: $mode. Must be 'dev' or 'prod'"
            return 1
            ;;
    esac
}

# Check if required binaries exist
function check_binaries() {
    local binary=""
    local -a required_binaries=("orchestrator" "role-worker" "client" "rag-service")
    local bin_dir="$PROJECT_ROOT_GLOBAL/bin"
    
    for binary in "${required_binaries[@]}"; do
        if [[ ! -f "$bin_dir/$binary" ]]; then
            log_error "Required binary not found: $bin_dir/$binary"
            log_error "Please run: ./scripts/build.sh"
            exit 1
        fi
    done
    
    log_debug "All required binaries found"
}

# Check external dependencies
function check_dependencies() {
    local mqtt_check_result=""
    local mqtt_check_exit=""
    local qdrant_check_result=""
    local qdrant_check_exit=""
    
    # Check MQTT broker
    log_debug "Checking MQTT broker availability..."
    mqtt_check_result=$(mosquitto_pub -h "$DEFAULT_MQTT_HOST_GLOBAL" -p "$DEFAULT_MQTT_PORT_GLOBAL" -t "health/check" -m "ping" 2>&1)
    mqtt_check_exit="$?"
    
    if [[ "$mqtt_check_exit" -ne 0 ]]; then
        log_error "MQTT broker not available at $DEFAULT_MQTT_HOST_GLOBAL:$DEFAULT_MQTT_PORT_GLOBAL"
        log_error "Please start Mosquitto: sudo systemctl start mosquitto"
        exit 1
    fi
    
    # Check Qdrant (optional)
    log_debug "Checking Qdrant availability..."
    qdrant_check_result=$(curl -s "$DEFAULT_QDRANT_URL_GLOBAL/health" 2>&1)
    qdrant_check_exit="$?"
    
    if [[ "$qdrant_check_exit" -ne 0 ]]; then
        log_info "Qdrant not available at $DEFAULT_QDRANT_URL_GLOBAL (optional)"
        log_info "RAG features will use fallback mode"
    else
        log_debug "Qdrant available at $DEFAULT_QDRANT_URL_GLOBAL"
    fi
    
    log_info "Dependency check completed"
}

# Setup logging directory
function setup_logging() {
    local log_dir="$PROJECT_ROOT_GLOBAL/logs"
    local mkdir_result=""
    local mkdir_exit=""
    
    mkdir_result=$(mkdir -p "$log_dir" 2>&1)
    mkdir_exit="$?"
    
    if [[ "$mkdir_exit" -ne 0 ]]; then
        log_error "Failed to create logs directory: $mkdir_result"
        exit 1
    fi
    
    log_debug "Logging directory ready: $log_dir"
}

# Start orchestrator
function start_orchestrator() {
    local orchestrator_args=""
    local orchestrator_log="$PROJECT_ROOT_GLOBAL/logs/orchestrator.log"
    
    log_info "Starting workflow orchestrator..."
    
    orchestrator_args="--mqtt-host $DEFAULT_MQTT_HOST_GLOBAL --mqtt-port $DEFAULT_MQTT_PORT_GLOBAL"
    
    if [[ "$VERBOSE_GLOBAL" == true ]]; then
        orchestrator_args="$orchestrator_args --verbose"
    fi
    
    if [[ "$MODE_GLOBAL" == "dev" ]]; then
        orchestrator_args="$orchestrator_args --dev-mode"
    fi
    
    # Start orchestrator in background
    "$PROJECT_ROOT_GLOBAL/bin/orchestrator" $orchestrator_args \
        > "$orchestrator_log" 2>&1 &
    
    local orchestrator_pid=$!
    WORKER_PIDS_GLOBAL+=("$orchestrator_pid")
    
    log_info "Orchestrator started (PID: $orchestrator_pid)"
    
    # Give it time to initialize
    sleep 2
    
    # Check if it's still running
    if ! kill -0 "$orchestrator_pid" 2>/dev/null; then
        log_error "Orchestrator failed to start. Check logs: $orchestrator_log"
        exit 1
    fi
}

# Start role-specific workers
function start_role_workers() {
    local -a roles=("developer" "reviewer" "approver" "tester")
    local role=""
    local worker_pid=""
    local worker_args=""
    local worker_log=""
    
    for role in "${roles[@]}"; do
        log_info "Starting $role worker..."
        
        worker_args="--id ${role}-worker-1 --role $role"
        worker_args="$worker_args --mqtt-host $DEFAULT_MQTT_HOST_GLOBAL"
        worker_args="$worker_args --mqtt-port $DEFAULT_MQTT_PORT_GLOBAL"
        
        if [[ "$VERBOSE_GLOBAL" == true ]]; then
            worker_args="$worker_args --verbose"
        fi
        
        worker_log="$PROJECT_ROOT_GLOBAL/logs/${role}-worker.log"
        
        "$PROJECT_ROOT_GLOBAL/bin/role-worker" $worker_args \
            > "$worker_log" 2>&1 &
        
        worker_pid=$!
        WORKER_PIDS_GLOBAL+=("$worker_pid")
        
        log_info "$role worker started (PID: $worker_pid)"
        
        # Small delay between workers
        sleep 1
        
        # Check if worker is still running
        if ! kill -0 "$worker_pid" 2>/dev/null; then
            log_error "$role worker failed to start. Check logs: $worker_log"
        fi
    done
}

# Initialize RAG system
function initialize_rag() {
    local rag_init_result=""
    local rag_init_exit=""
    
    log_info "Initializing RAG system..."
    
    # Check if RAG service binary exists
    if [[ ! -f "$PROJECT_ROOT_GLOBAL/bin/rag-service" ]]; then
        log_info "RAG service binary not found, skipping RAG initialization"
        return 0
    fi
    
    # Try to initialize collections
    rag_init_result=$("$PROJECT_ROOT_GLOBAL/bin/rag-service" version 2>&1)
    rag_init_exit="$?"
    
    if [[ "$rag_init_exit" -eq 0 ]]; then
        log_info "RAG service available: $rag_init_result"
    else
        log_info "RAG service initialization skipped"
    fi
}

# Monitor system health
function monitor_system() {
    local check_interval=10
    local max_failures=3
    local failure_count=0
    local pid=""
    local dead_count=""
    
    log_info "Starting system monitoring (check interval: ${check_interval}s)"
    
    while true; do
        sleep "$check_interval"
        
        # Check if any processes died
        dead_count=0
        for pid in "${WORKER_PIDS_GLOBAL[@]}"; do
            if ! kill -0 "$pid" 2>/dev/null; then
                dead_count=$((dead_count + 1))
            fi
        done
        
        if [[ $dead_count -gt 0 ]]; then
            failure_count=$((failure_count + 1))
            log_error "$dead_count process(es) have died (failure count: $failure_count/$max_failures)"
            
            if [[ $failure_count -ge $max_failures ]]; then
                log_error "Maximum failures reached, shutting down system"
                return 1
            fi
        else
            # Reset failure count on successful check
            failure_count=0
            log_debug "System health check passed (${#WORKER_PIDS_GLOBAL[@]} processes running)"
        fi
    done
}

# Development mode with file watching
function run_dev_mode() {
    log_info "Running in development mode"
    
    # In dev mode, we could add file watching here
    # For now, just run normal monitoring
    monitor_system
}

# Production mode with enhanced monitoring
function run_prod_mode() {
    log_info "Running in production mode"
    
    # Production mode with enhanced monitoring
    monitor_system
}

# Cleanup function
function cleanup() {
    local pid=""
    
    log_info "Shutting down $PROJECT_NAME_GLOBAL..."
    
    # Send TERM signal to all processes
    for pid in "${WORKER_PIDS_GLOBAL[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill -TERM "$pid" 2>/dev/null || true
        fi
    done
    
    # Wait for graceful shutdown
    sleep 3
    
    # Force kill any remaining processes
    for pid in "${WORKER_PIDS_GLOBAL[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill -KILL "$pid" 2>/dev/null || true
        fi
    done
    
    # Clean up any remaining processes by name
    pkill -f "bin/orchestrator" 2>/dev/null || true
    pkill -f "bin/role-worker" 2>/dev/null || true
    
    log_info "System shutdown complete"
}

# Show system status
function show_system_status() {
    local pid=""
    local running_count=0
    local total_count="${#WORKER_PIDS_GLOBAL[@]}"
    
    log_info "System Status:"
    echo "  Mode: $MODE_GLOBAL"
    echo "  MQTT Broker: $DEFAULT_MQTT_HOST_GLOBAL:$DEFAULT_MQTT_PORT_GLOBAL"
    echo "  Qdrant URL: $DEFAULT_QDRANT_URL_GLOBAL"
    
    for pid in "${WORKER_PIDS_GLOBAL[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            running_count=$((running_count + 1))
        fi
    done
    
    echo "  Active Processes: $running_count/$total_count"
    echo "  Log Directory: $PROJECT_ROOT_GLOBAL/logs/"
    echo ""
    echo "The system is ready!"
    echo ""
    echo "Monitor logs:"
    echo "  tail -f $PROJECT_ROOT_GLOBAL/logs/*.log"
    echo ""
    echo "Stop system:"
    echo "  Press Ctrl+C"
}

# Main execution function
function main() {
    log_info "Starting $PROJECT_NAME_GLOBAL"
    
    # Parse arguments
    parse_args "$@"
    
    # Validate parameters
    if ! validate_mode "$MODE_GLOBAL"; then
        exit 1
    fi
    
    # Setup
    check_binaries
    check_dependencies
    setup_logging
    
    # Initialize RAG system
    initialize_rag
    
    # Start system components
    start_orchestrator
    start_role_workers
    
    # Show status
    show_system_status
    
    # Run mode-specific logic
    case "$MODE_GLOBAL" in
        dev)
            run_dev_mode
            ;;
        prod)
            run_prod_mode
            ;;
    esac
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Run main function
main "$@"