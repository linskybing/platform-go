#!/bin/bash

# Generate HTML coverage report
# Used by: testing-best-practices skill
# Location: .github/skills/testing-best-practices/scripts/coverage-html.sh
#
# Generates HTML coverage report and opens it in browser

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=== Coverage HTML Report (testing-best-practices) ==="
echo ""

# Run tests with coverage
echo "Running tests with coverage..."
if ! go test ./... -coverprofile=/tmp/coverage.out -timeout 10m > /tmp/test-coverage.txt 2>&1; then
    echo "⚠ Some tests may have failed, but generating report anyway..."
fi

# Generate HTML
echo "Generating HTML report..."
go tool cover -html=/tmp/coverage.out -o coverage.html

echo -e "${GREEN}✓ Coverage report generated: coverage.html${NC}"
echo ""

# Show summary
echo "Coverage Summary:"
echo "================"
go tool cover -func=/tmp/coverage.out | tail -5
echo ""

# Try to open in browser
if command -v xdg-open &> /dev/null; then
    xdg-open coverage.html
elif command -v open &> /dev/null; then
    open coverage.html
else
    echo -e "${YELLOW}Open coverage.html in your browser to view the report${NC}"
fi

echo -e "${GREEN}✓ Done!${NC}"
