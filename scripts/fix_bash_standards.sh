#!/usr/bin/env bash

# Design: Bash Standards Compliance Checker and Fixer using AI helpers
# Purpose: Check and fix bash scripts to comply with BASH_CODING_STANDARD_CLAUDE.md
# Usage: ./scripts/fix_bash_standards.sh [--check|--fix] [--target script.sh] [--all]

set -euo pipefail

# ALL global variables declared at top
declare -r SCRIPT_NAME_GLOBAL="$(basename "${BASH_SOURCE[0]}")"
declare -r PROJECT_ROOT_GLOBAL="$(cd "$(dirname "$(dirname "${BASH_SOURCE[0]}")")" && pwd)"
declare MODE_GLOBAL="check"
declare TARGET_SCRIPT_GLOBAL=""
declare PROCESS_ALL_GLOBAL=false

# Parse command line arguments
function parse_args() {
    local arg=""
    
    if [[ $# -eq 0 ]]; then
        show_usage
        exit 1
    fi
    
    while [[ $# -gt 0 ]]; do
        arg="$1"
        case "$arg" in
            --check)
                MODE_GLOBAL="check"
                shift
                ;;
            --fix)
                MODE_GLOBAL="fix"
                shift
                ;;
            --target)
                if [[ -n "${2:-}" ]]; then
                    TARGET_SCRIPT_GLOBAL="$2"
                    shift 2
                else
                    echo "ERROR: Target parameter requires a script path" >&2
                    exit 1
                fi
                ;;
            --all)
                PROCESS_ALL_GLOBAL=true
                shift
                ;;
            --help)
                show_usage
                exit 0
                ;;
            *)
                if [[ -z "$TARGET_SCRIPT_GLOBAL" ]] && [[ -f "$arg" ]]; then
                    # Legacy mode: treat first argument as input file
                    TARGET_SCRIPT_GLOBAL="$arg"
                    if [[ -n "${2:-}" ]]; then
                        # Legacy mode: second argument as output file
                        MODE_GLOBAL="fix"
                        shift 2
                    else
                        shift
                    fi
                else
                    echo "ERROR: Unknown option: $arg" >&2
                    show_usage
                    exit 1
                fi
                ;;
        esac
    done
}

# Show usage information
function show_usage() {
    cat << EOF
Usage: $SCRIPT_NAME_GLOBAL [OPTIONS] [legacy: input_file output_file]

Bash Standards Compliance Checker and Fixer using AI helpers

OPTIONS:
    --check             Check compliance (default)
    --fix               Fix issues automatically using AI helpers
    --target SCRIPT     Target specific script file
    --all               Process all bash scripts in project
    --help              Show this help message

LEGACY USAGE:
    $SCRIPT_NAME_GLOBAL input_file output_file    Fix input_file and save to output_file

EXAMPLES:
    $SCRIPT_NAME_GLOBAL --check --target script.sh
    $SCRIPT_NAME_GLOBAL --fix --target script.sh  
    $SCRIPT_NAME_GLOBAL --check --all
    $SCRIPT_NAME_GLOBAL script.sh fixed_script.sh    (legacy mode)

EOF
}

# Check script for compliance issues
function check_script() {
    local script_path="$1"
    local issues_found=0
    
    echo "Checking: $script_path"
    
    # Use Cerebras to check compliance
    local cerebras_result=""
    local cerebras_exit=""
    
    cerebras_result=$(cerebras_code_analyzer "Analyze this bash script for compliance with BASH_CODING_STANDARD_CLAUDE.md. Report specific violations found: 1) Missing variable declarations at top 2) Improper error handling 3) Unquoted variables 4) Missing exit code checks. Provide a concise compliance report." "$script_path" 2>&1)
    cerebras_exit="$?"
    
    if [[ "$cerebras_exit" -ne 0 ]]; then
        echo "  ERROR: Analysis failed: $cerebras_result"
        return 1
    fi
    
    echo "  Analysis result:"
    echo "$cerebras_result" | sed 's/^/    /'
    
    # Simple heuristic to detect if issues were found
    if echo "$cerebras_result" | grep -qi "violation\|issue\|problem\|fix\|error"; then
        issues_found=1
    fi
    
    return "$issues_found"
}

