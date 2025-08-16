#!/bin/bash

# Quick Git-based Codebase Change Analysis
# Provides immediate insights using only git commands

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_usage() {
    echo -e "${BLUE}Quick Git-based Codebase Change Analysis${NC}"
    echo ""
    echo "Usage: $0 [OPTIONS] <repository_path>"
    echo ""
    echo "Options:"
    echo "  -t, --time-range <range>    Time range to analyze (default: '1 month ago')"
    echo "  -d, --detailed              Show detailed file-by-file breakdown"
    echo "  -h, --help                  Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 /path/to/repo"
    echo "  $0 -t '1 week ago' /path/to/repo"
    echo "  $0 --detailed /path/to/repo"
}

# Default values
TIME_RANGE="1 month ago"
DETAILED=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -t|--time-range)
            TIME_RANGE="$2"
            shift 2
            ;;
        -d|--detailed)
            DETAILED=true
            shift
            ;;
        -h|--help)
            print_usage
            exit 0
            ;;
        -*)
            echo -e "${RED}Error: Unknown option $1${NC}" >&2
            print_usage
            exit 1
            ;;
        *)
            REPO_PATH="$1"
            shift
            ;;
    esac
done

# Check if repository path is provided
if [[ -z "$REPO_PATH" ]]; then
    echo -e "${RED}Error: Repository path is required${NC}" >&2
    print_usage
    exit 1
fi

# Check if repository path exists and is a git repo
if [[ ! -d "$REPO_PATH/.git" ]]; then
    echo -e "${RED}Error: '$REPO_PATH' is not a Git repository${NC}" >&2
    exit 1
fi

echo -e "${GREEN}Quick Codebase Change Analysis${NC}"
echo -e "Repository: ${BLUE}$REPO_PATH${NC}"
echo -e "Time Range: ${BLUE}$TIME_RANGE${NC}"
echo ""

# Change to repository directory
cd "$REPO_PATH"

# Get repository name
REPO_NAME=$(basename "$(git config --get remote.origin.url 2>/dev/null | sed 's/\.git$//' | sed 's/.*\///')" 2>/dev/null || echo "unknown")

echo -e "${CYAN}=== Repository Overview ===${NC}"
echo -e "Repository: ${GREEN}$REPO_NAME${NC}"
echo -e "Current branch: ${GREEN}$(git branch --show-current)${NC}"
echo -e "Total commits: ${GREEN}$(git rev-list --count HEAD)${NC}"
echo -e "First commit: ${GREEN}$(git log --reverse --format=%cd --date=short | head -1)${NC}"
echo -e "Last commit: ${GREEN}$(git log -1 --format=%cd --date=short)${NC}"

echo ""
echo -e "${CYAN}=== Change Analysis (since $TIME_RANGE) ===${NC}"

# Get commit count in time range
COMMIT_COUNT=$(git log --since="$TIME_RANGE" --oneline | wc -l)
echo -e "Commits in time range: ${GREEN}$COMMIT_COUNT${NC}"

# Get unique authors
AUTHOR_COUNT=$(git log --since="$TIME_RANGE" --format=%an | sort | uniq | wc -l)
echo -e "Unique contributors: ${GREEN}$AUTHOR_COUNT${NC}"

# Get most active author
MOST_ACTIVE_AUTHOR=$(git log --since="$TIME_RANGE" --format=%an | sort | uniq -c | sort -nr | head -1 | sed 's/^ *//' | cut -d' ' -f2-)
MOST_ACTIVE_COUNT=$(git log --since="$TIME_RANGE" --format=%an | sort | uniq -c | sort -nr | head -1 | sed 's/^ *//' | cut -d' ' -f1)
echo -e "Most active contributor: ${GREEN}$MOST_ACTIVE_AUTHOR${NC} (${GREEN}$MOST_ACTIVE_COUNT${NC} commits)"

# Get files changed
FILES_CHANGED=$(git log --since="$TIME_RANGE" --name-only --pretty=format: | sort | uniq | grep -v '^$' | wc -l)
echo -e "Files changed: ${GREEN}$FILES_CHANGED${NC}"

