#!/bin/bash

# Skills Formatter - Convert existing skills to Agent Skills standard format
# Usage: ./scripts/format-skills.sh [skill-name]

set -u

SKILLS_DIR=".github/skills"
FRONTEND_SKILLS_DIR="frontend-go/.github/skills"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${BLUE}ℹ${NC} $1"; }
log_success() { echo -e "${GREEN}✓${NC} $1"; }
log_warning() { echo -e "${YELLOW}⚠${NC} $1"; }
log_error() { echo -e "${RED}✗${NC} $1"; }

# Format a single skill file to Agent Skills standard
format_skill() {
    local skill_file=$1
    local skill_name=$(basename "$(dirname "$skill_file")")
    
    if [ ! -f "$skill_file" ]; then
        log_error "Skill not found: $skill_file"
        return 1
    fi
    
    log_info "Formatting $skill_name..."
    
    # Extract existing metadata or create defaults
    local name=$skill_name
    local description=$(sed -n 's/^description: //p' "$skill_file" | head -1)
    
    if [ -z "$description" ]; then
        description=$(sed -n '/^# /p' "$skill_file" | head -1 | sed 's/^# //')
        if [ -z "$description" ]; then
            description="Implementation guide for $skill_name"
        fi
    fi
    
    # Limit description to 1024 chars
    if [ ${#description} -gt 1024 ]; then
        description="${description:0:1021}..."
    fi
    
    # Extract body (everything after first ---...--- block)
    local body=$(sed -n '/^---$/,/^---$/d; p' "$skill_file")
    
    # Create new formatted file
    local temp_file="${skill_file}.tmp"
    
    {
        echo "---"
        echo "name: $name"
        echo "description: $description"
        echo "license: Proprietary"
        echo "metadata:"
        echo "  author: platform-go"
        echo "  version: \"1.0\""
        echo "---"
        echo ""
        echo "$body"
    } > "$temp_file"
    
    # Replace original
    mv "$temp_file" "$skill_file"
    log_success "$skill_name formatted successfully"
    
    return 0
}

# Format all skills
format_all_skills() {
    log_info "Formatting all skills to Agent Skills standard format..."
    echo ""
    
    local total=0
    local success=0
    local failed=0
    
    # Backend skills
    for skill_file in $SKILLS_DIR/*/SKILL.md; do
        if [ -f "$skill_file" ]; then
            total=$((total + 1))
            if format_skill "$skill_file"; then
                success=$((success + 1))
            else
                failed=$((failed + 1))
            fi
        fi
    done
    
    # Frontend skills
    if [ -d "$FRONTEND_SKILLS_DIR" ]; then
        for skill_file in $FRONTEND_SKILLS_DIR/*/SKILL.md; do
            if [ -f "$skill_file" ]; then
                total=$((total + 1))
                if format_skill "$skill_file"; then
                    success=$((success + 1))
                else
                    failed=$((failed + 1))
                fi
            fi
        done
    fi
    
    echo ""
    if [ "$failed" -eq 0 ]; then
        log_success "All $success skills formatted successfully!"
        return 0
    else
        log_error "$failed skill(s) failed to format"
        return 1
    fi
}

show_help() {
    cat << 'EOF'
Skills Formatter - Convert skills to Agent Skills standard format

USAGE:
    ./scripts/format-skills.sh [skill-name|all]

COMMANDS:
    all                    Format all skills (default)
    <skill-name>          Format specific skill

EXAMPLES:
    ./scripts/format-skills.sh                    # Format all
    ./scripts/format-skills.sh all                # Format all
    ./scripts/format-skills.sh api-design-patterns # Format specific

This script converts existing SKILL.md files to follow the Agent Skills standard:
- Proper YAML frontmatter with required fields
- Correct field ordering
- Metadata section
- Preserved body content

For details, visit: https://agentskills.io/specification

EOF
}

main() {
    local target="${1:-all}"
    
    case "$target" in
        all)
            format_all_skills
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            format_skill "$SKILLS_DIR/$target/SKILL.md"
            ;;
    esac
}

main "$@"
