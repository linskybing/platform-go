---
name: error-handling-guide
description: Comprehensive error handling, custom error types, error wrapping, logging, and recovery patterns for platform-go
---

# Error Handling Guide

This skill provides comprehensive error handling patterns and best practices for platform-go.

## When to Use

Apply this skill when:
- Creating or calling functions that can fail
- Wrapping errors with additional context
- Implementing error recovery mechanisms
- Setting up error logging
- Handling panic recovery
- Creating custom error types
- Returning errors from goroutines
- Implementing retry logic

## Error Types & Custom Errors

### 1. Define Custom Error Types

```go
// Domain-specific error types
type ValidationError struct {
    Field   string
    Message string
}

func (e ValidationError) Error() string {
    return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Message)
}

type NotFoundError struct {
    ResourceType string
    ID           interface{}
}

func (e NotFoundError) Error() string {
    return fmt.Sprintf("%s with ID %v not found", e.ResourceType, e.ID)
}

type ConflictError struct {
    Message string
}

func (e ConflictError) Error() string {
    return fmt.Sprintf("conflict: %s", e.Message)
}

// Error constants for status mapping
var (
    ErrUnauthorized = errors.New("unauthorized")
    ErrForbidden    = errors.New("forbidden")
    ErrNotFound     = errors.New("not found")
    ErrConflict     = errors.New("conflict")
)
```

### 2. Error Wrapping & Context

```go
// Always wrap errors with context using %w
func CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // Validate input
    if err := validateRequest(req); err != nil {
        return nil, fmt.Errorf("invalid user request: %w", err)
    }
    
    // Create in database
    user, err := userRepo.Create(ctx, req)
    if err != nil {
        // Add context about what operation failed
        return nil, fmt.Errorf("failed to create user with email %s: %w", req.Email, err)
    }
    
    // Initialize related resources
    if err := initializeUserStorage(ctx, user.ID); err != nil {
        // Try to cleanup
        if cleanupErr := userRepo.Delete(ctx, user.ID); cleanupErr != nil {
            return nil, fmt.Errorf("failed to initialize storage for user (cleanup also failed): original=%w, cleanup=%w", err, cleanupErr)
        }
        return nil, fmt.Errorf("failed to initialize storage for user: %w", err)
    }
    
    return user, nil
}

// Use errors.Is for specific error checking
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
        return
    }
    
    user, err := h.userService.CreateUser(c.Request.Context(), &req)
    if err != nil {
        // Check for specific error type
        if errors.Is(err, ErrConflict) {
            c.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
            return
        }
        
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
        return
    }
    
    c.JSON(http.StatusCreated, user)
}

// Use errors.As for type assertions
func HandleError(err error) int {
    var validationErr ValidationError
    if errors.As(err, &validationErr) {
        return http.StatusBadRequest
    }
    
    var notFoundErr NotFoundError
    if errors.As(err, &notFoundErr) {
        return http.StatusNotFound
    }
    
    return http.StatusInternalServerError
}
```

## Error Logging

### 3. Structured Logging

```go
// Use structured logging for better debugging
import "log/slog"

func ProcessJob(ctx context.Context, jobID uint) error {
    logger := slog.With("job_id", jobID)
    
    job, err := jobRepo.Get(ctx, jobID)
    if err != nil {
        logger.Error("failed to fetch job", "error", err)
        return fmt.Errorf("failed to get job: %w", err)
    }
    
    result, err := executeJob(ctx, job)
    if err != nil {
        logger.Error("job execution failed",
            "error", err,
            "job_name", job.Name,
            "status", job.Status)
        return fmt.Errorf("job execution failed: %w", err)
    }
    
    logger.Info("job completed successfully",
        "duration_ms", time.Since(time.Now()).Milliseconds(),
        "result_count", len(result))
    
    return nil
}

// Log at appropriate levels
func Example() {
    logger := slog.Default()
    
    // Debug: Verbose information for debugging
    logger.Debug("starting operation", "param1", value1)
    
    // Info: General informational messages
    logger.Info("user created", "user_id", userID, "email", email)
    
    // Warn: Warning conditions
    logger.Warn("high memory usage", "percent", 85)
    
    // Error: Error conditions
    logger.Error("database connection failed", "error", err)
}
```

