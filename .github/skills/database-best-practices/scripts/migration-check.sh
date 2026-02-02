#!/bin/bash

# Migration validation script
# Used by: database-best-practices skill
# Location: .github/skills/database-best-practices/scripts/migration-check.sh
#
# Validates database migration files and safety

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=== Migration Safety Check (database-best-practices) ==="
echo ""

# Check migration directory
MIGRATION_DIR="infra/db/migrations"
if [ ! -d "$MIGRATION_DIR" ]; then
    echo -e "${YELLOW}⚠ Migration directory not found: $MIGRATION_DIR${NC}"
    exit 0
fi

echo "Checking migrations in: $MIGRATION_DIR"
echo ""

# 1. Check versioning format
echo -e "${YELLOW}[1/4] Checking migration version format...${NC}"
bad_format=0
for file in "$MIGRATION_DIR"/*.sql; do
    if [ -f "$file" ]; then
        basename_file=$(basename "$file")
        if ! [[ "$basename_file" =~ ^[0-9]{3,}_.*\.sql$ ]]; then
            echo -e "${RED}✗ Invalid format: $basename_file${NC}"
            echo "  Expected format: 001_description.sql, 002_next.sql, etc."
            bad_format=$((bad_format + 1))
        fi
    fi
done

if [ $bad_format -eq 0 ]; then
    echo -e "${GREEN}✓ All migrations properly versioned${NC}"
else
    exit 1
fi

# 2. Check for unsafe patterns
echo ""
echo -e "${YELLOW}[2/4] Checking for unsafe SQL patterns...${NC}"
unsafe_patterns=0

for file in "$MIGRATION_DIR"/*.sql; do
    if [ -f "$file" ]; then
        # Check for DROP without IF EXISTS
        if grep -i "DROP TABLE" "$file" | grep -v "IF EXISTS" > /dev/null; then
            echo -e "${RED}✗ Unsafe DROP in $(basename $file):${NC}"
            grep -i "DROP TABLE" "$file"
            unsafe_patterns=$((unsafe_patterns + 1))
        fi
        
        # Check for TRUNCATE
        if grep -i "TRUNCATE" "$file" > /dev/null; then
            echo -e "${YELLOW}⚠ TRUNCATE in $(basename $file):${NC}"
            grep -i "TRUNCATE" "$file" | sed 's/^/  - /'
        fi
    fi
done

if [ $unsafe_patterns -eq 0 ]; then
    echo -e "${GREEN}✓ No unsafe patterns detected${NC}"
else
    exit 1
fi

# 3. Check migration content
echo ""
echo -e "${YELLOW}[3/4] Analyzing migrations...${NC}"
total_migrations=$(ls "$MIGRATION_DIR"/*.sql 2>/dev/null | wc -l)
echo "Total migrations: $total_migrations"

# List all migrations
echo ""
echo "Migrations:"
ls -1 "$MIGRATION_DIR"/*.sql 2>/dev/null | nl -v 1 | sed 's/^/  /'

# 4. Check for reverse/rollback migrations
echo ""
echo -e "${YELLOW}[4/4] Checking for rollback support...${NC}"
down_migrations=$(ls "$MIGRATION_DIR"/*_down.sql 2>/dev/null | wc -l)
if [ $down_migrations -eq 0 ]; then
    echo -e "${YELLOW}⚠ No down migrations found (rollback support limited)${NC}"
else
    echo -e "${GREEN}✓ Down migrations found: $down_migrations${NC}"
fi

echo ""
echo -e "${GREEN}✓ Migration check complete!${NC}"
echo ""
echo "Best practices:"
echo "  1. Always use IF EXISTS in destructive operations"
echo "  2. Test migrations in development first"
echo "  3. Maintain down migrations for rollback"
echo "  4. Keep migrations small and focused"
echo "  5. Use transactions where possible"
