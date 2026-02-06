---
name: monitoring-observability
description: Logging, metrics collection, tracing, alerting, and observability best practices for production platform-go
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
---


# Monitoring and Observability

This skill provides comprehensive guidelines for logging, metrics, tracing, and alerting in platform-go.

## When to Use

Apply this skill when:
- Setting up application logging
- Implementing metrics collection
- Adding distributed tracing
- Configuring alerts and notifications
- Debugging production issues
- Monitoring performance metrics
- Tracking error rates
- Implementing health checks

## Structured Logging

### 1. Logging Best Practices

```go
import (
    "log/slog"
    "context"
)

// Initialize logger with structured format
func InitLogger() *slog.Logger {
    opts := &slog.HandlerOptions{
        Level:      slog.LevelInfo,
        AddSource:  true,
        ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
            if a.Key == slog.TimeKey {
                return slog.Attr{}  // Remove timestamp for Kubernetes
            }
            return a
        },
    }
    
    handler := slog.NewJSONHandler(os.Stdout, opts)
    return slog.New(handler)
}

// Use logger throughout application
func (s *UserService) CreateUser(ctx context.Context, username string) (*User, error) {
    logger := slog.With("operation", "create_user", "username", username)
    
    logger.Info("starting user creation")
    
    if username == "" {
        logger.Warn("empty username provided")
        return nil, fmt.Errorf("username required")
    }
    
    user, err := s.repo.Create(ctx, username)
    if err != nil {
        logger.Error("failed to create user", "error", err)
        return nil, fmt.Errorf("create failed: %w", err)
    }
    
    logger.Info("user created successfully", "user_id", user.ID)
    return user, nil
}

// Log levels usage
logger.Debug("detailed debug info")                     // Development only
logger.Info("important application events")            // Normal operations
logger.Warn("potentially problematic situations")       // Warning conditions
logger.Error("error conditions", "error", err)        // Error with context

// Never log sensitive information
func (h *Handler) LoginUser(c *gin.Context) {
    // GOOD: Log only non-sensitive info
    logger.Info("user login attempt", "username", username, "ip", c.ClientIP())
    
    // BAD: NEVER log passwords or tokens (FORBIDDEN)
    // logger.Info("login", "password", password)
    // logger.Info("token", token)
}
```

### 2. Contextual Logging

```go
// Add request-scoped logging
type contextKey string

const (
    requestIDKey contextKey = "request_id"
    userIDKey    contextKey = "user_id"
)

// Middleware to add context
func RequestContextMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        ctx := context.WithValue(c.Request.Context(), requestIDKey, requestID)
        
        c.Request = c.Request.WithContext(ctx)
        c.Header("X-Request-ID", requestID)
        
        c.Next()
    }
}

// Extract context in logger
func GetContextLogger(ctx context.Context) *slog.Logger {
    logger := slog.Default()
    
    if requestID, ok := ctx.Value(requestIDKey).(string); ok {
        logger = logger.With("request_id", requestID)
    }
    
    if userID, ok := ctx.Value(userIDKey).(uint); ok {
        logger = logger.With("user_id", userID)
    }
    
    return logger
}
```

## Metrics Collection

### 3. Prometheus Metrics

```go
import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

// Define metrics
var (
    // Counters - always increase
    httpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "platform_go_http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
    
    // Histograms - measure latency distributions
    httpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "platform_go_http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )
    
    // Gauges - current state
    activeDatabaseConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "platform_go_database_connections_active",
            Help: "Number of active database connections",
        },
    )
    
    // Job metrics
    jobsProcessed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "platform_go_jobs_processed_total",
            Help: "Total jobs processed",
        },
        []string{"status"},
    )
)

// HTTP middleware to record metrics
func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start).Seconds()
        
        httpRequestsTotal.WithLabelValues(
            c.Request.Method,
            c.Request.URL.Path,
            fmt.Sprintf("%d", c.Writer.Status()),
        ).Inc()
        
        httpRequestDuration.WithLabelValues(
            c.Request.Method,
            c.Request.URL.Path,
        ).Observe(duration)
    }
}

// Expose metrics endpoint
func RegisterMetricsEndpoints(r *gin.Engine) {
    r.GET("/metrics", gin.WrapH(promhttp.Handler()))
    r.GET("/health", HealthCheckHandler)
}

// Record job metrics
func (s *JobService) ProcessJob(ctx context.Context, jobID uint) error {
    defer func() {
        if err := recover(); err != nil {
            jobsProcessed.WithLabelValues("panic").Inc()
        }
    }()
    
    job, err := s.repo.Get(ctx, jobID)
    if err != nil {
        jobsProcessed.WithLabelValues("error").Inc()
        return err
    }
    
    if err := executeJob(ctx, job); err != nil {
        jobsProcessed.WithLabelValues("failed").Inc()
        return err
    }
    
    jobsProcessed.WithLabelValues("success").Inc()
    return nil
}
```

## Health Checks

### 4. Health Check Endpoints

