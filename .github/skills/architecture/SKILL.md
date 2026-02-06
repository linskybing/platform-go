---
name: architecture
description: System architecture, API design, database patterns, and scalability for platform-go
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
  consolidated_from:
    - api-design-patterns
    - file-structure-guidelines
    - database-best-practices
    - production-readiness-checklist
---

# Architecture Excellence

Comprehensive guidelines for RESTful API design, database optimization, scalable system architecture, and production readiness.

## System Architecture

### Layered Architecture
```
┌─────────────────────────────────────┐
│         API Handlers (HTTP)         │ < 50 lines/handler
├─────────────────────────────────────┤
│      Services/Application Layer     │ Business logic (< 200 lines)
├─────────────────────────────────────┤
│    Domain Models & Interfaces       │ Core entities
├─────────────────────────────────────┤
│      Repository/Data Access         │ Database operations
├─────────────────────────────────────┤
│      External Services (K8s, etc)   │ Kubernetes, MinIO
└─────────────────────────────────────┘
```

### Domain-Driven Design
- **Domain Layer**: Core business models, no framework dependencies
- **Application Layer**: Use cases, orchestration, < 200 lines per file
- **Infrastructure Layer**: Database, external APIs, framework specifics
- **API Layer**: HTTP handlers, request validation

## RESTful API Design

### Endpoint Patterns
```
GET    /api/v1/users              # List (with pagination)
GET    /api/v1/users/{id}         # Retrieve
POST   /api/v1/users              # Create
PUT    /api/v1/users/{id}         # Full update
PATCH  /api/v1/users/{id}         # Partial update
DELETE /api/v1/users/{id}         # Delete
```

### Response Format
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation completed",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Error Handling
```json
{
  "success": false,
  "error": {
    "code": "INVALID_INPUT",
    "message": "Name is required",
    "details": { "field": "name" }
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Status Codes
- **200**: Successful GET/PUT/PATCH
- **201**: Successful POST (resource created)
- **204**: Successful DELETE (no content)
- **400**: Invalid request
- **401**: Unauthorized
- **403**: Forbidden
- **404**: Not found
- **409**: Conflict (e.g., duplicate)
- **500**: Server error

## Gin Framework Best Practices

### Handler Structure
```go
// handlers/user.go (< 50 lines)
func CreateUser(c *gin.Context) {
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
        return
    }
    
    user, err := h.userService.Create(c.Context(), &req)
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
        return
    }
    
    response.Success(c, http.StatusCreated, user, "User created")
}
```

### Middleware Chain
```go
router.Use(
    middleware.Logger(),
    middleware.Recovery(),
    middleware.CORSHandler(),
)

// Protected routes with auth middleware
protected := router.Group("")
protected.Use(middleware.Auth())
protected.POST("/users", handlers.CreateUser)
```

## Database Design

### Schema Principles
- Use primary key (id, UUID)
- Add created_at, updated_at timestamps
- Use indexes for frequently queried fields
- Enforce foreign key constraints
- Use GORM auto-migration for schema management

### Query Optimization
```go
// Good: pagination + indexes
users, err := db.
    Where("status = ?", "active").
    Order("created_at DESC").
    Limit(pageSize).
    Offset(offset).
    Find(&users).
    Error

// Bad: loading all records
users, err := db.Find(&allUsers).Error // 10M records!
```

### Transaction Management
```go
tx := db.BeginTx(ctx, &sql.TxOptions{
    Isolation: sql.LevelRepeatableRead,
})

if err := createUser(tx); err != nil {
    tx.Rollback()
    return err
}

return tx.Commit().Error
```

## Production Readiness

### Pre-Deployment Checklist
- [ ] Tests passing (70%+ coverage)
- [ ] Load testing done (> 1000 req/s)
- [ ] Database migrations verified
- [ ] API documentation complete
- [ ] Error handling comprehensive
- [ ] Logging and monitoring setup
- [ ] Security scan passed
- [ ] Performance benchmarks baseline set

### Monitoring Points
- Request latency (p50, p95, p99)
- Error rates by endpoint
- Database query performance
- Goroutine count
- Memory usage
- Disk I/O

### Deployment Strategy
1. **Blue-Green Deployment**: Parallel environments
2. **Canary Releases**: Gradual rollout (10% → 50% → 100%)
3. **Health Checks**: Readiness and liveness probes
4. **Rollback Plan**: Quick reversion mechanism

## Scalability Patterns

### Horizontal Scaling
- Stateless service design
- Load balancing (round-robin)
- Session storage in Redis
- Database connection pooling

### Vertical Scaling
- Goroutine limits
- Memory pooling
- Efficient algorithms (O(n) vs O(n²))
- Index optimization

### Caching Strategy
- User authentication tokens → Redis (5 min TTL)
- Project lists → Redis (1 hour TTL)
- Database queries → In-memory (with invalidation)
- Static content → CDN

## Security Architecture

### Secrets Management
- Use environment variables for secrets
- Never commit credentials
- Rotate API keys regularly
- Use secure password hashing (bcrypt)

### API Security
- HTTPS only in production
- Rate limiting per IP/user
- Input validation on all endpoints
- SQL injection prevention (parameterized queries)
- CORS properly configured

## Tools & Validation

### Pre-Commit Checks
```bash
# Comprehensive architecture validation
bash .github/skills-consolidated/architecture/scripts/validate-architecture.sh

# Database migration validation
bash .github/skills-consolidated/architecture/scripts/validate-migrations.sh

# API contract verification
bash .github/skills-consolidated/architecture/scripts/validate-api.sh
```

## References
- REST API Best Practices: https://restfulapi.net/
- GORM Documentation: https://gorm.io/
- Gin Web Framework: https://github.com/gin-gonic/gin
- Domain-Driven Design: https://www.domainlanguage.com/ddd/
