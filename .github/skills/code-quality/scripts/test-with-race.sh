#!/bin/bash

# Race condition detection script
# Used by: testing-best-practices skill
# Location: .github/skills/testing-best-practices/scripts/test-with-race.sh
#
# Runs tests with Go race detector enabled

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=== Race Condition Detection (testing-best-practices) ==="
echo ""
echo "Running tests with -race flag (this may take longer)..."
echo ""

if go test -race ./... -timeout 10m -v > /tmp/race-output.txt 2>&1; then
    echo -e "${GREEN}✓ No race conditions detected${NC}"
    tail -20 /tmp/race-output.txt
    exit 0
else
    if grep -q "RACE" /tmp/race-output.txt; then
        echo -e "${RED}✗ Race conditions detected!${NC}"
        echo ""
        grep "RACE\|test\|package" /tmp/race-output.txt | head -30
        exit 1
    else
        echo -e "${YELLOW}⚠ Tests failed (check output above)${NC}"
        tail -30 /tmp/race-output.txt
        exit 1
    fi
fi
