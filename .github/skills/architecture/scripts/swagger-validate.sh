#!/bin/bash

# Swagger validation script
# Used by: api-design-patterns skill
# Location: .github/skills/api-design-patterns/scripts/swagger-validate.sh
#
# Validates Swagger/OpenAPI documentation for correctness

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo "=== Swagger Validation (api-design-patterns) ==="
echo ""

# Check if swagger.json exists
if [ ! -f "docs/swagger.json" ]; then
    echo -e "${RED}✗ docs/swagger.json not found${NC}"
    echo "Generate first with: $0/../swagger-generate.sh"
    exit 1
fi

echo "Checking Swagger specification..."
echo ""

# 1. Valid JSON
echo -e "${YELLOW}[1/4] Validating JSON structure...${NC}"
if ! python3 -m json.tool docs/swagger.json > /dev/null 2>&1; then
    echo -e "${RED}✗ Invalid JSON in swagger.json${NC}"
    exit 1
fi
echo -e "${GREEN}✓ Valid JSON${NC}"

# 2. Check required Swagger fields
echo ""
echo -e "${YELLOW}[2/4] Checking required fields...${NC}"
required_fields=("swagger\|openapi" "info" "paths")
for field in "${required_fields[@]}"; do
    if ! grep -q "$field" docs/swagger.json; then
        echo -e "${RED}✗ Missing required field: $field${NC}"
        exit 1
    fi
done
echo -e "${GREEN}✓ All required fields present${NC}"

# 3. Check for handlers without documentation
echo ""
echo -e "${YELLOW}[3/4] Checking for undocumented handlers...${NC}"
undocumented=$(grep -r "^func.*Handler.*\(.*\*gin.Context" --include="*.go" ./internal/api ./cmd 2>/dev/null | grep -v "// " | wc -l)
if [ "$undocumented" -gt 0 ]; then
    echo -e "${YELLOW}⚠ Found ${undocumented} handlers without Swagger comments${NC}"
    echo "Add @Summary, @Description, @Tags, etc. to handler comments"
else
    echo -e "${GREEN}✓ All handlers appear documented${NC}"
fi

# 4. API info summary
echo ""
echo -e "${YELLOW}[4/4] API Summary...${NC}"
echo -e "${CYAN}API Title:${NC} $(grep '"title"' docs/swagger.json | head -1 | sed 's/.*"title": "\([^"]*\)".*/\1/')"
echo -e "${CYAN}API Version:${NC} $(grep '"version"' docs/swagger.json | head -1 | sed 's/.*"version": "\([^"]*\)".*/\1/')"

path_count=$(grep -c '"/' docs/swagger.json || echo "0")
echo -e "${CYAN}Endpoints:${NC} ${path_count}"

echo ""
echo -e "${GREEN}✓ Swagger validation passed!${NC}"
echo ""
echo "Next steps:"
echo "  1. Start API server: make dev"
echo "  2. View API docs: http://localhost:8080/swagger/index.html"
