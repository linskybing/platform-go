#!/bin/bash

# Skills Management Utility
# Purpose: Manage, validate, and organize platform-go skills
# Usage: ./scripts/skills-manager.sh [command] [options]

set -u

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

SKILLS_DIR=".github/skills"
FRONTEND_SKILLS_DIR="frontend-go/.github/skills"

# ==============================================================================
# Helper Functions
# ==============================================================================

log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

# ==============================================================================
# Validation Functions
# ==============================================================================

validate_skill_frontmatter() {
    local skill_file=$1
    local skill_name=$(basename "$(dirname "$skill_file")")
    
    if [ ! -f "$skill_file" ]; then
        log_error "SKILL.md not found: $skill_file"
        return 1
    fi
    
    # Check if file starts with frontmatter
    if ! head -1 "$skill_file" | grep -q "^---$"; then
        log_error "$skill_name: Missing frontmatter start"
        return 1
    fi
    
    # Extract frontmatter using awk
    local frontmatter=$(awk '/^---$/{if(++count==2) exit; next} count==1' "$skill_file")
    
    if [ -z "$frontmatter" ]; then
        log_error "$skill_name: Empty frontmatter"
        return 1
    fi
    
    if ! echo "$frontmatter" | grep -q "^name:"; then
        log_error "$skill_name: Missing 'name' field"
        return 1
    fi
    
    if ! echo "$frontmatter" | grep -q "^description:"; then
        log_error "$skill_name: Missing 'description' field"
        return 1
    fi
    
    local name=$(echo "$frontmatter" | grep "^name:" | sed 's/name: //;s/ *$//')
    if ! echo "$name" | grep -qE '^[a-z0-9]([a-z0-9-]*[a-z0-9])?$'; then
        log_error "$skill_name: Invalid name format: $name"
        return 1
    fi
    
    local description=$(echo "$frontmatter" | grep "^description:" | sed 's/description: //')
    if [ ${#description} -gt 1024 ]; then
        log_error "$skill_name: Description too long"
        return 1
    fi
    
    return 0
}

# ==============================================================================
# List Commands
# ==============================================================================

list_all_skills() {
    log_info "Platform-go Skills:"
    echo ""
    
    local count=0
    for skill_file in $SKILLS_DIR/*/SKILL.md; do
        if [ -f "$skill_file" ]; then
            local skill_name=$(basename "$(dirname "$skill_file")")
            local description=$(sed -n 's/^description: //p' "$skill_file" | head -1)
            
            printf "  ${BLUE}%-40s${NC} %s\n" "$skill_name" "$description"
            count=$((count + 1))
        fi
    done
    
    echo ""
    if [ -d "$FRONTEND_SKILLS_DIR" ]; then
        log_info "Frontend Skills:"
        echo ""
        
        for skill_file in $FRONTEND_SKILLS_DIR/*/SKILL.md; do
            if [ -f "$skill_file" ]; then
                local skill_name=$(basename "$(dirname "$skill_file")")
                local description=$(sed -n 's/^description: //p' "$skill_file" | head -1)
                
                printf "  ${BLUE}%-40s${NC} %s\n" "$skill_name" "$description"
                count=$((count + 1))
            fi
        done
        echo ""
    fi
    
    log_success "Found $count total skills"
}

# ==============================================================================
# Validation Commands
# ==============================================================================

validate_all_skills() {
    log_info "Validating all SKILL.md files..."
    echo ""
    
    local total=0
    local valid=0
    local invalid=0
    
    for skill_file in $SKILLS_DIR/*/SKILL.md; do
        if [ -f "$skill_file" ]; then
            total=$((total + 1))
            if validate_skill_frontmatter "$skill_file"; then
                local skill_name=$(basename "$(dirname "$skill_file")")
                log_success "$skill_name"
                valid=$((valid + 1))
            else
                invalid=$((invalid + 1))
            fi
        fi
    done
    
    if [ -d "$FRONTEND_SKILLS_DIR" ]; then
        for skill_file in $FRONTEND_SKILLS_DIR/*/SKILL.md; do
            if [ -f "$skill_file" ]; then
                total=$((total + 1))
                if validate_skill_frontmatter "$skill_file"; then
                    local skill_name=$(basename "$(dirname "$skill_file")")
                    log_success "$skill_name"
                    valid=$((valid + 1))
                else
                    invalid=$((invalid + 1))
                fi
            fi
        done
    fi
    
    echo ""
    log_info "Validation Results: $valid/$total valid"
    
    if [ "$invalid" -eq 0 ]; then
        log_success "All skills are valid!"
        return 0
    else
        log_error "$invalid skill(s) failed validation"
        return 1
    fi
}

validate_skill() {
    local skill_name=$1
    
    if [ -z "$skill_name" ]; then
        log_error "Skill name required"
        return 1
    fi
    
    local skill_file="$SKILLS_DIR/$skill_name/SKILL.md"
    
    if [ ! -f "$skill_file" ]; then
        log_error "Skill not found: $skill_name"
        return 1
    fi
    
    log_info "Validating $skill_name..."
    
    if validate_skill_frontmatter "$skill_file"; then
        log_success "$skill_name is valid"
        return 0
    else
        return 1
    fi
}

# ==============================================================================
# Statistics Commands
# ==============================================================================

show_stats() {
    log_info "Skills Statistics"
    echo ""
    
    local total=0
    local backend=0
    local frontend=0
    local total_lines=0
    
    for skill_file in $SKILLS_DIR/*/SKILL.md; do
        if [ -f "$skill_file" ]; then
            total=$((total + 1))
            backend=$((backend + 1))
            local lines=$(wc -l < "$skill_file")
            total_lines=$((total_lines + lines))
        fi
    done
    
    if [ -d "$FRONTEND_SKILLS_DIR" ]; then
        for skill_file in $FRONTEND_SKILLS_DIR/*/SKILL.md; do
            if [ -f "$skill_file" ]; then
                total=$((total + 1))
                frontend=$((frontend + 1))
                local lines=$(wc -l < "$skill_file")
                total_lines=$((total_lines + lines))
            fi
        done
    fi
    
    echo "Total Skills:                    $total"
    echo "  Backend Skills:                $backend"
    echo "  Frontend Skills:               $frontend"
    echo "Total Documentation Lines:       $total_lines"
    
    if [ "$total" -gt 0 ]; then
        local avg=$((total_lines / total))
        echo "Average Lines/Skill:            $avg"
    fi
}

# ==============================================================================
# Search Commands
# ==============================================================================

search_skills() {
    local query=$1
    
    if [ -z "$query" ]; then
        log_error "Search query required"
        return 1
    fi
    
    log_info "Searching for '$query'..."
    echo ""
    
    local found=0
    
    for skill_file in $SKILLS_DIR/*/SKILL.md; do
        if [ -f "$skill_file" ]; then
            local skill_name=$(basename "$(dirname "$skill_file")")
            if grep -qi "$query" "$skill_file"; then
                local matches=$(grep -ci "$query" "$skill_file")
                log_success "$skill_name ($matches matches)"
                found=$((found + 1))
            fi
        fi
    done
    
    if [ -d "$FRONTEND_SKILLS_DIR" ]; then
        for skill_file in $FRONTEND_SKILLS_DIR/*/SKILL.md; do
            if [ -f "$skill_file" ]; then
                local skill_name=$(basename "$(dirname "$skill_file")")
                if grep -qi "$query" "$skill_file"; then
                    local matches=$(grep -ci "$query" "$skill_file")
                    log_success "$skill_name ($matches matches)"
                    found=$((found + 1))
                fi
            fi
        done
    fi
    
    echo ""
    if [ "$found" -eq 0 ]; then
        log_warning "No skills found matching '$query'"
        return 1
    fi
    
    log_success "Found in $found skill(s)"
}

# ==============================================================================
# Helper Commands
# ==============================================================================

show_skill_content() {
    local skill_name=$1
    
    if [ -z "$skill_name" ]; then
        log_error "Skill name required"
        return 1
    fi
    
    local skill_file="$SKILLS_DIR/$skill_name/SKILL.md"
    
    if [ ! -f "$skill_file" ]; then
        log_error "Skill not found: $skill_name"
        return 1
    fi
    
    log_info "Content of $skill_name:"
    echo ""
    head -50 "$skill_file"
}

# ==============================================================================
# Generate Commands
# ==============================================================================

generate_skills_index() {
    local output_file="docs/SKILLS_INDEX.md"
    
    log_info "Generating skills index to $output_file..."
    
    {
        echo "# Platform-go Skills Index"
        echo ""
        echo "Complete reference of all available skills for the platform-go project."
        echo ""
        echo "## Backend Skills"
        echo ""
        
        for skill_file in $SKILLS_DIR/*/SKILL.md; do
            if [ -f "$skill_file" ]; then
                local skill_dir=$(dirname "$skill_file")
                local skill_name=$(basename "$skill_dir")
                local description=$(sed -n 's/^description: //p' "$skill_file" | head -1)
                
                echo "### $skill_name"
                echo ""
                echo "$description"
                echo ""
                echo "[View Full Skill]($SKILLS_DIR/$skill_name/SKILL.md)"
                echo ""
            fi
        done
        
        if [ -d "$FRONTEND_SKILLS_DIR" ]; then
            echo "## Frontend Skills"
            echo ""
            
            for skill_file in $FRONTEND_SKILLS_DIR/*/SKILL.md; do
                if [ -f "$skill_file" ]; then
                    local skill_dir=$(dirname "$skill_file")
                    local skill_name=$(basename "$skill_dir")
                    local description=$(sed -n 's/^description: //p' "$skill_file" | head -1)
                    
                    echo "### $skill_name"
                    echo ""
                    echo "$description"
                    echo ""
                    echo "[View Full Skill]($FRONTEND_SKILLS_DIR/$skill_name/SKILL.md)"
                    echo ""
                fi
            done
        fi
    } > "$output_file"
    
    log_success "Skills index generated: $output_file"
}

# ==============================================================================
# Main Help
# ==============================================================================

show_help() {
    cat << 'EOF'
Skills Manager - Manage platform-go Skills

USAGE:
    ./scripts/skills-manager.sh [command] [options]

COMMANDS:
    list                    List all available skills
    validate               Validate all skills format
    validate <skill>       Validate specific skill
    stats                  Show skills statistics
    search <query>         Search skills by keyword
    show <skill>           Show first 50 lines of skill content
    generate-index         Generate SKILLS_INDEX.md

EXAMPLES:
    ./scripts/skills-manager.sh list
    ./scripts/skills-manager.sh validate
    ./scripts/skills-manager.sh validate api-design-patterns
    ./scripts/skills-manager.sh search authentication
    ./scripts/skills-manager.sh show error-handling-guide
    ./scripts/skills-manager.sh generate-index
    ./scripts/skills-manager.sh stats

SKILL LOCATION:
    Backend:  .github/skills/
    Frontend: frontend-go/.github/skills/

For more information, visit:
    https://agentskills.io/specification

EOF
}

# ==============================================================================
# Main Entry Point
# ==============================================================================

main() {
    local command="${1:-help}"
    
    case "$command" in
        list)
            list_all_skills
            ;;
        validate)
            if [ -n "${2:-}" ]; then
                validate_skill "$2"
            else
                validate_all_skills
            fi
            ;;
        stats)
            show_stats
            ;;
        search)
            if [ -z "${2:-}" ]; then
                log_error "Search query required"
                exit 1
            fi
            search_skills "$2"
            ;;
        show)
            if [ -z "${2:-}" ]; then
                log_error "Skill name required"
                exit 1
            fi
            show_skill_content "$2"
            ;;
        generate-index)
            generate_skills_index
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "Unknown command: $command"
            echo ""
            show_help
            exit 1
            ;;
    esac
}

main "$@"
