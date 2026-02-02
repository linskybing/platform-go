---
name: code-validation-standards
description: Code validation standards, pre-commit checks, validation scripts, and automated quality gates to ensure code correctness before generation
---

# Code Validation Standards

This skill ensures developers validate code correctness through systematic checks before writing or generating code. It provides scripts, tools, and standards to catch errors early.

## When to Use

Apply this skill when:
- Writing new code (BEFORE generating code)
- Reviewing code written by others
- Setting up pre-commit hooks
- Creating validation scripts for your workflow
- Checking code quality before committing
- Validating all generated code
- Running CI/CD pipeline checks
- Ensuring consistency across the project

## Core Principle: Check First, Generate Second

**NEVER** write or generate code without validation. Always:

1. **Plan** - Understand what code you'll write
2. **Check** - Validate against standards
3. **Generate** - Write the actual code
4. **Verify** - Run validation again
5. **Test** - Ensure it works

---

## Quick Start: Using Validation Scripts

This skill includes ready-to-use validation scripts in the `scripts/` directory:

```bash
# Pre-push comprehensive validation (RECOMMENDED)
bash .github/skills/code-validation-standards/scripts/validate-before-push.sh

# Pre-commit quick check
bash .github/skills/code-validation-standards/scripts/pre-commit-validate.sh

# Setup pre-commit hook (automatic on each git commit)
ln -s ../../.github/skills/code-validation-standards/scripts/pre-commit-validate.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit
```

---

## 1. Pre-Code Validation Checklist

Before writing ANY new code, verify these points:

### Type Safety & Structure
- [ ] Is the function signature correct (parameters, return types)?
- [ ] Are all error types properly defined and handled?
- [ ] Are there any type mismatches or nil pointer risks?
- [ ] Do request/response DTOs match API specifications?

### Error Handling
- [ ] Does every error-returning function have error handling?
- [ ] Are errors wrapped with context using `fmt.Errorf`?
- [ ] Is there a plan for error recovery?
- [ ] Are panic conditions documented?

### Database Operations
- [ ] Are SQL queries parameterized (no string concatenation)?
- [ ] Are database transactions properly scoped?
- [ ] Are migrations versioned correctly?
- [ ] Is connection pooling configured?

### Concurrency & Performance
- [ ] Are goroutines properly managed (WaitGroup, context)?
- [ ] Is there a bounded channel size to prevent deadlocks?
- [ ] Are shared resources protected with mutexes?
- [ ] Are expensive operations cached appropriately?

### Security
- [ ] Is user input validated at the API boundary?
- [ ] Are passwords hashed before storage?
- [ ] Are SQL injection vulnerabilities prevented?
- [ ] Is sensitive data (passwords, tokens) never logged?

### Testing
- [ ] Can this code be unit tested?
- [ ] Are edge cases and error conditions covered?
- [ ] Is the code testable (no hard dependencies)?
- [ ] Will this require integration tests?

---

## 2. Automated Validation Scripts

### 2.1 Pre-Commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
set -e

echo "=== Pre-commit Validation ==="

# 1. Format check
echo "Checking code format..."
if ! gofmt -l . | grep -q .; then
    echo "✓ Code format is correct"
else
    echo "✗ Format issues found. Run: make fmt"
    exit 1
fi

# 2. Vet check
echo "Running go vet..."
if ! go vet ./...; then
    echo "✗ Go vet found issues"
    exit 1
fi

# 3. Compile check
echo "Checking if code compiles..."
if ! go build -o /tmp/validation-test ./cmd/api; then
    echo "✗ Code does not compile"
    exit 1
fi

# 4. Lint check (if available)
echo "Running linter..."
if command -v golangci-lint &> /dev/null; then
    if ! golangci-lint run --timeout=5m ./...; then
        echo "✗ Linting failed"
        exit 1
    fi
else
    echo "⚠ golangci-lint not installed"
fi

# 5. Test check (quick unit tests)
echo "Running quick tests..."
if ! go test -short -timeout 30s ./pkg/... ./internal/....; then
    echo "✗ Quick tests failed"
    exit 1
fi

echo ""
echo "✓ All pre-commit checks passed!"
```

Install with:
```bash
chmod +x .git/hooks/pre-commit
```

### 2.2 Pre-Push Validation Script

Create `scripts/validate-before-push.sh`:

```bash
#!/bin/bash
set -e

echo "=== Comprehensive Pre-Push Validation ==="

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

failed=0

# 1. Check format
echo -e "${YELLOW}[1/7] Format check...${NC}"
if gofmt -l . | grep -q .; then
    echo -e "${RED}✗ Format issues found${NC}"
    gofmt -l . | sed 's/^/  - /'
    failed=1
else
    echo -e "${GREEN}✓ Format OK${NC}"
fi

# 2. Check vet
echo -e "${YELLOW}[2/7] Running go vet...${NC}"
if ! go vet ./... 2>&1 | tee /tmp/vet-output.txt; then
    echo -e "${RED}✗ Go vet found issues${NC}"
    failed=1
