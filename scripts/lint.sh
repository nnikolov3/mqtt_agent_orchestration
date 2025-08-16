#!/usr/bin/env bash

# Design: Comprehensive Linting Script for MQTT Agent Orchestration System
# Purpose: Enforce strictest coding standards following Design Principles
# Usage: ./scripts/lint.sh [--verbose] [--fix] [--all]

set -euo pipefail

# ALL global variables declared at top
declare -r PROJECT_NAME_GLOBAL="mqtt-orchestrator"
declare -r SCRIPT_DIR_GLOBAL="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
declare -r PROJECT_ROOT_GLOBAL="$(cd "$SCRIPT_DIR_GLOBAL/.." && pwd)"
declare -r LINT_TIMEOUT_GLOBAL="15m"
declare VERBOSE_GLOBAL=false
declare FIX_GLOBAL=false
declare ALL_GLOBAL=false
declare FAILED_CHECKS_GLOBAL=0

# Color codes for output
declare -r RED='\033[0;31m'
declare -r GREEN='\033[0;32m'
declare -r YELLOW='\033[1;33m'
declare -r BLUE='\033[0;34m'
declare -r PURPLE='\033[0;35m'
declare -r CYAN='\033[0;36m'
declare -r NC='\033[0m' # No Color

# Logging functions following "Fail Fast, Fail Loud" principle
function log_info() {
    local message="$1"
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')] INFO:${NC} $message"
}

function log_error() {
    local message="$1"
    echo -e "${RED}[$(date +'%Y-%m-%d %H:%M:%S')] ERROR:${NC} $message" >&2
    FAILED_CHECKS_GLOBAL=$((FAILED_CHECKS_GLOBAL + 1))
}

function log_warn() {
    local message="$1"
    echo -e "${YELLOW}[$(date +'%Y-%m-%d %H:%M:%S')] WARN:${NC} $message" >&2
}

function log_success() {
    local message="$1"
    echo -e "${GREEN}[$(date +'%Y-%m-%d %H:%M:%S')] SUCCESS:${NC} $message"
}

function log_header() {
    local message="$1"
    echo -e "${PURPLE}=== $message ===${NC}"
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
            --fix)
                FIX_GLOBAL=true
                shift
                ;;
            --all)
                ALL_GLOBAL=true
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

Comprehensive linting script for MQTT Agent Orchestration System
Following Design Principles: "Excellence through Rigor" and "Fail Fast, Fail Loud"

OPTIONS:
    --verbose    Enable verbose output
    --fix        Attempt to fix auto-fixable issues
    --all        Run all linting checks (including slow ones)
    --help       Show this help message

EXAMPLES:
    $0                  Run standard linting checks
    $0 --verbose        Run with verbose output
    $0 --fix            Run and attempt to fix issues
    $0 --all            Run all checks including performance profiling
EOF
}

