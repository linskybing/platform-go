#!/bin/bash

# Test coverage validation script
# Used by: testing-best-practices skill
# Location: .github/skills/testing-best-practices/scripts/test-coverage.sh
#
# Validates test coverage meets minimum 70% threshold

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=== Test Coverage Validation (testing-best-practices) ==="
echo ""

# Run tests with coverage
echo "Running tests with coverage..."
if ! go test ./... -v -coverprofile=/tmp/coverage.out -timeout 10m > /tmp/test-output.txt 2>&1; then
    echo -e "${RED}✗ Tests failed${NC}"
    grep "FAIL\|error\|panic" /tmp/test-output.txt | head -10
    exit 1
fi

echo -e "${GREEN}✓ Tests passed${NC}"
echo ""

# Calculate coverage percentage
total=$(go tool cover -func=/tmp/coverage.out | grep total | grep -oP '[\d.]+(?=%)' | tail -1)

echo "Coverage Report:"
echo "================"
go tool cover -func=/tmp/coverage.out | tail -20
echo ""

echo "Total Coverage: ${total}%"
echo ""

# Check threshold
THRESHOLD=70
if (( $(echo "$total >= $THRESHOLD" | bc -l 2>/dev/null || echo "0") )); then
    echo -e "${GREEN}✓ Coverage meets threshold (${total}% >= ${THRESHOLD}%)${NC}"
    exit 0
else
    echo -e "${RED}✗ Coverage below threshold (${total}% < ${THRESHOLD}%)${NC}"
    echo ""
    echo "Generate HTML coverage report:"
    echo "  go tool cover -html=/tmp/coverage.out -o coverage.html"
    echo ""
    echo "View uncovered code and add tests"
    exit 1
fi
