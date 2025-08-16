#!/usr/bin/env bash

# MQTT Orchestration Project Cleanup Script
# Description: Removes unnecessary files, organizes project structure, and sets up CI/CD hooks
# Usage: ./cleanup.sh [options]
# Author: Project Maintainer
# Version: 1.0.0

# Exit on error, undefined variables, and pipeline failures
set -euo pipefail

# Enable debugging if DEBUG environment variable is set
[[ "${DEBUG:-0}" == "1" ]] && set -x

# Color codes for output formatting
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Script configuration
readonly SCRIPT_NAME="$(basename "${BASH_SOURCE[0]}")"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="${SCRIPT_DIR}"
readonly BACKUP_DIR="${PROJECT_ROOT}/.backup"
readonly LOG_FILE="${PROJECT_ROOT}/cleanup.log"

# Default configuration values
declare -i DRY_RUN=0
declare -i VERBOSE=0
declare -i BACKUP_ENABLED=1

# File patterns to remove (configurable arrays)
declare -a TEMP_FILES=(
    "*.tmp"
    "*.temp"
    "*~"
    "*.swp"
    "*.swo"
    ".DS_Store"
    "Thumbs.db"
)

declare -a LOG_FILES=(
    "*.log"
    "logs/*"
    "log/*"
)

declare -a BUILD_ARTIFACTS=(
    "dist/*"
    "build/*"
    "*.pyc"
    "__pycache__"
    "*.class"
    "*.o"
    "*.obj"
)

declare -a NODE_MODULES=(
    "node_modules"
    "package-lock.json"
    "yarn.lock"
)

declare -a PYTHON_CACHE=(
    "*.pyc"
    "__pycache__"
    ".pytest_cache"
    ".coverage"
    "htmlcov"
    ".tox"
)

# CI/CD configuration files to create
readonly CI_CONFIG_DIR=".github/workflows"
readonly CI_CONFIG_FILE="${CI_CONFIG_DIR}/mqtt-orchestration.yml"

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $*" | tee -a "${LOG_FILE}"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*" | tee -a "${LOG_FILE}"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*" | tee -a "${LOG_FILE}" >&2
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $*" | tee -a "${LOG_FILE}" >&2
}

# Function to log messages without color codes
log_message() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*" >> "${LOG_FILE}"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to create backup of files before deletion
create_backup() {
    local file_path="$1"
    local backup_path="${BACKUP_DIR}/$(dirname "${file_path}")"
    
    if [[ ${BACKUP_ENABLED} -eq 1 ]]; then
        mkdir -p "${backup_path}"
        if [[ -f "${file_path}" ]] || [[ -d "${file_path}" ]]; then
            cp -r "${file_path}" "${backup_path}/" 2>/dev/null || true
            print_info "Backed up: ${file_path}"
        fi
    fi
}

