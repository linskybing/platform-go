#!/bin/bash

# Swagger documentation generation
# Used by: api-design-patterns skill
# Location: .github/skills/api-design-patterns/scripts/swagger-generate.sh
#
# Generates Swagger/OpenAPI documentation from Go code comments

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=== Swagger Generation (api-design-patterns) ==="
echo ""

# Check if swag is installed
if ! command -v swag &> /dev/null; then
    echo -e "${YELLOW}swag not found. Installing...${NC}"
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Generate swagger docs
echo "Generating Swagger documentation from cmd/api/main.go..."
if swag init -g cmd/api/main.go 2>&1 | tee /tmp/swagger-gen.txt; then
    if [ -f "docs/swagger.json" ]; then
        echo -e "${GREEN}✓ Swagger documentation generated${NC}"
        echo ""
        echo "Generated files:"
        ls -lh docs/ | grep -E "swagger|docs" | awk '{print "  " $9 " (" $5 ")"}'
        echo ""
        echo "Next steps:"
        echo "  1. Verify: ./$(dirname $0)/swagger-validate.sh"
        echo "  2. View API docs at: http://localhost:8080/swagger/index.html"
    else
        echo -e "${RED}✗ Swagger generation failed - swagger.json not created${NC}"
        cat /tmp/swagger-gen.txt
        exit 1
    fi
else
    echo -e "${RED}✗ Swagger generation failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Done!${NC}"
