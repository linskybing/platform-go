---
name: testing-best-practices
description: Comprehensive testing guidelines for unit tests, integration tests, table-driven tests, and test coverage in platform-go
---

# Testing Best Practices

This skill ensures comprehensive test coverage and follows Golang testing best practices.

## When to Use

Apply this skill when:
- Writing unit tests for functions and methods
- Creating integration tests that require external services (database, Kubernetes)
- Implementing table-driven tests for multiple scenarios
- Mocking external dependencies
- Testing error cases and edge conditions
- Writing concurrent code that needs race condition testing
- Benchmarking performance-critical code
- Setting up test fixtures and helper functions

## Quick Start: Using Test Scripts

This skill includes ready-to-use test validation scripts:

```bash
# Generate HTML coverage report
bash .github/skills/testing-best-practices/scripts/coverage-html.sh

# Test with race condition detection
bash .github/skills/testing-best-practices/scripts/test-with-race.sh

# Quick coverage validation (70% threshold)
bash .github/skills/testing-best-practices/scripts/test-coverage.sh
```

## Testing Principles

### 1. Test Structure (AAA Pattern)

```go
func TestCreateUser(t *testing.T) {
    // Arrange - Setup test data and dependencies
    repo := NewMockUserRepository()
    service := NewUserService(repo)
    input := &CreateUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
    }
    
    // Act - Execute the function under test
    user, err := service.CreateUser(context.Background(), input)
    
    // Assert - Verify the results
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Username != input.Username {
        t.Errorf("expected username %s, got %s", input.Username, user.Username)
    }
}
```

### 2. Table-Driven Tests (Preferred)

```go
func TestValidateUsername(t *testing.T) {
    tests := []struct {
        name     string
        username string
        wantErr  bool
        errMsg   string
    }{
        {
            name:     "valid username",
            username: "alice",
            wantErr:  false,
        },
        {
            name:     "empty username",
            username: "",
            wantErr:  true,
            errMsg:   "username cannot be empty",
        },
        {
            name:     "username too long",
            username: strings.Repeat("a", 65),
            wantErr:  true,
            errMsg:   "username too long",
        },
        {
            name:     "invalid characters",
            username: "user@123",
            wantErr:  true,
            errMsg:   "invalid characters",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateUsername(tt.username)
            
            if tt.wantErr {
                if err == nil {
                    t.Errorf("expected error but got none")
                }
                if err != nil && !strings.Contains(err.Error(), tt.errMsg) {
                    t.Errorf("expected error containing %q, got %v", tt.errMsg, err)
                }
            } else {
                if err != nil {
                    t.Errorf("unexpected error: %v", err)
                }
            }
        })
    }
}
```

### 3. Mocking External Dependencies

#### Interface-Based Mocking
```go
// Define interface
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    FindByID(ctx context.Context, id uint) (*User, error)
}

// Mock implementation
type MockUserRepository struct {
    CreateFunc    func(ctx context.Context, user *User) error
    FindByIDFunc  func(ctx context.Context, id uint) (*User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *User) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, user)
    }
    return nil
}

func (m *MockUserRepository) FindByID(ctx context.Context, id uint) (*User, error) {
    if m.FindByIDFunc != nil {
        return m.FindByIDFunc(ctx, id)
    }
    return &User{ID: id}, nil
}

// Usage in tests
func TestUserService(t *testing.T) {
    mockRepo := &MockUserRepository{
        CreateFunc: func(ctx context.Context, user *User) error {
            user.ID = 123
            return nil
        },
    }
    
    service := NewUserService(mockRepo)
    // ... test service
}
```

### 4. Testing with Context

```go
func TestOperationTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    
    // Simulate slow operation
    err := SlowOperation(ctx)
    
    if err == nil {
        t.Error("expected timeout error")
    }
    if !errors.Is(err, context.DeadlineExceeded) {
        t.Errorf("expected context.DeadlineExceeded, got %v", err)
    }
}

func TestContextCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    
    go func() {
        time.Sleep(50 * time.Millisecond)
        cancel()
    }()
    
    err := LongRunningOperation(ctx)
    
    if !errors.Is(err, context.Canceled) {
        t.Errorf("expected context.Canceled, got %v", err)
    }
}
```

### 5. Integration Tests

```go
// +build integration

package integration_test

import (
    "testing"
    "context"
    "github.com/linskybing/platform-go/test/integration/testutil"
)

func TestUserCRUD(t *testing.T) {
    // Skip if database not available
    db := testutil.SetupTestDB(t)
    if db == nil {
        t.Skip("database not available")
    }
    defer testutil.TeardownTestDB(t, db)
    
    // Run integration test
    repo := repository.NewUserRepository(db)
    
    user := &User{Username: "testuser"}
    err := repo.Create(context.Background(), user)
    if err != nil {
        t.Fatalf("failed to create user: %v", err)
    }
    
    retrieved, err := repo.FindByID(context.Background(), user.ID)
    if err != nil {
        t.Fatalf("failed to retrieve user: %v", err)
    }
    
    if retrieved.Username != user.Username {
        t.Errorf("username mismatch: expected %s, got %s", user.Username, retrieved.Username)
    }
    
    // Cleanup
    err = repo.Delete(context.Background(), user.ID)
    if err != nil {
        t.Errorf("failed to cleanup: %v", err)
    }
}
```

### 6. Test Fixtures & Helpers