# Function to safely remove files/directories
safe_remove() {
    local pattern="$1"
    local description="${2:-}"
    
    # Find all matching files/directories
    local files
    mapfile -t files < <(find . -name "${pattern}" -not -path "./.git/*" 2>/dev/null || true)
    
    if [[ ${#files[@]} -gt 0 ]]; then
        for file in "${files[@]}"; do
            if [[ ${DRY_RUN} -eq 1 ]]; then
                print_info "[DRY RUN] Would remove ${description}: ${file}"
            else
                create_backup "${file}"
                rm -rf "${file}"
                print_success "Removed ${description}: ${file}"
            fi
        done
    elif [[ ${VERBOSE} -eq 1 ]]; then
        print_info "No ${description} found matching pattern: ${pattern}"
    fi
}

# Function to organize project structure
organize_structure() {
    print_info "Organizing project structure..."
    
    # Create standard directories if they don't exist
    local -a directories=(
        "src"
        "tests"
        "docs"
        "config"
        "scripts"
        "data"
        "logs"
    )
    
    for dir in "${directories[@]}"; do
        if [[ ! -d "${dir}" ]]; then
            if [[ ${DRY_RUN} -eq 1 ]]; then
                print_info "[DRY RUN] Would create directory: ${dir}"
            else
                mkdir -p "${dir}"
                print_success "Created directory: ${dir}"
            fi
        fi
    done
    
    # Move source files to src directory (example patterns)
    if [[ -d "src" ]]; then
        local -a source_patterns=(
            "*.py"
            "*.js"
            "*.ts"
            "*.java"
            "*.go"
        )
        
        for pattern in "${source_patterns[@]}"; do
            local files
            mapfile -t files < <(find . -maxdepth 1 -name "${pattern}" -type f 2>/dev/null || true)
            
            if [[ ${#files[@]} -gt 0 ]]; then
                for file in "${files[@]}"; do
                    # Skip if file is in a standard directory or is the cleanup script itself
                    if [[ "${file}" != "./${SCRIPT_NAME}" ]] && [[ ! -d "$(dirname "${file}")" ]] || [[ "$(dirname "${file}")" == "." ]]; then
                        if [[ ${DRY_RUN} -eq 1 ]]; then
                            print_info "[DRY RUN] Would move ${file} to src/"
                        else
                            mv "${file}" src/
                            print_success "Moved ${file} to src/"
                        fi
                    fi
                done
            fi
        done
    fi
}

# Function to setup CI/CD hooks
setup_ci_cd() {
    print_info "Setting up CI/CD hooks..."
    
    # Create GitHub Actions workflow directory
    if [[ ${DRY_RUN} -eq 1 ]]; then
        print_info "[DRY RUN] Would create directory: ${CI_CONFIG_DIR}"
    else
        mkdir -p "${CI_CONFIG_DIR}"
        print_success "Created CI/CD directory: ${CI_CONFIG_DIR}"
    fi
    
    # Create basic CI/CD workflow file
    if [[ ! -f "${CI_CONFIG_FILE}" ]]; then
        if [[ ${DRY_RUN} -eq 1 ]]; then
            print_info "[DRY RUN] Would create CI/CD workflow file: ${CI_CONFIG_FILE}"
        else
            cat > "${CI_CONFIG_FILE}" << 'EOF'
name: MQTT Orchestration CI/CD

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  build:
    runs-on: ubuntu-latest
    
    strategy:
      matrix:
        python-version: [3.8, 3.9, '3.10', '3.11']
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Python ${{ matrix.python-version }}
      uses: actions/setup-python@v4
      with:
        python-version: ${{ matrix.python-version }}
    
    - name: Install dependencies
      run: |
        python -m pip install --upgrade pip
        pip install -r requirements.txt
        pip install pytest
    
    - name: Run tests
      run: |
        pytest tests/
    
    - name: Lint code
      run: |
        # Add your linting commands here
        # pip install flake8
        # flake8 src/
    
    - name: Build package
      run: |
        # Add your build commands here
        # python setup.py sdist bdist_wheel
EOF
            print_success "Created CI/CD workflow file: ${CI_CONFIG_FILE}"
        fi
    else
        print_warning "CI/CD workflow file already exists: ${CI_CONFIG_FILE}"
    fi
    
    # Create pre-commit hook if git hooks directory exists
    local hooks_dir=".git/hooks"
    if [[ -d "${hooks_dir}" ]]; then
        local pre_commit_hook="${hooks_dir}/pre-commit"
        
        if [[ ! -f "${pre_commit_hook}" ]] || [[ ${DRY_RUN} -eq 1 ]]; then
            if [[ ${DRY_RUN} -eq 1 ]]; then
                print_info "[DRY RUN] Would create pre-commit hook: ${pre_commit_hook}"
            else
                cat > "${pre_commit_hook}" << 'EOF'
#!/bin/bash

# MQTT Orchestration Pre-commit Hook

# Run code formatting
echo "Running code formatting checks..."
# Add your formatting commands here

# Run linting
echo "Running linting checks..."
# Add your linting commands here

# Run tests
echo "Running tests..."
# pytest tests/ || exit 1

echo "Pre-commit checks passed!"
EOF
                
                chmod +x "${pre_commit_hook}"
                print_success "Created pre-commit hook: ${pre_commit_hook}"
            fi
        else
            print_warning "Pre-commit hook already exists: ${pre_commit_hook}"
        fi
    fi
}

# Function to remove temporary files
remove_temp_files() {
    print_info "Removing temporary files..."
    
    for pattern in "${TEMP_FILES[@]}"; do
        safe_remove "${pattern}" "temporary file"
    done
}

# Function to remove log files
remove_log_files() {
    print_info "Removing log files..."
    
    for pattern in "${LOG_FILES[@]}"; do
        safe_remove "${pattern}" "log file"
    done
}

# Function to remove build artifacts
remove_build_artifacts() {
    print_info "Removing build artifacts..."
    
    for pattern in "${BUILD_ARTIFACTS[@]}"; do
        safe_remove "${pattern}" "build artifact"
    done
}

# Function to remove dependency directories
remove_dependency_dirs() {
    print_info "Removing dependency directories..."
    
    for pattern in "${NODE_MODULES[@]}"; do
        safe_remove "${pattern}" "dependency directory"
    done
}

# Function to remove Python cache files
remove_python_cache() {
    print_info "Removing Python cache files..."
    
    for pattern in "${PYTHON_CACHE[@]}"; do
        safe_remove "${pattern}" "Python cache"
    done
}

# Function to clean up backup directory
cleanup_backup() {
    if [[ ${BACKUP_ENABLED} -eq 1 ]] && [[ -d "${BACKUP_DIR}" ]]; then
        print_info "Cleaning up backup directory older than 7 days..."
        if [[ ${DRY_RUN} -eq 1 ]]; then
            print_info "[DRY RUN] Would remove old backups from: ${BACKUP_DIR}"
        else
            find "${BACKUP_DIR}" -type f -mtime +7 -delete 2>/dev/null || true
            find "${BACKUP_DIR}" -type d -empty -delete 2>/dev/null || true
            print_success "Cleaned up old backups"
        fi
    fi
}

# Function to show usage information
show_usage() {
    cat << EOF
Usage: ${SCRIPT_NAME} [OPTIONS]

Options:
  -h, --help          Show this help message and exit
  -n, --dry-run       Show what would be done without actually doing it
  -v, --verbose       Enable verbose output
  -q, --quiet         Suppress output (except errors)
  --no-backup         Disable backup creation before deletion
  --backup-only       Only create backups, don't delete files

Examples:
  ${SCRIPT_NAME}              # Run cleanup with default settings
  ${SCRIPT_NAME} -n           # Dry run to see what would be cleaned
  ${SCRIPT_NAME} -v --no-backup  # Verbose cleanup without backups

EOF
}

# Function to parse command line arguments
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -n|--dry-run)
                DRY_RUN=1
                shift
                ;;
            -v|--verbose)
                VERBOSE=1
                shift
                ;;
            -q|--quiet)
                VERBOSE=0
                shift
                ;;
            --no-backup)
                BACKUP_ENABLED=0
                shift
                ;;
            --backup-only)
                BACKUP_ENABLED=1
                DRY_RUN=1
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
}

