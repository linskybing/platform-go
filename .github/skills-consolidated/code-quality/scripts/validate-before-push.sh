#!/bin/bash

# Comprehensive pre-push validation script
# Used by: code-validation-standards skill
# Location: .github/skills/code-validation-standards/scripts/validate-before-push.sh
#
# Runs all validation checks before pushing to repository

set -e

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

failed=0
warnings=0

echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}Comprehensive Pre-Push Validation${NC}"
echo -e "${CYAN}Code Validation Standards Skill${NC}"
echo -e "${CYAN}========================================${NC}"
echo ""

# 1. Check format
echo -e "${YELLOW}[1/8] Format check...${NC}"
format_issues=$(gofmt -l . 2>/dev/null | grep -v ".git" | grep "\.go$" || true)
if [ -n "$format_issues" ]; then
    echo -e "${RED}✗ Format issues found:${NC}"
    echo "$format_issues" | sed 's/^/  - /'
    failed=1
else
    echo -e "${GREEN}✓ Format OK${NC}"
fi

# 2. Check vet
echo ""
echo -e "${YELLOW}[2/8] Running go vet...${NC}"
if ! go vet ./... 2>&1 | tee /tmp/vet-output.txt > /dev/null; then
    if grep -q "vet found\|error" /tmp/vet-output.txt; then
        echo -e "${RED}✗ Go vet found issues${NC}"
        head -10 /tmp/vet-output.txt | sed 's/^/  - /'
        failed=1
    else
        echo -e "${GREEN}✓ Vet OK${NC}"
    fi
else
    echo -e "${GREEN}✓ Vet OK${NC}"
fi

# 3. Check compile
echo ""
echo -e "${YELLOW}[3/8] Compile check...${NC}"
compile_errors=0
if ! go build -o /tmp/api-test ./cmd/api 2>&1 | tee /tmp/build-api.txt > /dev/null; then
    echo -e "${RED}✗ API compilation failed${NC}"
    head -10 /tmp/build-api.txt | sed 's/^/  - /'
    failed=1
    compile_errors=$((compile_errors + 1))
else
    echo -e "${GREEN}✓ API compiles${NC}"
fi

if ! go build -o /tmp/scheduler-test ./cmd/scheduler 2>&1 | tee /tmp/build-scheduler.txt > /dev/null; then
    echo -e "${RED}✗ Scheduler compilation failed${NC}"
    head -10 /tmp/build-scheduler.txt | sed 's/^/  - /'
    failed=1
    compile_errors=$((compile_errors + 1))
else
    echo -e "${GREEN}✓ Scheduler compiles${NC}"
fi

# 4. Lint check
echo ""
echo -e "${YELLOW}[4/8] Lint check...${NC}"
if command -v golangci-lint &> /dev/null; then
    if golangci-lint run --timeout=5m ./... > /tmp/lint-output.txt 2>&1; then
        echo -e "${GREEN}✓ Lint OK${NC}"
    else
        lint_count=$(grep -c "error:" /tmp/lint-output.txt || echo "0")
        echo -e "${YELLOW}⚠ Linting found ${lint_count} issues (non-blocking)${NC}"
        head -5 /tmp/lint-output.txt | sed 's/^/  - /'
        warnings=$((warnings + 1))
    fi
else
    echo -e "${YELLOW}⚠ golangci-lint not installed${NC}"
fi

# 5. Test with coverage
echo ""
echo -e "${YELLOW}[5/8] Test coverage check...${NC}"
if go test ./... -v -coverprofile=/tmp/coverage.out -timeout 10m > /tmp/test-output.txt 2>&1; then
    total=$(go tool cover -func=/tmp/coverage.out | grep total | grep -oP '[\d.]+(?=%)' | tail -1)
    echo -e "${GREEN}✓ Tests passed${NC}"
    
    if (( $(echo "$total >= 70" | bc -l 2>/dev/null || echo "0") )); then
        echo -e "${GREEN}  Coverage: ${total}% (target: 70%)${NC}"
    else
        echo -e "${YELLOW}⚠ Coverage is ${total}% (target: 70%)${NC}"
        warnings=$((warnings + 1))
    fi
else
    echo -e "${RED}✗ Tests failed${NC}"
    grep "FAIL\|error\|panic" /tmp/test-output.txt | head -5 | sed 's/^/  - /'
    failed=1
fi

# 6. Race detection
echo ""
echo -e "${YELLOW}[6/8] Race condition detection...${NC}"
if go test -race ./... -timeout 10m > /tmp/race-output.txt 2>&1; then
    echo -e "${GREEN}✓ No race conditions${NC}"
else
    if grep -q "RACE" /tmp/race-output.txt; then
        echo -e "${RED}✗ Race conditions detected${NC}"
        grep "RACE" /tmp/race-output.txt | head -5 | sed 's/^/  - /'
        failed=1
    else
        echo -e "${GREEN}✓ No race conditions${NC}"
    fi
fi

# 7. Security checks
echo ""
echo -e "${YELLOW}[7/8] Security checks...${NC}"

# Check for SQL injection risks
sql_risks=$(grep -r "Query.*fmt\.Sprintf\|Exec.*fmt\.Sprintf" --include="*.go" ./internal ./pkg ./cmd 2>/dev/null | grep -v "test.go" | grep -v "\.git" || true)
if [ -n "$sql_risks" ]; then
    echo -e "${RED}✗ Potential SQL injection risks found:${NC}"
    echo "$sql_risks" | head -3 | sed 's/^/  - /'
    failed=1
else
    echo -e "${GREEN}✓ No SQL injection patterns${NC}"
fi

# 8. Dependency check
echo ""
echo -e "${YELLOW}[8/8] Dependency check...${NC}"
if go mod verify > /tmp/mod-verify.txt 2>&1; then
    echo -e "${GREEN}✓ Dependencies OK${NC}"
else
    echo -e "${YELLOW}⚠ Module verification issues:${NC}"
    cat /tmp/mod-verify.txt | head -3 | sed 's/^/  - /'
    warnings=$((warnings + 1))
fi

# Summary
echo ""
echo -e "${CYAN}========================================${NC}"
echo -e "${CYAN}Validation Summary${NC}"
echo -e "${CYAN}========================================${NC}"

if [ $failed -eq 0 ] && [ $warnings -eq 0 ]; then
    echo -e "${GREEN}✓ All validation checks passed!${NC}"
    echo -e "${GREEN}Ready to push.${NC}"
    exit 0
elif [ $failed -eq 0 ]; then
    echo -e "${YELLOW}⚠ Some warnings found, but all critical checks passed${NC}"
    echo -e "${YELLOW}You can still push, but review warnings above${NC}"
    exit 0
else
    echo -e "${RED}✗ Some validation checks failed${NC}"
    echo -e "${RED}Please fix the issues above before pushing${NC}"
    exit 1
fi
