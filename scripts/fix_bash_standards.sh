#!/usr/bin/env bash

# Design: Clean bash code fixer using AI helpers
# Purpose: Extract only the corrected bash code, no explanations
# Usage: ./scripts/fix_bash_standards.sh <input_file> <output_file>

set -euo pipefail

# ALL global variables declared at top
declare input_file=""
declare output_file=""
declare temp_response=""
declare cerebras_result=""
declare cerebras_exit=""

input_file="$1"
output_file="$2"
temp_response=$(mktemp)

# Use Cerebras to fix bash standards, extract only the code block
cerebras_result=$(cerebras_code_analyzer "Fix ALL bash coding standard violations following BASH_CODING_STANDARD_CLAUDE.md strictly: 1) ALL local variables declared at function top 2) NEVER use /dev/null redirections 3) NEVER suppress error output 4) Explicit exit code capture in variables 5) Use proper command testing without redirections 6) No 2>&1 unless capturing errors. Output ONLY the corrected bash script between \`\`\`bash and \`\`\`, no explanations." "$input_file" 2>&1)
cerebras_exit="$?"

if [[ "$cerebras_exit" -ne 0 ]]; then
    echo "ERROR: Cerebras failed: $cerebras_result" >&2
    exit 1
fi

# Extract just the bash code block and write to output
echo "$cerebras_result" | sed -n '/```bash/,/```/p' | sed '1d;$d' > "$output_file"

# Cleanup
rm -f "$temp_response"

echo "Fixed bash script written to: $output_file"