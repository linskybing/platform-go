#!/bin/bash

# Production standards validation
# Used by: golang-production-standards skill
# Location: .github/skills/golang-production-standards/scripts/lint-check.sh
#
# Runs comprehensive linting and code quality checks

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo -e "${CYAN}=== Production Standards Lint Check ${NC}"
echo -e "${CYAN}(golang-production-standards) ${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

failed=0
warnings=0

# 1. gofmt
echo -e "${YELLOW}[1/4] gofmt - Code formatting...${NC}"
if gofmt -l . 2>/dev/null | grep -v ".git" | grep "\.go$" | head -5 > /tmp/fmt-check.txt; then
    if [ -s /tmp/fmt-check.txt ]; then
        echo -e "${RED}✗ Format issues found${NC}"
        cat /tmp/fmt-check.txt | sed 's/^/  - /'
        failed=1
    else
        echo -e "${GREEN}✓ Format OK${NC}"
    fi
else
    echo -e "${GREEN}✓ Format OK${NC}"
fi

# 2. go vet
echo ""
echo -e "${YELLOW}[2/4] go vet - Static analysis...${NC}"
if go vet ./... 2>&1 | tee /tmp/vet-check.txt > /dev/null; then
    echo -e "${GREEN}✓ Vet OK${NC}"
else
    if grep -q "error\|vet found" /tmp/vet-check.txt; then
        echo -e "${RED}✗ Vet issues found${NC}"
        head -10 /tmp/vet-check.txt | sed 's/^/  - /'
        failed=1
    else
        echo -e "${GREEN}✓ Vet OK${NC}"
    fi
fi

# 3. golangci-lint
echo ""
echo -e "${YELLOW}[3/4] golangci-lint - Comprehensive linting...${NC}"
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run --timeout=5m ./... > /tmp/lint-check.txt 2>&1; then
        echo -e "${GREEN}✓ Lint OK${NC}"
    else
        lint_count=$(grep -c "error:" /tmp/lint-check.txt || echo "0")
        if [ "$lint_count" -gt 10 ]; then
            echo -e "${RED}✗ ${lint_count} lint issues found${NC}"
            head -10 /tmp/lint-check.txt | sed 's/^/  - /'
            failed=1
        else
            echo -e "${YELLOW}⚠ ${lint_count} lint issues found (reviewing)${NC}"
            cat /tmp/lint-check.txt | head -10 | sed 's/^/  - /'
            warnings=$((warnings + 1))
        fi
    fi
else
    echo -e "${YELLOW}⚠ golangci-lint not installed${NC}"
    echo "  Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    warnings=$((warnings + 1))
fi

# 4. Check for production antipatterns
echo ""
echo -e "${YELLOW}[4/4] Production antipattern checks...${NC}"

# Check for log.Fatal in non-main
fatal_in_lib=$(grep -r "log\.Fatal\|panic" --include="*.go" ./internal ./pkg 2>/dev/null | grep -v "test\|_test.go" | grep -v "main.go" | wc -l)
if [ "$fatal_in_lib" -gt 0 ]; then
    echo -e "${RED}✗ Found ${fatal_in_lib} log.Fatal/panic in library code${NC}"
    echo "  Should use return errors instead"
    failed=1
fi

# Check for TODOs
todos=$(grep -r "TODO\|FIXME\|XXX\|HACK" --include="*.go" ./internal ./pkg ./cmd 2>/dev/null | grep -v "test" | wc -l)
if [ "$todos" -gt 0 ]; then
    echo -e "${YELLOW}⚠ Found ${todos} TODO/FIXME comments${NC}"
    warnings=$((warnings + 1))
else
    echo -e "${GREEN}✓ No outstanding TODOs${NC}"
fi

# Check for magic numbers
magic=$(grep -rE "if.*==[[:space:]]*[0-9]|:=[[:space:]]*[0-9]{2}" --include="*.go" ./internal ./pkg 2>/dev/null | grep -v "test\|const\|var" | wc -l)
if [ "$magic" -gt 5 ]; then
    echo -e "${YELLOW}⚠ Found ${magic} potential magic numbers${NC}"
    warnings=$((warnings + 1))
fi

echo -e "${GREEN}✓ Antipattern check complete${NC}"

# Summary
echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}Lint Check Summary${NC}"
echo -e "${CYAN}========================================${NC}"

if [ $failed -eq 0 ]; then
    if [ $warnings -eq 0 ]; then
        echo -e "${GREEN}✓ All checks passed!${NC}"
        exit 0
    else
        echo -e "${YELLOW}✓ Passed with ${warnings} warning(s)${NC}"
        exit 0
    fi
else
    echo -e "${RED}✗ ${failed} check(s) failed${NC}"
    echo "Fix issues above and run again"
    exit 1
fi