# Function to validate prerequisites
validate_prerequisites() {
    # Check if we're in a git repository
    if [[ ! -d ".git" ]]; then
        print_warning "Not in a git repository. Some cleanup operations may be limited."
    fi
    
    # Check if backup directory is writable
    if [[ ${BACKUP_ENABLED} -eq 1 ]]; then
        mkdir -p "${BACKUP_DIR}" 2>/dev/null || {
            print_error "Cannot create backup directory: ${BACKUP_DIR}"
            exit 1
        }
    fi
    
    # Check if log file is writable
    touch "${LOG_FILE}" 2>/dev/null || {
        print_error "Cannot write to log file: ${LOG_FILE}"
        exit 1
    }
}

# Function to show summary of operations
show_summary() {
    print_success "Cleanup completed successfully!"
    if [[ ${DRY_RUN} -eq 1 ]]; then
        print_info "This was a dry run. No files were actually removed."
    else
        print_info "Check ${LOG_FILE} for detailed operation log."
        if [[ ${BACKUP_ENABLED} -eq 1 ]]; then
            print_info "Backups are stored in ${BACKUP_DIR}"
        fi
    fi
}

# Main execution function
main() {
    # Parse command line arguments
    parse_arguments "$@"
    
    # Initialize log file
    echo "=== MQTT Orchestration Cleanup Log - $(date) ===" > "${LOG_FILE}"
    
    # Validate prerequisites
    validate_prerequisites
    
    # Print header
    print_info "Starting MQTT Orchestration Project Cleanup"
    print_info "Project root: ${PROJECT_ROOT}"
    
    # Execute cleanup operations
    remove_temp_files
    remove_log_files
    remove_build_artifacts
    remove_dependency_dirs
    remove_python_cache
    organize_structure
    setup_ci_cd
    cleanup_backup
    
    # Show summary
    show_summary
}

# Run main function with all arguments
main "$@"