else
    echo -e "${GREEN}✓ Vet OK${NC}"
fi

# 3. Check compile
echo -e "${YELLOW}[3/7] Compile check...${NC}"
if ! go build -o /tmp/api-test ./cmd/api || ! go build -o /tmp/scheduler-test ./cmd/scheduler; then
    echo -e "${RED}✗ Compilation failed${NC}"
    failed=1
else
    echo -e "${GREEN}✓ Compilation OK${NC}"
fi

# 4. Lint check
echo -e "${YELLOW}[4/7] Lint check...${NC}"
if command -v golangci-lint &> /dev/null; then
    if ! golangci-lint run --timeout=5m ./... 2>&1 | head -20; then
        echo -e "${YELLOW}⚠ Linting issues found (non-blocking)${NC}"
    else
        echo -e "${GREEN}✓ Lint OK${NC}"
    fi
else
    echo -e "${YELLOW}⚠ golangci-lint not installed${NC}"
fi

# 5. Test coverage
echo -e "${YELLOW}[5/7] Test coverage check...${NC}"
if go test ./... -v -coverprofile=/tmp/coverage.out -timeout 5m; then
    total=$(go tool cover -func=/tmp/coverage.out | grep total | grep -oP '\d+\.\d+' | tail -1)
    if (( $(echo "$total >= 70" | bc -l) )); then
        echo -e "${GREEN}✓ Coverage OK (${total}%)${NC}"
    else
        echo -e "${YELLOW}⚠ Coverage is ${total}% (target: 70%)${NC}"
    fi
else
    echo -e "${RED}✗ Tests failed${NC}"
    failed=1
fi

# 6. Race detection
echo -e "${YELLOW}[6/7] Race condition detection...${NC}"
if go test -race ./... -timeout 5m 2>&1 | grep -q "RACE"; then
    echo -e "${RED}✗ Race conditions detected${NC}"
    failed=1
else
    echo -e "${GREEN}✓ No race conditions${NC}"
fi

# 7. SQL injection check
echo -e "${YELLOW}[7/7] SQL injection check...${NC}"
if grep -r "Query.*fmt\.Sprintf\|Exec.*fmt\.Sprintf" --include="*.go" . 2>/dev/null | grep -v test | grep -v "\.git"; then
    echo -e "${RED}✗ Potential SQL injection found${NC}"
    failed=1
else
    echo -e "${GREEN}✓ SQL injection check OK${NC}"
fi

echo ""
if [ $failed -eq 0 ]; then
    echo -e "${GREEN}✓ All validation checks passed! Ready to push.${NC}"
    exit 0
else
    echo -e "${RED}✗ Some validation checks failed. Please fix before pushing.${NC}"
    exit 1
fi
```

Install with:
```bash
chmod +x scripts/validate-before-push.sh
```

### 2.3 Type Checking Script

Create `scripts/type-check.sh`:

```bash
#!/bin/bash
set -e

echo "=== Type Safety Validation ==="

# Check for unsafe patterns
echo "Checking for type safety issues..."

# 1. Uninitialized slices
echo "  - Checking for uninitialized slices..."
if grep -r "var.*\[\]" --include="*.go" ./internal ./pkg | grep -v "var.*\[\].*=" | grep -v test; then
    echo "    ⚠ Uninitialized slice found (may be intended)"
fi

# 2. Type assertions without checking
echo "  - Checking type assertions..."
if grep -r "\.\(.*)\s*$" --include="*.go" ./internal ./pkg | grep -v "ok" | head -5; then
    echo "    ⚠ Unchecked type assertion found"
fi

# 3. Missing error handling
echo "  - Checking for ignored errors..."
if grep -r "_ = .*Error()" --include="*.go" ./internal ./pkg | head -5; then
    echo "    ⚠ Intentionally ignored errors found"
fi

# 4. Pointer dereferences
echo "  - Checking for nil dereference risks..."
if grep -r "\..*\s*$" --include="*.go" ./internal ./pkg | grep -v "// nil" | head -5; then
    echo "    (Use linter for detailed checks)"
fi

echo "✓ Type safety check complete"
```

---

## 3. Per-Skill Validation Standards

### 3.1 API Design Validation

Before creating an endpoint:

```bash
# Check Swagger compliance
- [ ] Handler has @Summary comment (≤120 chars)
- [ ] Handler has @Description comment
- [ ] Request DTO has field examples
- [ ] Response models defined
- [ ] All error cases documented (@Failure)
- [ ] Security scheme specified (@Security)
- [ ] HTTP status codes match spec

# Generate and validate
swag init -g cmd/api/main.go
swagger validate ./docs/swagger.json
```

### 3.2 Database Validation

Before writing database code:

```bash
# Check safety
- [ ] No string concatenation in queries
- [ ] Transactions properly scoped
- [ ] Migrations versioned (001_, 002_, etc.)
- [ ] Indexes defined for frequently queried columns
- [ ] Foreign key constraints documented

