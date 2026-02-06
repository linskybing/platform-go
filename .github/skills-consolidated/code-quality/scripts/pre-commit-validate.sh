#!/bin/bash

# Pre-commit validation hook
# Used by: code-validation-standards skill
# Location: .github/skills/code-validation-standards/scripts/pre-commit-validate.sh
# 
# This runs before each commit to ensure code quality
# Install: ln -s ../../.github/skills/code-validation-standards/scripts/pre-commit-validate.sh .git/hooks/pre-commit

set -e

echo "=== Pre-commit Validation (code-validation-standards) ==="

# 1. Format check
echo "Checking code format..."
if gofmt -l . 2>/dev/null | grep -v ".git" | grep "\.go$" | head -10 > /tmp/fmt-issues.txt; then
    if [ -s /tmp/fmt-issues.txt ]; then
        echo "✗ Format issues found:"
        cat /tmp/fmt-issues.txt | sed 's/^/  - /'
        echo ""
        echo "Fix with: make fmt"
        exit 1
    fi
fi
echo "✓ Code format is correct"

# 2. Vet check
echo "Running go vet..."
if ! go vet ./... 2>&1 | tee /tmp/vet-output.txt | head -20; then
    if grep -q "vet found" /tmp/vet-output.txt || [ -s /tmp/vet-output.txt ]; then
        echo "✗ Go vet found issues"
        exit 1
    fi
fi
echo "✓ Go vet passed"

# 3. Compile check
echo "Checking if code compiles..."
if ! go build -o /tmp/validation-api ./cmd/api 2>&1; then
    echo "✗ API does not compile"
    exit 1
fi
if ! go build -o /tmp/validation-scheduler ./cmd/scheduler 2>&1; then
    echo "✗ Scheduler does not compile"
    exit 1
fi
echo "✓ Code compiles successfully"

# 4. Lint check (optional but recommended)
echo "Running linter..."
if command -v golangci-lint &> /dev/null; then
    if ! golangci-lint run --timeout=5m ./... 2>&1 | head -20; then
        echo "⚠ Linting found issues (continuing...)"
    else
        echo "✓ Linting passed"
    fi
else
    echo "⚠ golangci-lint not installed, skipping"
fi

echo ""
echo "✓ All pre-commit checks passed!"
