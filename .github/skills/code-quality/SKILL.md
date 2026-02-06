---
name: code-quality
description: Golang production-grade code standards, testing, validation, and error handling for platform-go
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
  consolidated_from:
    - golang-production-standards
    - testing-best-practices
    - error-handling-guide
    - code-validation-standards
    - file-structure-guidelines
---

# Code Quality Excellence

Comprehensive guidelines for production-grade code standards, comprehensive testing, robust error handling, and clean code organization.

## Core Principles

### 1. Production Standards
- Clean architecture: domain → application → API layers
- One responsibility per package
- Context propagation for all operations
- Proper error wrapping with context
- No Chinese characters or emojis in code

**Error Handling Pattern:**
```go
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}
```

### 2. Code Organization (200-Line File Limit)
- Maximum 200 lines per file for maintainability
- Separate concerns across files
- Clear file naming conventions
- Domain-driven design structure

**Directory Structure Example:**
```
domain/user/
├── model.go          (< 100 lines)
├── dto.go            (< 150 lines)
├── errors.go         (< 50 lines)
├── repository.go     (interface < 50 lines)
├── service.go        (interface < 50 lines)
└── README.md
```

### 3. Comprehensive Testing (70%+ Coverage)

#### Unit Tests
- Table-driven tests for multiple cases
- Mock external dependencies
- Clear test names with given-when-then pattern
- Minimum 70% code coverage

```go
func TestCreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   *CreateUserRequest
        wantErr bool
    }{
        {"valid user", &CreateUserRequest{Name: "alice"}, false},
        {"empty name", &CreateUserRequest{Name: ""}, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := CreateUser(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error = %v, want %v", err, tt.wantErr)
            }
        })
    }
}
```

#### Integration Tests
- Test external dependencies (database, K8s, APIs)
- Use t.Skip() when unavailable
- Proper resource cleanup in teardown

### 4. Concurrency & Performance
- Use sync.WaitGroup for goroutine coordination
- Limit concurrency with semaphores
- Preallocate slices with expected size
- Connection pooling for databases
- Resource pooling for frequently allocated objects

```go
// Limit concurrent operations
sem := make(chan struct{}, maxConcurrency)
for _, task := range tasks {
    sem <- struct{}{}
    go func(t Task) {
        defer func() { <-sem }()
        execute(t)
    }(task)
}
```

### 5. Database Operations
- Use transactions for multiple operations
- Add indexes for frequently queried fields
- Use pagination for large result sets
- GORM prevents SQL injection automatically

```go
tx := db.Begin()
defer func() {
    if r := recover(); r != nil {
        tx.Rollback()
    }
}()

if err := tx.Create(&user).Error; err != nil {
    tx.Rollback()
    return fmt.Errorf("create user failed: %w", err)
}
return tx.Commit().Error
```

## Validation Checklist

### Code Review
- [ ] No Chinese characters or emojis
- [ ] All imports used
- [ ] Error wrapping with context
- [ ] Context as first parameter
- [ ] Clear function names
- [ ] File < 200 lines
- [ ] Proper test coverage

### Testing
- [ ] Unit tests written (70%+ coverage)
- [ ] Integration tests for external calls
- [ ] Table-driven tests for multiple cases
- [ ] Error cases tested
- [ ] Mock dependencies properly

### Error Handling
- [ ] All errors wrapped with context
- [ ] Custom error types for domain errors
- [ ] Proper error logging
- [ ] No panic in production code
- [ ] Graceful degradation

### Performance
- [ ] Connection pooling used
- [ ] Queries optimized (indexes, pagination)
- [ ] Goroutine limits enforced
- [ ] Memory allocations minimized
- [ ] No blocking operations in loops

## Tools & Scripts

### Validation Scripts
```bash
# Run comprehensive code quality checks
bash .github/skills-consolidated/code-quality/scripts/validate-code.sh

# Run all tests with coverage
bash .github/skills-consolidated/code-quality/scripts/run-tests.sh

# Check code formatting
bash .github/skills-consolidated/code-quality/scripts/format-check.sh
```

## References
- Effective Go: https://golang.org/doc/effective_go
- Go Code Review Comments: https://github.com/golang/go/wiki/CodeReviewComments
- GORM Best Practices: https://gorm.io/docs
