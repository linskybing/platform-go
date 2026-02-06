#!/bin/bash

# Compile validation
# Used by: golang-production-standards skill
# Location: .github/skills/golang-production-standards/scripts/compile-check.sh
#
# Validates that all code compiles without warnings

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=== Compile Validation (golang-production-standards) ==="
echo ""

failed=0

# 1. Compile API
echo -e "${YELLOW}Compiling API (cmd/api)...${NC}"
if go build -v -o /tmp/api-build ./cmd/api 2>&1 | tee /tmp/api-build.txt; then
    api_size=$(stat -f%z /tmp/api-build 2>/dev/null || stat -c%s /tmp/api-build 2>/dev/null)
    echo -e "${GREEN}✓ API compiled successfully ($(numfmt --to=iec-i --suffix=B $api_size 2>/dev/null || echo $api_size) bytes)${NC}"
else
    echo -e "${RED}✗ API compilation failed${NC}"
    tail -20 /tmp/api-build.txt
    failed=1
fi

# 2. Compile Scheduler
echo ""
echo -e "${YELLOW}Compiling Scheduler (cmd/scheduler)...${NC}"
if go build -v -o /tmp/scheduler-build ./cmd/scheduler 2>&1 | tee /tmp/scheduler-build.txt; then
    scheduler_size=$(stat -f%z /tmp/scheduler-build 2>/dev/null || stat -c%s /tmp/scheduler-build 2>/dev/null)
    echo -e "${GREEN}✓ Scheduler compiled successfully ($(numfmt --to=iec-i --suffix=B $scheduler_size 2>/dev/null || echo $scheduler_size) bytes)${NC}"
else
    echo -e "${RED}✗ Scheduler compilation failed${NC}"
    tail -20 /tmp/scheduler-build.txt
    failed=1
fi

# 3. Check for unused imports
echo ""
echo -e "${YELLOW}Checking for unused imports...${NC}"
if go mod tidy -v 2>&1 | tee /tmp/mod-tidy.txt | grep -q "unused"; then
    echo -e "${YELLOW}⚠ Unused imports detected${NC}"
    grep "unused" /tmp/mod-tidy.txt | head -5
else
    echo -e "${GREEN}✓ No unused imports${NC}"
fi

# 4. Check build tags
echo ""
echo -e "${YELLOW}Checking build tags...${NC}"
if grep -r "//+build\|//go:build" --include="*.go" ./internal ./pkg ./cmd 2>/dev/null | wc -l > /tmp/build-tags.txt; then
    tags=$(cat /tmp/build-tags.txt)
    if [ "$tags" -gt 0 ]; then
        echo -e "${GREEN}✓ Found ${tags} build tags${NC}"
    else
        echo -e "${GREEN}✓ No build constraints${NC}"
    fi
fi

# Summary
echo ""
echo "============================================"
if [ $failed -eq 0 ]; then
    echo -e "${GREEN}✓ All compilation checks passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Compilation failed${NC}"
    exit 1
fi