```go
// Health check response
type HealthStatus struct {
    Status      string                 `json:"status"`
    Timestamp   time.Time              `json:"timestamp"`
    Checks      map[string]CheckResult `json:"checks"`
    Version     string                 `json:"version"`
}

type CheckResult struct {
    Status  string `json:"status"`
    Message string `json:"message,omitempty"`
    Duration string `json:"duration"`
}

// Implement health checks
func (h *HealthHandler) Check(c *gin.Context) {
    results := map[string]CheckResult{
        "database": h.checkDatabase(c.Request.Context()),
        "redis":    h.checkRedis(c.Request.Context()),
        "disk":     h.checkDiskSpace(),
        "memory":   h.checkMemory(),
    }
    
    status := "healthy"
    for _, result := range results {
        if result.Status != "healthy" {
            status = "degraded"
            break
        }
    }
    
    healthStatus := HealthStatus{
        Status:    status,
        Timestamp: time.Now(),
        Checks:    results,
        Version:   os.Getenv("APP_VERSION"),
    }
    
    statusCode := http.StatusOK
    if status != "healthy" {
        statusCode = http.StatusServiceUnavailable
    }
    
    c.JSON(statusCode, healthStatus)
}

func (h *HealthHandler) checkDatabase(ctx context.Context) CheckResult {
    start := time.Now()
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    sqlDB, _ := h.db.DB()
    err := sqlDB.PingContext(ctx)
    
    if err != nil {
        return CheckResult{
            Status:   "unhealthy",
            Message:  fmt.Sprintf("database unreachable: %v", err),
            Duration: time.Since(start).String(),
        }
    }
    
    return CheckResult{
        Status:   "healthy",
        Duration: time.Since(start).String(),
    }
}

func (h *HealthHandler) checkRedis(ctx context.Context) CheckResult {
    start := time.Now()
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    
    status := h.redis.Ping(ctx).Val()
    
    if status != "PONG" {
        return CheckResult{
            Status:   "degraded",
            Message:  "redis not responding",
            Duration: time.Since(start).String(),
        }
    }
    
    return CheckResult{
        Status:   "healthy",
        Duration: time.Since(start).String(),
    }
}

// Startup probe (K8s)
func (h *HealthHandler) StartupCheck(c *gin.Context) {
    result := h.checkDatabase(c.Request.Context())
    if result.Status == "healthy" {
        c.JSON(http.StatusOK, gin.H{"status": "ready"})
    } else {
        c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
    }
}

// Readiness probe (K8s)
func (h *HealthHandler) ReadinessCheck(c *gin.Context) {
    h.Check(c)
}

// Liveness probe (K8s)
func (h *HealthHandler) LivenessCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "alive"})
}
```

## Distributed Tracing

### 5. Tracing Implementation

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/sdk/trace"
    "go.opentelemetry.io/otel/sdk/resource"
)

// Initialize tracer provider
func InitTracer(ctx context.Context, serviceName string) (*trace.TracerProvider, error) {
    resource, err := resource.New(ctx,
        resource.WithAttributes(
            semconv.ServiceNameKey.String(serviceName),
            semconv.ServiceVersionKey.String("1.0.0"),
        ),
    )
    if err != nil {
        return nil, err
    }
    
    // Configure exporter (e.g., OTLP, Jaeger)
    exporter, err := otlptrace.New(ctx)
    if err != nil {
        return nil, err
    }
    
    provider := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
        trace.WithResource(resource),
    )
    
    otel.SetTracerProvider(provider)
    return provider, nil
}

// Record spans
func (s *UserService) GetUser(ctx context.Context, userID uint) (*User, error) {
    tracer := otel.Tracer("user-service")
    ctx, span := tracer.Start(ctx, "GetUser")
    defer span.End()
    
    span.SetAttribute("user_id", userID)
    
    // Call database
    user, err := s.repo.Get(ctx, userID)
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, "failed to get user")
        return nil, err
    }
    
    span.AddEvent("user retrieved")
    return user, nil
}
```

## Performance Monitoring

### 6. Benchmark & Profiling

```bash
# Run benchmarks
go test -bench=. -benchmem ./...

# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profiling
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## Observability Checklist

- [ ] Structured JSON logging is configured
- [ ] Request IDs are tracked through request lifecycle
- [ ] All errors are logged with context
- [ ] Sensitive data (passwords, tokens) is never logged
- [ ] Prometheus metrics are exposed at /metrics
- [ ] Health check endpoints are implemented
- [ ] Readiness and liveness probes for Kubernetes
- [ ] HTTP request/response times are tracked
- [ ] Database query performance is monitored
- [ ] Job processing success/failure rates tracked
- [ ] Distributed tracing is configured
- [ ] Error rates and patterns are monitored
- [ ] Memory and CPU usage is tracked
- [ ] Alerting rules are defined for critical metrics
- [ ] Log aggregation (ELK stack or Loki) configured

## Alerting Guidelines

- Alert on error rate > 5% in 5 minute window
- Alert on P95 latency > 1 second
- Alert on database connections > 80%
- Alert on memory usage > 80%
- Alert on job processing failure rate > 10%
- Alert on API response timeout > 30 seconds

## Anti-Patterns (FORBIDDEN)

- Never log sensitive data (passwords, tokens, PII)
- Never use print statements for logging (use slog)
- Never create unlimited goroutines without monitoring
- Never skip error logging
- Never ignore panic in goroutines
- Never log at Debug level in production