### 4. Error Recovery

```go
// Defer cleanup on error
func ProcessFile(ctx context.Context, filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("failed to open file: %w", err)
    }
    defer func() {
        if closeErr := file.Close(); closeErr != nil {
            // Log but don't overwrite original error
            log.Printf("failed to close file: %v", closeErr)
        }
    }()
    
    // Process file
    return processFileContent(file)
}

// Panic recovery for goroutines
func SafeGoroutine(fn func() error) <-chan error {
    errChan := make(chan error, 1)
    go func() {
        defer func() {
            if r := recover(); r != nil {
                errChan <- fmt.Errorf("panic recovered: %v", r)
            }
        }()
        errChan <- fn()
    }()
    return errChan
}

// Use in handlers
func (h *JobHandler) ExecuteJob(c *gin.Context) {
    jobID := c.Param("id")
    
    errChan := SafeGoroutine(func() error {
        return h.service.ExecuteJob(c.Request.Context(), jobID)
    })
    
    select {
    case err := <-errChan:
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        }
    case <-time.After(30 * time.Second):
        c.JSON(http.StatusRequestTimeout, gin.H{"error": "execution timeout"})
    }
}

// Recover in HTTP handlers
func PanicRecoveryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic recovered: %v", err)
                c.JSON(http.StatusInternalServerError, gin.H{
                    "error": "internal server error",
                })
            }
        }()
        c.Next()
    }
}
```

## Retry Logic

### 5. Retry Strategies

```go
// Exponential backoff retry
func RetryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
    var lastErr error
    backoff := 100 * time.Millisecond
    
    for attempt := 0; attempt < maxRetries; attempt++ {
        if err := fn(); err == nil {
            return nil
        } else {
            lastErr = err
        }
        
        // Check context cancellation
        select {
        case <-ctx.Done():
            return fmt.Errorf("retry cancelled: %w", ctx.Err())
        case <-time.After(backoff):
        }
        
        // Exponential backoff with jitter
        backoff = time.Duration(float64(backoff) * 1.5)
        if backoff > 30*time.Second {
            backoff = 30 * time.Second
        }
    }
    
    return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// Usage example
func (s *Service) GetUserWithRetry(ctx context.Context, userID uint) (*User, error) {
    var user *User
    err := RetryWithBackoff(ctx, 3, func() error {
        u, err := s.userRepo.Get(ctx, userID)
        if err != nil {
            return err
        }
        user = u
        return nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("failed to get user after retries: %w", err)
    }
    
    return user, nil
}

// Retry with exponential backoff for K8s operations
func (k *KubernetesManager) CreatePodWithRetry(ctx context.Context, pod *corev1.Pod) error {
    backoff := 100 * time.Millisecond
    
    for attempt := 0; attempt < 5; attempt++ {
        _, err := k.clientset.CoreV1().Pods(pod.Namespace).Create(ctx, pod, metav1.CreateOptions{})
        
        if err == nil {
            return nil
        }
        
        // Check for transient errors
        if isTransientError(err) {
            if attempt < 4 {
                time.Sleep(backoff)
                backoff = time.Duration(float64(backoff) * 1.5)
                continue
            }
        }
        
        return fmt.Errorf("failed to create pod: %w", err)
    }
    
    return fmt.Errorf("pod creation failed after retries")
}

func isTransientError(err error) bool {
    // Retry on connection errors, timeouts
    return errors.Is(err, context.DeadlineExceeded) ||
        strings.Contains(err.Error(), "connection refused")
}
```

## Error HTTP Status Mapping