# Validate GORM usage
grep -r "Raw(" internal/ | grep -v "// approved"
```

### 3.3 Testing Validation

Before committing tests:

```bash
# Validate test structure
- [ ] Uses AAA pattern (Arrange, Act, Assert)
- [ ] Uses table-driven tests for multiple cases
- [ ] Error cases explicitly tested
- [ ] Edge conditions covered
- [ ] Concurrent access tested with -race flag

# Run validation
go test ./... -v -race -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -1
```

### 3.4 Kubernetes Code Validation

Before committing K8s code:

```bash
# Check safety
- [ ] Uses context with timeout (30s for K8s)
- [ ] Handles nil clients gracefully
- [ ] Labels/selectors properly formatted
- [ ] Resource limits specified
- [ ] Retries use exponential backoff

# Validate manifests (if applicable)
kubectl apply -f k8s/ --dry-run=client
```

---

## 4. Code Generation Validation Checklist

When using any code generation tool:

### Before Generation
```bash
- [ ] Tool is installed and up-to-date
- [ ] Configuration file reviewed
- [ ] Output directory is clean
- [ ] Backup of existing code taken
```

### During Generation
```bash
- [ ] Generation command runs without warnings
- [ ] All output files created successfully
- [ ] Generated code has expected structure
```

### After Generation
```bash
- [ ] Code compiles: go build ./...
- [ ] No type errors: go vet ./...
- [ ] Format is correct: gofmt -l .
- [ ] Lint passes: golangci-lint run
- [ ] Tests pass: go test ./...
- [ ] Generated code reviewed manually
```

Example:
```bash
#!/bin/bash
set -e

# Backup
cp -r internal internal.backup

# Generate
swag init -g cmd/api/main.go

# Validate
go build ./...
go vet ./...
gofmt -l . || true
golangci-lint run ./...
go test ./...

echo "✓ Generated code validated successfully"
```

---

## 5. Related Skills Cross-Check

Ensure consistency with:

- **golang-production-standards**: Type safety, error handling, performance
- **testing-best-practices**: Test structure, coverage, mocking
- **database-best-practices**: Query safety, transactions, migrations
- **api-design-patterns**: Request/response validation, Swagger compliance
- **kubernetes-integration**: Client safety, resource handling
- **security-best-practices**: Input validation, injection prevention
- **error-handling-guide**: Error wrapping, logging standards

---

## 6. Validation Commands Quick Reference

```bash
# Format check
gofmt -l .

# Static analysis
go vet ./...

# Compile check
go build ./cmd/api ./cmd/scheduler

# Lint
golangci-lint run ./...

# Test with race detection
go test -race ./...

# Coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Full validation
./scripts/validate-before-push.sh
```

---

## 7. IDE Integration

### VS Code Settings (.vscode/settings.json)

```json
{
  "go.lintOnSave": "package",
  "go.lintTool": "golangci-lint",
  "[go]": {
    "editor.formatOnSave": true,
    "editor.defaultFormatter": "golang.go",
    "editor.codeActionsOnSave": {
      "source.fixAll": true,
      "source.organizeImports": true
    }
  }
}
```

### GoLand/IntelliJ Settings

- Enable `go vet` inspection
- Enable `go fmt` code style
- Enable `golangci-lint` inspection
- Set minimum coverage threshold to 70%

---

## 8. Common Validation Issues & Fixes

### Issue: Code doesn't compile
**Fix**: 
```bash
go build ./... 2>&1 | head -20  # See first 20 errors
```

### Issue: Test coverage below 70%
**Fix**:
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out  # View uncovered code
```

### Issue: SQL injection detected
**Fix**:
```bash
# Use parameterized queries
// BAD
db.Query(fmt.Sprintf("SELECT * FROM users WHERE id = %d", id))

// GOOD
db.Query("SELECT * FROM users WHERE id = ?", id)
```

### Issue: Race condition detected
**Fix**:
```bash
go test -race ./...  # Run with race detector
# Fix: Use mutexes, channels, or atomic operations
```

### Issue: Linting failures
**Fix**:
```bash
golangci-lint run --fix ./...  # Auto-fix some issues
```

---

## Validation Standards Checklist

- [ ] All code passes `gofmt` format check
- [ ] All code passes `go vet` analysis
- [ ] Code compiles without warnings
- [ ] `golangci-lint` runs with no errors
- [ ] All tests pass: `go test -race ./...`
- [ ] Test coverage >= 70%
- [ ] No SQL injection vulnerabilities
- [ ] No race conditions detected
- [ ] Error handling complete
- [ ] Database migrations versioned
- [ ] API endpoints have Swagger docs
- [ ] Security checks passed

---

**Last Updated**: 2026-02-02
**Related Skills**: golang-production-standards, testing-best-practices, database-best-practices, api-design-patterns

