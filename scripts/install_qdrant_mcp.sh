#!/usr/bin/env bash

# Qdrant MCP Server Installation Script for MQTT Agent Orchestration System
# Purpose: Install and configure the official Qdrant MCP server
# Usage: ./scripts/install_qdrant_mcp.sh [--local] [--remote]

set -euo pipefail

# Configuration
declare -r SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
declare -r PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
declare -r QDRANT_URL="${QDRANT_URL:-http://localhost:6333}"
declare -r COLLECTION_NAME="${COLLECTION_NAME:-project_knowledge}"
declare -r EMBEDDING_MODEL="${EMBEDDING_MODEL:-sentence-transformers/all-MiniLM-L6-v2}"
declare -r MCP_PORT="${MCP_PORT:-8000}"

# Colors for output
declare -r RED='\033[0;31m'
declare -r GREEN='\033[0;32m'
declare -r YELLOW='\033[1;33m'
declare -r BLUE='\033[0;34m'
declare -r NC='\033[0m' # No Color

# Logging functions
function log_info() {
    local message="$1"
    echo -e "${BLUE}[INFO]${NC} $message"
}

function log_success() {
    local message="$1"
    echo -e "${GREEN}[SUCCESS]${NC} $message"
}

function log_warning() {
    local message="$1"
    echo -e "${YELLOW}[WARNING]${NC} $message"
}

function log_error() {
    local message="$1"
    echo -e "${RED}[ERROR]${NC} $message"
}