### 6. Handler Error Mapping

```go
// Map domain errors to HTTP status codes
func MapErrorToHTTPStatus(err error) int {
    if err == nil {
        return http.StatusOK
    }
    
    if errors.Is(err, context.DeadlineExceeded) {
        return http.StatusGatewayTimeout
    }
    
    if errors.Is(err, context.Canceled) {
        return http.StatusRequestTimeout
    }
    
    var validationErr ValidationError
    if errors.As(err, &validationErr) {
        return http.StatusBadRequest
    }
    
    var notFoundErr NotFoundError
    if errors.As(err, &notFoundErr) {
        return http.StatusNotFound
    }
    
    var conflictErr ConflictError
    if errors.As(err, &conflictErr) {
        return http.StatusConflict
    }
    
    if errors.Is(err, ErrUnauthorized) {
        return http.StatusUnauthorized
    }
    
    if errors.Is(err, ErrForbidden) {
        return http.StatusForbidden
    }
    
    return http.StatusInternalServerError
}

// Generic error response builder
func (h *Handler) WriteError(c *gin.Context, err error) {
    statusCode := MapErrorToHTTPStatus(err)
    
    response := gin.H{
        "error":  err.Error(),
        "status": statusCode,
    }
    
    // Add additional details for validation errors
    var validationErr ValidationError
    if errors.As(err, &validationErr) {
        response["field"] = validationErr.Field
    }
    
    c.JSON(statusCode, response)
}
```

## Error Handling in Goroutines

### 7. Error Handling in Concurrent Code

```go
// Collect errors from multiple goroutines
func ProcessItemsConcurrently(ctx context.Context, items []Item) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(items))
    
    for _, item := range items {
        wg.Add(1)
        go func(i Item) {
            defer wg.Done()
            if err := ProcessItem(ctx, i); err != nil {
                errChan <- fmt.Errorf("failed to process item %d: %w", i.ID, err)
            }
        }(item)
    }
    
    go func() {
        wg.Wait()
        close(errChan)
    }()
    
    // Collect any errors
    var errors []error
    for err := range errChan {
        errors = append(errors, err)
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("processing failed with %d errors: %v", len(errors), errors)
    }
    
    return nil
}

// First error pattern
func ProcessWithFirstError(ctx context.Context, tasks []Task) error {
    errChan := make(chan error, len(tasks))
    var wg sync.WaitGroup
    
    for _, task := range tasks {
        wg.Add(1)
        go func(t Task) {
            defer wg.Done()
            if err := Execute(ctx, t); err != nil {
                errChan <- err
            }
        }(task)
    }
    
    // Return first error
    select {
    case err := <-errChan:
        return fmt.Errorf("execution failed: %w", err)
    default:
    }
    
    wg.Wait()
    return nil
}
```

## Error Handling Checklist

- [ ] All errors are wrapped with context using %w
- [ ] Custom error types are defined for domain-specific errors
- [ ] errors.Is() and errors.As() are used for error checking
- [ ] Errors are logged with appropriate context
- [ ] Panic recovery is implemented for goroutines
- [ ] Retry logic includes exponential backoff
- [ ] Cleanup/defer is used to ensure resource release
- [ ] HTTP status codes are properly mapped to errors
- [ ] Error messages do not leak sensitive information
- [ ] Timeouts are handled with context.DeadlineExceeded
- [ ] Goroutine errors are collected and reported
- [ ] Error messages are user-friendly in API responses

## Error Handling Anti-Patterns (FORBIDDEN)

- Never ignore errors with blank identifier (_) in production code
- Never panic in library code
- Never return a nil interface for error (return nil, not interface{})
- Never wrap errors more than twice in same layer
- Never log passwords or tokens when logging errors
- Never use generic errors.New without context
- Never return errors without adding context
- Never use error as control flow (using it for non-error conditions)
- Never catch all panics without proper handling
- Never lose the original error in error chain
