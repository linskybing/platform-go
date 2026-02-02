---
name: golang-production-standards
description: Enforce production-grade Golang coding standards, best practices, error handling, and performance optimization for the platform-go project
---

# Golang Production Standards

This skill ensures all code written for platform-go follows production-grade Golang standards, focusing on maintainability, performance, and reliability.

## When to Use

Apply this skill when:
- Writing new Go code for platform-go project
- Creating functions, packages, or services
- Handling errors and logging
- Implementing concurrent operations
- Optimizing performance-critical code
- Writing code that will interact with external systems (databases, Kubernetes, APIs)
- Performing code review on Go files

## Quick Start: Using Production Standards Scripts

This skill includes ready-to-use validation scripts:

```bash
# Comprehensive production standard checks
bash .github/skills/golang-production-standards/scripts/lint-check.sh

# Verify code compiles correctly
bash .github/skills/golang-production-standards/scripts/compile-check.sh
```

## Core Principles

### 1. Code Organization
- Clean architecture: domain → application → API layers
- One responsibility per package
- internal/ for app code, pkg/ for reusable libraries
- Repository pattern for data access

### 2. Error Handling (Critical)
```go
// Always wrap errors with context
if err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

// Use custom error types for domain errors
type ValidationError struct {
    Field string
    Msg   string
}

// Never ignore errors
_ = someFunc() // Instead of: someFunc()
```

### 3. Context Propagation
- Always pass `context.Context` as first parameter
- Use context for cancellation, timeouts, request-scoped values
- Set timeouts: 10s for DB, 30s for K8s operations

```go
func (s *Service) CreateResource(ctx context.Context, req *Request) error {
    ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
    defer cancel()
    return s.repo.Create(ctx, req)
}
```

### 4. Concurrency & Performance

#### Goroutine Management
```go
// Use sync.WaitGroup for coordination
var wg sync.WaitGroup
for _, item := range items {
    wg.Add(1)
    go func(i Item) {
        defer wg.Done()
        process(i)
    }(item)
}
wg.Wait()

// Limit concurrency with semaphore
sem := make(chan struct{}, maxConcurrency)
for _, task := range tasks {
    sem <- struct{}{}
    go func(t Task) {
        defer func() { <-sem }()
        execute(t)
    }(task)
}
```

#### Resource Pooling
- Use connection pools for DB/HTTP clients
- Reuse buffers with sync.Pool for frequent allocations
- Cache expensive computations

#### Optimization
```go
// Preallocate slices
results := make([]Result, 0, expectedSize)

// Use strings.Builder for concatenation
var b strings.Builder
for _, s := range strs {
    b.WriteString(s)
}
```

### 5. Testing Requirements

#### Unit Tests
- Minimum 70% code coverage
- Table-driven tests for multiple cases
- Mock external dependencies

```go
func TestCreateUser(t *testing.T) {
    tests := []struct {
        name    string
        input   *User
        wantErr bool
    }{
        {"valid user", &User{Name: "alice"}, false},
        {"empty name", &User{Name: ""}, true},
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
- Test external dependencies in test/integration/
- Use t.Skip() when unavailable
- Clean up resources in teardown

### 6. Database Operations
```go
// Use transactions for multiple operations
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

if err := tx.Create(&profile).Error; err != nil {
    tx.Rollback()
    return fmt.Errorf("create profile failed: %w", err)
}

return tx.Commit().Error
```

- Add indexes for frequently queried fields
- Use pagination for large result sets
- GORM prevents SQL injection automatically

### 7. Kubernetes Client Usage
```go
// Check if Clientset exists (test environments)
if k8s.Clientset == nil {
    return errors.New("k8s client not initialized")
}

// Use label selectors for queries
listOpts := metav1.ListOptions{
    LabelSelector: "app=platform,tier=backend",
}

// Set resource limits in pod specs
resources := corev1.ResourceRequirements{
    Requests: corev1.ResourceList{
        corev1.ResourceCPU:    resource.MustParse("100m"),
        corev1.ResourceMemory: resource.MustParse("128Mi"),
    },
    Limits: corev1.ResourceList{
        corev1.ResourceCPU:    resource.MustParse("500m"),
        corev1.ResourceMemory: resource.MustParse("512Mi"),
    },
}
```

### 8. Security Best Practices
- Never log sensitive data (passwords, tokens, secrets)
- Validate all user input at API boundary
- Use bcrypt for password hashing (min cost 10)
- Sanitize file paths to prevent traversal
- Use HTTPS for external communications

```go
// Validate input
if !isValidUsername(input.Username) {
    return errors.New("invalid username")
}

// Sanitize paths
safePath := filepath.Clean(userInput)
if !strings.HasPrefix(safePath, allowedDir) {
    return errors.New("access denied")
}
```

### 9. Code Quality Checklist

Before committing:
- [ ] All errors wrapped and handled
- [ ] Context propagated through calls
- [ ] No goroutine leaks
- [ ] Resources closed (defer)
- [ ] Tests cover happy path and errors
- [ ] No hardcoded values
- [ ] Code formatted with gofmt
- [ ] No linter warnings
- [ ] Public APIs have godoc comments

### 10. Common Anti-Patterns

```go
// Don't use panic in library code
panic("error") // Use error returns

// Check context cancellation
for {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        doWork()
    }
}

// Copy loop variables before goroutine
for _, item := range items {
    go process(item) // Correct (not &item)
}

// Synchronize shared state
// Use sync.Mutex or channels
```

## Project-Specific Guidelines

### Manager Pattern
- Use managers for domain logic
- Managers must be stateless or properly synchronized
- Inject dependencies through constructors

### DTO Usage
- Separate domain models from DTOs
- DTOs for API request/response
- Validate at API layer before passing

### Repository Pattern
- All database access through repository layer
- Return domain models, not GORM models
- Use interfaces for contracts

## Performance Targets

- API response time: < 200ms (p95)
- Database query time: < 100ms
- Kubernetes API calls: < 500ms
- Memory usage: < 512MB per pod under normal load
- Goroutine count: Monitor and prevent leaks

## Tools & Commands

```bash
# Format code
go fmt ./...

# Run tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Check for race conditions
go test -race ./...

# Build optimized binary
go build -ldflags="-s -w" -o bin/api cmd/api/main.go

# Profile CPU usage
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof
```

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