```go
// testutil/fixtures.go
package testutil

func CreateTestUser(t *testing.T, username string) *User {
    t.Helper() // Mark as helper function
    
    return &User{
        ID:       uint(rand.Intn(10000)),
        Username: username,
        Email:    username + "@example.com",
        Role:     "user",
    }
}

func CreateTestProject(t *testing.T, name string, userID uint) *Project {
    t.Helper()
    
    return &Project{
        ID:      uint(rand.Intn(10000)),
        Name:    name,
        OwnerID: userID,
        Status:  "active",
    }
}

// Usage
func TestProjectCreation(t *testing.T) {
    user := testutil.CreateTestUser(t, "alice")
    project := testutil.CreateTestProject(t, "test-project", user.ID)
    // ... test logic
}
```

### 7. Testing Concurrency

```go
func TestConcurrentAccess(t *testing.T) {
    manager := NewResourceManager()
    
    var wg sync.WaitGroup
    errors := make(chan error, 100)
    
    // Simulate 100 concurrent operations
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            
            err := manager.AllocateResource(id)
            if err != nil {
                errors <- err
            }
        }(i)
    }
    
    wg.Wait()
    close(errors)
    
    // Check for race conditions or errors
    for err := range errors {
        t.Errorf("unexpected error: %v", err)
    }
}

// Run with: go test -race ./...
```

### 8. Benchmarking

```go
func BenchmarkCreateUser(b *testing.B) {
    repo := NewMockUserRepository()
    service := NewUserService(repo)
    ctx := context.Background()
    
    user := &CreateUserRequest{
        Username: "benchuser",
        Email:    "bench@example.com",
    }
    
    b.ResetTimer() // Reset timer after setup
    
    for i := 0; i < b.N; i++ {
        _, err := service.CreateUser(ctx, user)
        if err != nil {
            b.Fatalf("unexpected error: %v", err)
        }
    }
}

func BenchmarkParallelCreateUser(b *testing.B) {
    repo := NewMockUserRepository()
    service := NewUserService(repo)
    ctx := context.Background()
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            user := &CreateUserRequest{
                Username: "benchuser",
                Email:    "bench@example.com",
            }
            service.CreateUser(ctx, user)
        }
    })
}
```

### 9. Testing Error Cases

```go
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name          string
        setupMock     func(*MockRepository)
        expectedError string
    }{
        {
            name: "database connection error",
            setupMock: func(m *MockRepository) {
                m.CreateFunc = func(ctx context.Context, u *User) error {
                    return errors.New("connection refused")
                }
            },
            expectedError: "connection refused",
        },
        {
            name: "constraint violation",
            setupMock: func(m *MockRepository) {
                m.CreateFunc = func(ctx context.Context, u *User) error {
                    return errors.New("duplicate key value")
                }
            },
            expectedError: "duplicate key",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := &MockRepository{}
            tt.setupMock(mock)
            
            service := NewService(mock)
            err := service.CreateUser(context.Background(), &User{})
            
            if err == nil {
                t.Error("expected error but got none")
            }
            if !strings.Contains(err.Error(), tt.expectedError) {
                t.Errorf("expected error containing %q, got %v", tt.expectedError, err)
            }
        })
    }
}
```

### 10. Test Coverage

```bash
# Generate coverage report
go test ./... -coverprofile=coverage.out

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html

# Set minimum coverage threshold
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//' | \
    awk '{if ($1 < 70) exit 1}'
```

## Testing Checklist

Before committing, ensure:

- [ ] All tests pass: `go test ./...`
- [ ] No race conditions: `go test -race ./...`
- [ ] Coverage meets threshold (≥70%): `go test -coverprofile=coverage.out`
- [ ] Table-driven tests for functions with multiple cases
- [ ] Error cases are tested
- [ ] Integration tests clean up resources
- [ ] Benchmarks for performance-critical code
- [ ] Tests are deterministic (no random failures)
- [ ] Mock external dependencies (DB, K8s, APIs)
- [ ] Tests run quickly (< 1s for unit tests)

## Common Testing Patterns

### Pattern: Setup and Teardown
```go
func TestMain(m *testing.M) {
    // Global setup
    setup()
    
    // Run tests
    code := m.Run()
    
    // Global teardown
    teardown()
    
    os.Exit(code)
}

func TestWithCleanup(t *testing.T) {
    resource := createResource()
    t.Cleanup(func() {
        cleanupResource(resource)
    })
    
    // Test code
}
```

### Pattern: Subtests
```go
func TestUserOperations(t *testing.T) {
    user := setupTestUser(t)
    
    t.Run("Create", func(t *testing.T) {
        // Test creation
    })
    
    t.Run("Update", func(t *testing.T) {
        // Test update
    })
    
    t.Run("Delete", func(t *testing.T) {
        // Test deletion
    })
}
```

### Pattern: Test Database with Docker
```go
func setupTestDB(t *testing.T) *gorm.DB {
    t.Helper()
    
    dsn := "host=localhost user=test password=test dbname=test_db port=5432"
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Skip("test database not available")
    }
    
    // Migrate schema
    db.AutoMigrate(&User{}, &Project{})
    
    return db
}
```

## Anti-Patterns to Avoid

```go
// Don't use time.Sleep in tests
time.Sleep(1 * time.Second) // Use channels or sync primitives instead

// Don't test implementation details
if len(manager.internalCache) != 5 { ... } // Test behavior, not internals

// Don't share state between tests
var globalUser *User // Each test should be independent

// Don't ignore cleanup
CreateResource() // Always cleanup in defer or t.Cleanup

// Don't use random test data without seeding
rand.Intn(100) // Use rand.Seed() or fixed values for reproducibility
```

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Advanced Testing with Go](https://www.youtube.com/watch?v=8hQG7QlcLBk)