# Fix script using AI helper
function fix_script() {
    local input_file="$1"
    local output_file="${2:-}"
    local backup_file=""
    local cerebras_result=""
    local cerebras_exit=""
    local temp_output=""
    
    if [[ -z "$output_file" ]]; then
        # In-place fixing with backup
        backup_file="${input_file}.backup.$(date +%Y%m%d_%H%M%S)"
        output_file="$input_file"
        
        echo "Creating backup: $backup_file"
        cp "$input_file" "$backup_file"
    fi
    
    echo "Fixing: $input_file"
    
    temp_output=$(mktemp)
    
    # Use Cerebras to fix bash standards
    cerebras_result=$(cerebras_code_analyzer "Fix ALL bash coding standard violations following BASH_CODING_STANDARD_CLAUDE.md strictly: 1) ALL local variables declared at function top 2) Proper error handling with set -euo pipefail 3) Quote all variables 4) Explicit exit code capture 5) Proper function structure. Output ONLY the corrected bash script between \`\`\`bash and \`\`\`, no explanations." "$input_file" 2>&1)
    cerebras_exit="$?"
    
    if [[ "$cerebras_exit" -ne 0 ]]; then
        echo "  ERROR: Fix failed: $cerebras_result"
        rm -f "$temp_output"
        return 1
    fi
    
    # Extract just the bash code block
    echo "$cerebras_result" | sed -n '/```bash/,/```/p' | sed '1d;$d' > "$temp_output"
    
    # Check if we got valid output
    if [[ ! -s "$temp_output" ]]; then
        echo "  ERROR: No valid bash code extracted from AI response"
        rm -f "$temp_output"
        return 1
    fi
    
    # Move to final location
    mv "$temp_output" "$output_file"
    
    echo "  Fixed script written to: $output_file"
    if [[ -n "$backup_file" ]]; then
        echo "  Backup available at: $backup_file"
    fi
    
    return 0
}

# Find all bash scripts in project
function find_bash_scripts() {
    local -a scripts=()
    local script=""
    
    # Find .sh files and executable scripts with bash shebang
    while IFS= read -r -d '' script; do
        if [[ -f "$script" ]] && [[ -r "$script" ]]; then
            # Check if it's a bash script
            local first_line=""
            first_line=$(head -n 1 "$script" 2>/dev/null || echo "")
            if [[ "$script" =~ \.sh$ ]] || [[ "$first_line" =~ bash$ ]]; then
                scripts+=("$script")
            fi
        fi
    done < <(find "$PROJECT_ROOT_GLOBAL" -type f \( -name "*.sh" -o -executable \) -print0 2>/dev/null)
    
    printf '%s\n' "${scripts[@]}"
}

# Process all scripts
function process_all_scripts() {
    local mode="$1"
    local -a scripts=()
    local script=""
    local total=0
    local success=0
    local issues=0
    
    mapfile -t scripts < <(find_bash_scripts)
    total=${#scripts[@]}
    
    echo "Found $total bash scripts to process"
    echo ""
    
    for script in "${scripts[@]}"; do
        if [[ "$mode" == "check" ]]; then
            if check_script "$script"; then
                success=$((success + 1))
            else
                issues=$((issues + 1))
            fi
        elif [[ "$mode" == "fix" ]]; then
            if fix_script "$script"; then
                success=$((success + 1))
            else
                issues=$((issues + 1))
            fi
        fi
        echo ""
    done
    
    echo "Summary:"
    echo "  Total scripts: $total"
    echo "  Successful: $success"
    echo "  With issues: $issues"
}

# Main execution function
function main() {
    parse_args "$@"
    
    # Check if cerebras is available
    if ! command -v cerebras_code_analyzer; then
        echo "ERROR: cerebras_code_analyzer not found in PATH"
        echo "Please ensure AI helper tools are available"
        exit 1
    fi
    
    if [[ "$PROCESS_ALL_GLOBAL" == true ]]; then
        process_all_scripts "$MODE_GLOBAL"
    elif [[ -n "$TARGET_SCRIPT_GLOBAL" ]]; then
        if [[ ! -f "$TARGET_SCRIPT_GLOBAL" ]]; then
            echo "ERROR: Target script not found: $TARGET_SCRIPT_GLOBAL"
            exit 1
        fi
        
        if [[ "$MODE_GLOBAL" == "check" ]]; then
            if check_script "$TARGET_SCRIPT_GLOBAL"; then
                echo "Script is compliant"
                exit 0
            else
                echo "Script has compliance issues"
                exit 1
            fi
        elif [[ "$MODE_GLOBAL" == "fix" ]]; then
            if fix_script "$TARGET_SCRIPT_GLOBAL"; then
                echo "Script fixed successfully"
                exit 0
            else
                echo "Failed to fix script"
                exit 1
            fi
        fi
    else
        echo "ERROR: No target specified. Use --target, --all, or provide file arguments"
        show_usage
        exit 1
    fi
}

# Run main function
main "$@"