# Help function
function show_help() {
    cat << EOF
Qdrant MCP Server Installation Script

Usage: $0 [OPTIONS]

Options:
  --local     Install for local development (stdio transport)
  --remote    Install for remote access (SSE transport)
  --docker    Use Docker instead of uvx
  --help      Show this help message

Environment Variables:
  QDRANT_URL         Qdrant server URL (default: http://localhost:6333)
  COLLECTION_NAME    Collection name (default: project_knowledge)
  EMBEDDING_MODEL    Embedding model (default: sentence-transformers/all-MiniLM-L6-v2)
  MCP_PORT          MCP server port (default: 8000)

Examples:
  $0 --local                    # Local development setup
  $0 --remote                   # Remote access setup
  QDRANT_URL=http://remote:6333 $0 --remote  # Remote Qdrant setup
EOF
}

# Check dependencies
function check_dependencies() {
    local missing_deps=()
    local dep=""
    local check_result=""
    local check_exit=""
    
    # Check for uvx (preferred method)
    check_result=$(command -v uvx 2>&1)
    check_exit="$?"
    if [[ "$check_exit" -ne 0 ]]; then
        missing_deps+=("uvx")
    fi
    
    # Check for Docker (alternative method)
    check_result=$(command -v docker 2>&1)
    check_exit="$?"
    if [[ "$check_exit" -ne 0 ]]; then
        missing_deps+=("docker")
    fi
    
    # Check for curl
    check_result=$(command -v curl 2>&1)
    check_exit="$?"
    if [[ "$check_exit" -ne 0 ]]; then
        missing_deps+=("curl")
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        log_error "Missing dependencies: ${missing_deps[*]}"
        log_info "Install uvx: curl -LsSf https://astral.sh/uv/install.sh | sh"
        log_info "Install Docker: https://docs.docker.com/get-docker/"
        exit 1
    fi
    
    log_success "All dependencies found"
}

# Test Qdrant connection
test_qdrant_connection() {
    log_info "Testing Qdrant connection..."
    
    if curl -s "$QDRANT_URL/health" >/dev/null 2>&1; then
        log_success "Qdrant is accessible at $QDRANT_URL"
    else
        log_error "Cannot connect to Qdrant at $QDRANT_URL"
        log_info "Make sure Qdrant is running and accessible"
        exit 1
    fi
}

# Install using uvx
install_with_uvx() {
    local transport="$1"
    
    log_info "Installing Qdrant MCP server using uvx..."
    
    # Test if mcp-server-qdrant is available
    if ! uvx mcp-server-qdrant --help >/dev/null 2>&1; then
        log_info "Installing mcp-server-qdrant..."
        uvx install mcp-server-qdrant
    fi
    
    log_success "Qdrant MCP server installed via uvx"
    
    # Create systemd service for production
    if [[ "$transport" == "sse" ]]; then
        create_systemd_service
    fi
}

# Install using Docker
install_with_docker() {
    log_info "Installing Qdrant MCP server using Docker..."
    
    # Pull the Docker image
    docker pull qdrant/mcp-server-qdrant:latest
    
    log_success "Qdrant MCP server Docker image pulled"
    
    # Create Docker Compose file
    create_docker_compose
}

# Create systemd service
create_systemd_service() {
    local service_file="/etc/systemd/system/qdrant-mcp-server.service"
    
    log_info "Creating systemd service..."
    
    sudo tee "$service_file" >/dev/null << EOF
[Unit]
Description=Qdrant MCP Server
After=network.target

[Service]
Type=simple
User=niko
Environment=QDRANT_URL=$QDRANT_URL
Environment=COLLECTION_NAME=$COLLECTION_NAME
Environment=EMBEDDING_MODEL=$EMBEDDING_MODEL
Environment=FASTMCP_HOST=0.0.0.0
Environment=FASTMCP_PORT=$MCP_PORT
ExecStart=/usr/local/bin/uvx mcp-server-qdrant --transport sse
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
    
    sudo systemctl daemon-reload
    sudo systemctl enable qdrant-mcp-server
    
    log_success "Systemd service created and enabled"
    log_info "Start with: sudo systemctl start qdrant-mcp-server"
    log_info "Check status with: sudo systemctl status qdrant-mcp-server"
}

# Create Docker Compose file
create_docker_compose() {
    local compose_file="$PROJECT_ROOT/docker-compose.mcp.yml"
    
    log_info "Creating Docker Compose file..."
    
    cat > "$compose_file" << EOF
version: '3.8'

services:
  qdrant-mcp-server:
    image: qdrant/mcp-server-qdrant:latest
    container_name: qdrant-mcp-server
    ports:
      - "$MCP_PORT:8000"
    environment:
      - QDRANT_URL=$QDRANT_URL
      - COLLECTION_NAME=$COLLECTION_NAME
      - EMBEDDING_MODEL=$EMBEDDING_MODEL
      - FASTMCP_HOST=0.0.0.0
      - FASTMCP_PORT=8000
    restart: unless-stopped
    networks:
      - mqtt-network

networks:
  mqtt-network:
    external: true
EOF
    
    log_success "Docker Compose file created: $compose_file"
    log_info "Start with: docker-compose -f $compose_file up -d"
}

# Test MCP server
test_mcp_server() {
    local transport="$1"
    
    log_info "Testing MCP server..."
    
    if [[ "$transport" == "sse" ]]; then
        # Test SSE endpoint
        if curl -s "http://localhost:$MCP_PORT/sse" >/dev/null 2>&1; then
            log_success "MCP server is running on port $MCP_PORT"
        else
            log_warning "MCP server may not be running. Start it manually."
        fi
    else
        log_info "For stdio transport, test by running: uvx mcp-server-qdrant"
    fi
}

# Update project configuration
update_project_config() {
    log_info "Updating project configuration..."
    
    # Update MCP configuration
    local mcp_config="$PROJECT_ROOT/configs/mcp.yaml"
    
    if [[ -f "$mcp_config" ]]; then
        log_success "MCP configuration updated"
    else
        log_warning "MCP configuration file not found: $mcp_config"
    fi
}

# Main installation function
main() {
    local install_method="uvx"
    local transport="stdio"
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --local)
                transport="stdio"
                shift
                ;;
            --remote)
                transport="sse"
                shift
                ;;
            --docker)
                install_method="docker"
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    log_info "Installing Qdrant MCP Server"
    log_info "Qdrant URL: $QDRANT_URL"
    log_info "Collection: $COLLECTION_NAME"
    log_info "Embedding Model: $EMBEDDING_MODEL"
    log_info "Transport: $transport"
    log_info "Install Method: $install_method"
    
    check_dependencies
    test_qdrant_connection
    
    if [[ "$install_method" == "docker" ]]; then
        install_with_docker
    else
        install_with_uvx "$transport"
    fi
    
    test_mcp_server "$transport"
    update_project_config
    
    log_success "Qdrant MCP Server installation completed!"
    
    if [[ "$transport" == "sse" ]]; then
        log_info "MCP server will be available at: http://localhost:$MCP_PORT/sse"
    fi
    
    log_info "Update your MCP client configuration to use the new server"
}

# Run main function
main "$@"