# Get total lines changed
LINES_ADDED=$(git log --since="$TIME_RANGE" --numstat | awk '{sum+=$1} END {print sum+0}')
LINES_REMOVED=$(git log --since="$TIME_RANGE" --numstat | awk '{sum+=$2} END {print sum+0}')
echo -e "Lines added: ${GREEN}$LINES_ADDED${NC}"
echo -e "Lines removed: ${GREEN}$LINES_REMOVED${NC}"

# Get change frequency by directory
echo ""
echo -e "${CYAN}=== Directory Change Frequency ===${NC}"
git log --since="$TIME_RANGE" --name-only --pretty=format: | grep -v '^$' | while read file; do
    if [[ -n "$file" ]]; then
        dirname "$file"
    fi
done | sort | uniq -c | sort -nr | head -10 | while read count dir; do
    if [[ "$dir" == "." ]]; then
        echo -e "  ${GREEN}root${NC}: ${YELLOW}$count${NC} changes"
    else
        echo -e "  ${GREEN}$dir${NC}: ${YELLOW}$count${NC} changes"
    fi
done

# Get most changed files
echo ""
echo -e "${CYAN}=== Most Changed Files ===${NC}"
git log --since="$TIME_RANGE" --name-only --pretty=format: | grep -v '^$' | sort | uniq -c | sort -nr | head -10 | while read count file; do
    echo -e "  ${GREEN}$file${NC}: ${YELLOW}$count${NC} changes"
done

# Get file type distribution
echo ""
echo -e "${CYAN}=== File Type Distribution ===${NC}"
git log --since="$TIME_RANGE" --name-only --pretty=format: | grep -v '^$' | while read file; do
    if [[ -n "$file" ]]; then
        extension="${file##*.}"
        if [[ "$extension" == "$file" ]]; then
            echo "no-extension"
        else
            echo "$extension"
        fi
    fi
done | sort | uniq -c | sort -nr | head -10 | while read count ext; do
    echo -e "  ${GREEN}.$ext${NC}: ${YELLOW}$count${NC} files"
done

# Get commit activity over time
echo ""
echo -e "${CYAN}=== Commit Activity Timeline ===${NC}"
git log --since="$TIME_RANGE" --format=%cd --date=short | sort | uniq -c | tail -10 | while read count date; do
    echo -e "  ${GREEN}$date${NC}: ${YELLOW}$count${NC} commits"
done

# Detailed analysis if requested
if [[ "$DETAILED" == true ]]; then
    echo ""
    echo -e "${CYAN}=== Detailed File Analysis ===${NC}"
    
    # Get detailed stats for each file
    git log --since="$TIME_RANGE" --name-only --pretty=format: | grep -v '^$' | sort | uniq | while read file; do
        if [[ -n "$file" ]]; then
            # Count commits for this file
            file_commits=$(git log --since="$TIME_RANGE" --follow --oneline "$file" | wc -l)
            
            # Get authors for this file
            file_authors=$(git log --since="$TIME_RANGE" --follow --format=%an "$file" | sort | uniq | wc -l)
            
            # Get last modified
            last_modified=$(git log --since="$TIME_RANGE" --follow --format=%cd --date=short "$file" | head -1)
            
            echo -e "  ${GREEN}$file${NC}"
            echo -e "    Commits: ${YELLOW}$file_commits${NC}"
            echo -e "    Authors: ${YELLOW}$file_authors${NC}"
            echo -e "    Last modified: ${YELLOW}$last_modified${NC}"
            echo ""
        fi
    done
fi

# Get recent commit messages
echo ""
echo -e "${CYAN}=== Recent Commit Messages ===${NC}"
git log --since="$TIME_RANGE" --oneline -10 | while read hash message; do
    echo -e "  ${GREEN}$hash${NC} $message"
done

echo ""
echo -e "${GREEN}Analysis completed!${NC}"