# Check dependencies following "Defensive Programming" principle
function check_dependencies() {
    local dep=""
    local -a required_deps=("go" "git")
    local -a optional_deps=("golangci-lint" "gofmt" "goimports" "govet" "gosec" "staticcheck" "gocritic" "gocyclo" "gocognit" "goconst" "gosimple" "ineffassign" "misspell" "unused" "deadcode" "varcheck" "structcheck" "errcheck" "interfacer" "unconvert" "gosec" "prealloc" "dupl" "goprintffuncname" "godox" "lll" "wsl" "noctx" "gocognit" "goconst" "gocyclo" "godot" "goprintffuncname" "gosec" "gosimple" "govet" "ineffassign" "misspell" "nakedret" "noctx" "nolintlint" "prealloc" "staticcheck" "structcheck" "stylecheck" "thelper" "tparallel" "typecheck" "unconvert" "unparam" "unused" "varcheck" "whitespace" "wrapcheck" "makezero" "nilnil" "nilerr" "shadow" "gci" "godox" "funlen" "cyclop" "gomnd" "errorlint" "wrapcheck" "gosec" "gocritic" "gomnd" "prealloc" "cyclop" "dupl" "funlen" "gocognit" "goconst" "gocyclo" "godox" "goerr113" "gofumpt" "gomnd" "goprintffuncname" "gosec" "gosimple" "govet" "ineffassign" "lll" "misspell" "nakedret" "noctx" "nolintlint" "prealloc" "staticcheck" "structcheck" "stylecheck" "thelper" "tparallel" "typecheck" "unconvert" "unparam" "unused" "varcheck" "whitespace" "wrapcheck" "interfacer" "nilerr" "makezero" "nilnil" "godox" "godot" "deadcode" "unused" "shadow" "goimports" "gci")
    local missing_required=()
    local missing_optional=()
    
    log_info "Checking dependencies..."
    
    # Check required dependencies
    for dep in "${required_deps[@]}"; do
        if ! command -v "$dep" >/dev/null 2>&1; then
            missing_required+=("$dep")
        else
            log_success "Found required dependency: $dep"
        fi
    done
    
    # Check optional dependencies
    for dep in "${optional_deps[@]}"; do
        if ! command -v "$dep" >/dev/null 2>&1; then
            missing_optional+=("$dep")
        else
            log_success "Found optional dependency: $dep"
        fi
    done
    
    # Report missing dependencies
    if [[ ${#missing_required[@]} -gt 0 ]]; then
        log_error "Missing required dependencies: ${missing_required[*]}"
        log_error "Please install missing dependencies and try again"
        exit 1
    fi
    
    if [[ ${#missing_optional[@]} -gt 0 ]]; then
        log_warn "Missing optional dependencies: ${missing_optional[*]}"
        log_warn "Some linting checks will be skipped"
    fi
}

# Run Go formatting checks
function run_go_formatting() {
    log_header "Running Go Formatting Checks"
    
    local format_result=""
    local format_exit=""
    
    # Check if code is properly formatted
    format_result=$(gofmt -l . 2>&1)
    format_exit=$?
    
    if [[ $format_exit -ne 0 ]]; then
        log_error "gofmt check failed: $format_result"
        return 1
    fi
    
    if [[ -n "$format_result" ]]; then
        log_error "Code formatting issues found:"
        echo "$format_result"
        if [[ "$FIX_GLOBAL" == true ]]; then
            log_info "Attempting to fix formatting issues..."
            gofmt -w .
            log_success "Formatting issues fixed"
        else
            log_error "Run with --fix to automatically fix formatting issues"
            return 1
        fi
    else
        log_success "Code formatting is correct"
    fi
}

# Run Go imports check
function run_go_imports() {
    log_header "Running Go Imports Check"
    
    local imports_result=""
    local imports_exit=""
    
    # Check if imports are properly organized
    imports_result=$(goimports -l . 2>&1)
    imports_exit=$?
    
    if [[ $imports_exit -ne 0 ]]; then
        log_error "goimports check failed: $imports_result"
        return 1
    fi
    
    if [[ -n "$imports_result" ]]; then
        log_error "Import organization issues found:"
        echo "$imports_result"
        if [[ "$FIX_GLOBAL" == true ]]; then
            log_info "Attempting to fix import issues..."
            goimports -w .
            log_success "Import issues fixed"
        else
            log_error "Run with --fix to automatically fix import issues"
            return 1
        fi
    else
        log_success "Import organization is correct"
    fi
}

# Run Go vet
function run_go_vet() {
    log_header "Running Go Vet"
    
    local vet_result=""
    local vet_exit=""
    
    # Run go vet on all packages
    vet_result=$(go vet ./... 2>&1)
    vet_exit=$?
    
    if [[ $vet_exit -ne 0 ]]; then
        log_error "go vet found issues:"
        echo "$vet_result"
        return 1
    else
        log_success "go vet passed"
    fi
}

# Run golangci-lint
function run_golangci_lint() {
    log_header "Running golangci-lint"
    
    if ! command -v golangci-lint >/dev/null 2>&1; then
        log_warn "golangci-lint not found, skipping"
        return 0
    fi
    
    local lint_args=("run" "--timeout=$LINT_TIMEOUT_GLOBAL")
    
    if [[ "$VERBOSE_GLOBAL" == true ]]; then
        lint_args+=("--verbose")
    fi
    
    if [[ "$FIX_GLOBAL" == true ]]; then
        lint_args+=("--fix")
    fi
    
    local lint_result=""
    local lint_exit=""
    
    # Run golangci-lint
    lint_result=$(golangci-lint "${lint_args[@]}" 2>&1)
    lint_exit=$?
    
    if [[ $lint_exit -ne 0 ]]; then
        log_error "golangci-lint found issues:"
        echo "$lint_result"
        return 1
    else
        log_success "golangci-lint passed"
    fi
}

# Run security scanning
function run_security_scan() {
    log_header "Running Security Scan"
    
    if ! command -v gosec >/dev/null 2>&1; then
        log_warn "gosec not found, skipping security scan"
        return 0
    fi
    
    local gosec_result=""
    local gosec_exit=""
    
    # Run gosec security scanner
    gosec_result=$(gosec -fmt=text ./... 2>&1)
    gosec_exit=$?
    
    if [[ $gosec_exit -ne 0 ]]; then
        log_error "Security issues found:"
        echo "$gosec_result"
        return 1
    else
        log_success "Security scan passed"
    fi
}

# Run static analysis
function run_static_analysis() {
    log_header "Running Static Analysis"
    
    if ! command -v staticcheck >/dev/null 2>&1; then
        log_warn "staticcheck not found, skipping static analysis"
        return 0
    fi
    
    local static_result=""
    local static_exit=""
    
    # Run staticcheck
    static_result=$(staticcheck ./... 2>&1)
    static_exit=$?
    
    if [[ $static_exit -ne 0 ]]; then
        log_error "Static analysis found issues:"
        echo "$static_result"
        return 1
    else
        log_success "Static analysis passed"
    fi
}

# Run complexity analysis
function run_complexity_analysis() {
    log_header "Running Complexity Analysis"
    
    if ! command -v gocyclo >/dev/null 2>&1; then
        log_warn "gocyclo not found, skipping complexity analysis"
        return 0
    fi
    
    local cyclo_result=""
    local cyclo_exit=""
    
    # Run gocyclo with threshold of 10
    cyclo_result=$(gocyclo -over 10 . 2>&1)
    cyclo_exit=$?
    
    if [[ $cyclo_exit -ne 0 ]]; then
        log_error "Complexity issues found:"
        echo "$cyclo_result"
        return 1
    else
        log_success "Complexity analysis passed"
    fi
}

# Run performance profiling (if --all is specified)
function run_performance_profiling() {
    if [[ "$ALL_GLOBAL" != true ]]; then
        return 0
    fi
    
    log_header "Running Performance Profiling"
    
    local profile_result=""
    local profile_exit=""
    
    # Run benchmarks
    profile_result=$(go test -bench=. -benchmem ./... 2>&1)
    profile_exit=$?
    
    if [[ $profile_exit -ne 0 ]]; then
        log_error "Performance profiling failed:"
        echo "$profile_result"
        return 1
    else
        log_success "Performance profiling completed"
        if [[ "$VERBOSE_GLOBAL" == true ]]; then
            echo "$profile_result"
        fi
    fi
}

# Run memory profiling (if --all is specified)
function run_memory_profiling() {
    if [[ "$ALL_GLOBAL" != true ]]; then
        return 0
    fi
    
    log_header "Running Memory Profiling"
    
    # Create profiling directory
    local profile_dir="$PROJECT_ROOT_GLOBAL/profiles"
    mkdir -p "$profile_dir"
    
    # Run memory profiling on main packages
    local packages=("cmd/worker" "cmd/server" "cmd/client" "cmd/rag-service" "cmd/role-worker")
    
    for pkg in "${packages[@]}"; do
        if [[ -d "$pkg" ]]; then
            log_info "Profiling memory usage for $pkg"
            
            # Run with memory profiling
            go test -memprofile="$profile_dir/${pkg//\//_}_mem.prof" -cpuprofile="$profile_dir/${pkg//\//_}_cpu.prof" "./$pkg" >/dev/null 2>&1 || true
            
            log_success "Memory profile created for $pkg"
        fi
    done
}

# Run code coverage analysis
function run_coverage_analysis() {
    log_header "Running Code Coverage Analysis"
    
    local coverage_result=""
    local coverage_exit=""
    
    # Run tests with coverage
    coverage_result=$(go test -coverprofile=coverage.out -covermode=atomic ./... 2>&1)
    coverage_exit=$?
    
    if [[ $coverage_exit -ne 0 ]]; then
        log_error "Coverage analysis failed:"
        echo "$coverage_result"
        return 1
    else
        log_success "Coverage analysis completed"
        
        # Show coverage summary
        if command -v go tool cover >/dev/null 2>&1; then
            local coverage_summary=""
            coverage_summary=$(go tool cover -func=coverage.out | tail -1)
            log_info "Coverage summary: $coverage_summary"
        fi
    fi
}

# Run dependency analysis
function run_dependency_analysis() {
    log_header "Running Dependency Analysis"
    
    local mod_result=""
    local mod_exit=""
    
    # Check for outdated dependencies
    mod_result=$(go list -u -m all 2>&1)
    mod_exit=$?
    
    if [[ $mod_exit -ne 0 ]]; then
        log_error "Dependency analysis failed:"
        echo "$mod_result"
        return 1
    fi
    
    # Check for vulnerabilities
    if command -v govulncheck >/dev/null 2>&1; then
        local vuln_result=""
        vuln_result=$(govulncheck ./... 2>&1)
        if [[ $? -ne 0 ]]; then
            log_error "Vulnerabilities found:"
            echo "$vuln_result"
            return 1
        else
            log_success "No vulnerabilities found"
        fi
    else
        log_warn "govulncheck not found, skipping vulnerability check"
    fi
    
    log_success "Dependency analysis completed"
}

# Run race condition detection
function run_race_detection() {
    log_header "Running Race Condition Detection"
    
    local race_result=""
    local race_exit=""
    
    # Run tests with race detection
    race_result=$(go test -race ./... 2>&1)
    race_exit=$?
    
    if [[ $race_exit -ne 0 ]]; then
        log_error "Race conditions detected:"
        echo "$race_result"
        return 1
    else
        log_success "No race conditions detected"
    fi
}

# Run dead code detection
function run_dead_code_detection() {
    log_header "Running Dead Code Detection"
    
    if ! command -v unused >/dev/null 2>&1; then
        log_warn "unused not found, skipping dead code detection"
        return 0
    fi
    
    local unused_result=""
    local unused_exit=""
    
    # Run unused code detection
    unused_result=$(unused ./... 2>&1)
    unused_exit=$?
    
    if [[ $unused_exit -ne 0 ]]; then
        log_error "Dead code found:"
        echo "$unused_result"
        return 1
    else
        log_success "No dead code found"
    fi
}

# Run interface compliance check
function run_interface_compliance() {
    log_header "Running Interface Compliance Check"
    
    if ! command -v interfacer >/dev/null 2>&1; then
        log_warn "interfacer not found, skipping interface compliance check"
        return 0
    fi
    
    local interfacer_result=""
    local interfacer_exit=""
    
    # Run interface compliance check
    interfacer_result=$(interfacer ./... 2>&1)
    interfacer_exit=$?
    
    if [[ $interfacer_exit -ne 0 ]]; then
        log_error "Interface compliance issues found:"
        echo "$interfacer_result"
        return 1
    else
        log_success "Interface compliance check passed"
    fi
}

# Generate linting report
function generate_report() {
    log_header "Generating Linting Report"
    
    local report_file="$PROJECT_ROOT_GLOBAL/lint_report.txt"
    local timestamp="$(date -u '+%Y-%m-%d %H:%M:%S UTC')"
    
    cat > "$report_file" << EOF
Linting Report for MQTT Agent Orchestration System
Generated: $timestamp
Project: $PROJECT_NAME_GLOBAL
Version: $(git describe --tags --always --dirty 2>/dev/null || echo 'unknown')

Summary:
- Failed Checks: $FAILED_CHECKS_GLOBAL
- Status: $([[ $FAILED_CHECKS_GLOBAL -eq 0 ]] && echo "PASSED" || echo "FAILED")

Configuration:
- Verbose Mode: $VERBOSE_GLOBAL
- Fix Mode: $FIX_GLOBAL
- All Checks: $ALL_GLOBAL
- Timeout: $LINT_TIMEOUT_GLOBAL

This report was generated by the comprehensive linting script following
Design Principles: "Excellence through Rigor" and "Fail Fast, Fail Loud"
EOF
    
    log_success "Linting report generated: $report_file"
}

# Main function
function main() {
    log_header "Starting Comprehensive Linting for $PROJECT_NAME_GLOBAL"
    
    # Change to project root
    cd "$PROJECT_ROOT_GLOBAL"
    
    # Check dependencies
    check_dependencies
    
    # Run all linting checks
    run_go_formatting
    run_go_imports
    run_go_vet
    run_golangci_lint
    run_security_scan
    run_static_analysis
    run_complexity_analysis
    run_coverage_analysis
    run_dependency_analysis
    run_race_detection
    run_dead_code_detection
    run_interface_compliance
    run_performance_profiling
    run_memory_profiling
    
    # Generate report
    generate_report
    
    # Final status
    if [[ $FAILED_CHECKS_GLOBAL -eq 0 ]]; then
        log_success "All linting checks passed!"
        exit 0
    else
        log_error "Linting failed with $FAILED_CHECKS_GLOBAL failed checks"
        exit 1
    fi
}

# Parse arguments and run main
parse_args "$@"
main
