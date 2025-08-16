#!/usr/bin/env bash

# Design: Autonomous MQTT workflow system startup script
# Purpose: Start the complete multi-role worker system for autonomous document generation
# Usage: ./start_autonomous_system.sh

set -euo pipefail

# ALL global variables declared at top
declare -r PROJECT_NAME_GLOBAL="MQTT Agent Orchestration System"
declare -r MQTT_HOST_GLOBAL="localhost"
declare -r MQTT_PORT_GLOBAL="1883"
declare LOG_DIR_GLOBAL="logs"
declare -a WORKER_PIDS_GLOBAL=()

function log_info() {
    local message="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] INFO: $message"
}

function log_error() {
    local message="$1"
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $message" >&2
}

# Check dependencies
function check_dependencies() {
    local dep=""
    local -a required_deps=("./bin/orchestrator" "./bin/role-worker" "./bin/client")
    
    for dep in "${required_deps[@]}"; do
        if [[ ! -f "$dep" ]]; then
            log_error "Required binary not found: $dep"
            log_error "Please run: ./scripts/build.sh"
            exit 1
        fi
    done
    
    # Check mosquitto service
    local mosquitto_status=""
    mosquitto_status=$(systemctl is-active mosquitto)
    if [[ "$mosquitto_status" != "active" ]]; then
        log_error "Mosquitto MQTT broker is not running"
        log_error "Please start with: sudo systemctl start mosquitto"
        exit 1
    fi
    
    log_info "All dependencies available"
}

# Setup logging directory
function setup_logging() {
    local mkdir_result=""
    local mkdir_exit=""
    
    mkdir_result=$(mkdir -p "$LOG_DIR_GLOBAL" 2>&1)
    mkdir_exit="$?"
    
    if [[ "$mkdir_exit" -ne 0 ]]; then
        log_error "Failed to create logs directory: $mkdir_result"
        exit 1
    fi
}

# Start orchestrator
function start_orchestrator() {
    log_info "Starting workflow orchestrator..."
    
    "./bin/orchestrator" \
        --mqtt-host "$MQTT_HOST_GLOBAL" \
        --mqtt-port "$MQTT_PORT_GLOBAL" \
        --verbose > "$LOG_DIR_GLOBAL/orchestrator.log" 2>&1 &
    
    local orchestrator_pid=$!
    WORKER_PIDS_GLOBAL+=("$orchestrator_pid")
    
    log_info "Orchestrator started (PID: $orchestrator_pid)"
}

# Start role-specific workers
function start_role_workers() {
    local -a roles=("developer" "reviewer" "approver" "tester")
    local role=""
    local worker_pid=""
    
    for role in "${roles[@]}"; do
        log_info "Starting $role worker..."
        
        "./bin/role-worker" \
            --id "${role}-worker-1" \
            --role "$role" \
            --mqtt-host "$MQTT_HOST_GLOBAL" \
            --mqtt-port "$MQTT_PORT_GLOBAL" \
            --verbose > "$LOG_DIR_GLOBAL/${role}-worker.log" 2>&1 &
        
        worker_pid=$!
        WORKER_PIDS_GLOBAL+=("$worker_pid")
        
        log_info "$role worker started (PID: $worker_pid)"
        
        # Small delay between workers
        sleep 1
    done
}

# Check system health
function check_system_health() {
    local attempts=0
    local max_attempts=10
    local health_check_result=""
    local health_check_exit=""
    
    log_info "Performing system health check..."
    
    while [[ $attempts -lt $max_attempts ]]; do
        attempts=$((attempts + 1))
        
        # Test MQTT connectivity
        health_check_result=$(mosquitto_pub -h "$MQTT_HOST_GLOBAL" -p "$MQTT_PORT_GLOBAL" -t "health/check" -m "ping" 2>&1)
        health_check_exit="$?"
        
        if [[ "$health_check_exit" -eq 0 ]]; then
            log_info "System health check passed (attempt $attempts/$max_attempts)"
            return 0
        fi
        
        if [[ $attempts -lt $max_attempts ]]; then
            log_info "Health check failed, retrying... (attempt $attempts/$max_attempts)"
            sleep 2
        fi
    done
    
    log_error "System health check failed after $max_attempts attempts"
    return 1
}

# Show system status
function show_system_status() {
    local pid=""
    local running_count=0
    
    log_info "System Status:"
    echo "  MQTT Broker: RUNNING (mosquitto)"
    
    for pid in "${WORKER_PIDS_GLOBAL[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            running_count=$((running_count + 1))
        fi
    done
    
    echo "  Active Workers: $running_count/${#WORKER_PIDS_GLOBAL[@]}"
    echo "  Log Directory: $LOG_DIR_GLOBAL/"
    echo ""
    echo "The autonomous workflow system is ready!"
    echo ""
    echo "To create a Go coding standards document:"
    echo "  ./bin/client --doc-type go_coding_standards --output /home/niko/.claude/GO_CODING_STANDARD_CLAUDE.md"
    echo ""
    echo "To see available document types:"
    echo "  ./bin/client --list"
    echo ""
    echo "Monitor logs:"
    echo "  tail -f $LOG_DIR_GLOBAL/*.log"
}

# Cleanup function
function cleanup() {
    local pid=""
    
    log_info "Stopping autonomous workflow system..."
    
    for pid in "${WORKER_PIDS_GLOBAL[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" 2>/dev/null || true
        fi
    done
    
    # Wait a moment for graceful shutdown
    sleep 2
    
    # Force kill any remaining processes
    pkill -f "bin/orchestrator" 2>/dev/null || true
    pkill -f "bin/role-worker" 2>/dev/null || true
    
    log_info "System stopped"
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Main function
function main() {
    log_info "Starting $PROJECT_NAME_GLOBAL"
    
    check_dependencies
    setup_logging
    
    # Start system components
    start_orchestrator
    sleep 2  # Let orchestrator initialize
    
    start_role_workers
    sleep 3  # Let workers connect
    
    # Verify system health
    if check_system_health; then
        show_system_status
        
        log_info "System started successfully!"
        log_info "Press Ctrl+C to stop the system"
        
        # Keep script running
        while true; do
            sleep 10
            
            # Check if any processes died
            local dead_count=0
            local pid=""
            for pid in "${WORKER_PIDS_GLOBAL[@]}"; do
                if ! kill -0 "$pid" 2>/dev/null; then
                    dead_count=$((dead_count + 1))
                fi
            done
            
            if [[ $dead_count -gt 0 ]]; then
                log_error "$dead_count process(es) have died unexpectedly"
                log_error "Check logs in $LOG_DIR_GLOBAL/ for details"
                break
            fi
        done
    else
        log_error "System health check failed - stopping"
        cleanup
        exit 1
    fi
}

# Start the system
main