---
name: code-validation-standards
description: Code validation standards, pre-commit checks, validation scripts, and automated quality gates to ensure code correctness before generation
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
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


## 5. Related Skills Cross-Check

Ensure consistency with:

- **golang-production-standards**: Type safety, error handling, performance
- **testing-best-practices**: Test structure, coverage, mocking
- **database-best-practices**: Query safety, transactions, migrations
- **api-design-patterns**: Request/response validation, Swagger compliance
- **kubernetes-integration**: Client safety, resource handling
- **security-best-practices**: Input validation, injection prevention
- **error-handling-guide**: Error wrapping, logging standards


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
