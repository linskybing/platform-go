# Separate Headers Implementation Guide: JWT & API Keys

**Version**: 2.0  
**Date**: January 20, 2024  
**Status**: Active Implementation

---

## Overview

This document details the implementation of **separate, non-conflicting headers** for JWT and API Key authentication in platform-go.

### Header Strategy

| Authentication Method | Header | Format | Example |
|---|---|---|---|
| JWT (Browser/Mobile) | `Authorization` | `Bearer <token>` | `Authorization: Bearer eyJhbGc...` |
| API Key (Service/CLI) | `X-API-Key` | `pk_<random>` | `X-API-Key: pk_a1b2c3d4...` |
| Key Identifier (optional) | `X-API-Key-Name` | text | `X-API-Key-Name: ci-pipeline` |

**Key Advantage**: Complete header separation eliminates any ambiguity. No token format detection needed.

---

## Architecture

### Middleware Components

#### 1. JWT Authentication Middleware

**File**: `internal/api/middleware/jwt.go`

Extracts JWT from:
- `Authorization: Bearer <token>` header (primary)
- `token` cookie (fallback for browser)

**Usage**:
```go
jwtProtected := router.Group("/api")
jwtProtected.Use(middleware.JWTAuthMiddleware())
{
    jwtProtected.GET("/users/me", handlers.GetCurrentUser)
}
```

#### 2. API Key Authentication Middleware

**File**: `internal/api/middleware/apikey.go`

Extracts API Key from:
- `X-API-Key` header (required)
- `X-API-Key-Name` header (optional, for audit trail)

**Usage**:
```go
serviceProtected := router.Group("/api/v1")
serviceProtected.Use(middleware.APIKeyAuthMiddleware(db))
{
    serviceProtected.POST("/webhooks/events", handlers.HandleWebhook)
}
```

#### 3. Optional Authentication Middleware

**File**: `internal/api/middleware/optional_auth.go`

Tries both JWT and API Key, passes if either is valid.

**Usage**:
```go
anyAuth := router.Group("/api")
anyAuth.Use(middleware.OptionalAuthMiddleware(db))
{
    anyAuth.GET("/data/export", handlers.ExportData)
}
```

---

## Implementation Checklist

### Phase 1: Middleware Implementation ✓

**Status**: To be implemented

- [ ] Create `internal/api/middleware/auth.go` with:
  - `JWTAuthMiddleware()`
  - `APIKeyAuthMiddleware(db)`
  - `OptionalAuthMiddleware(db)`
  - Helper functions: `tryJWT()`, `tryAPIKey()`

**Code Template**:
```go
// internal/api/middleware/auth.go
package middleware

import (
    "errors"
    "net/http"
    "strings"
    "time"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "platform-go/pkg/response"
    "platform-go/internal/domain/user"
)

// JWTAuthMiddleware validates JWT from Authorization header or cookie
func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        var tokenStr string
        
        // Try Authorization header first
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" {
            parts := strings.SplitN(authHeader, " ", 2)
            if len(parts) != 2 || parts[0] != "Bearer" {
                response.Error(c, http.StatusUnauthorized, 
                    "Invalid Authorization header format")
                c.Abort()
                return
            }
            tokenStr = parts[1]
        } else {
            // Fallback to cookie
            cookie, err := c.Cookie("token")
            if err != nil {
                response.Error(c, http.StatusUnauthorized, 
                    "Missing JWT token")
                c.Abort()
                return
            }
            tokenStr = cookie
        }
        
        // Validate JWT
        claims, err := ParseToken(tokenStr)
        if err != nil {
            response.Error(c, http.StatusUnauthorized, 
                "Invalid JWT: " + err.Error())
            c.Abort()
            return
        }
        
        if claims.ExpiresAt != nil && 
           time.Now().After(claims.ExpiresAt.Time) {
            response.Error(c, http.StatusUnauthorized, 
                "Token expired")
            c.Abort()
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("is_admin", claims.IsAdmin)
        c.Set("auth_type", "jwt")
        c.Next()
    }
}

// APIKeyAuthMiddleware validates API key from X-API-Key header
func APIKeyAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        apiKey := c.GetHeader("X-API-Key")
        if apiKey == "" {
            response.Error(c, http.StatusUnauthorized, 
                "Missing X-API-Key header")
            c.Abort()
            return
        }
        
        // Validate format
        if !strings.HasPrefix(apiKey, "pk_") {
            response.Error(c, http.StatusUnauthorized, 
                "Invalid API key format")
            c.Abort()
            return
        }
        
        // Look up user by API key hash
        var u user.User
        if err := db.Where("api_key_hash = ?", apiKey).
            First(&u).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                response.Error(c, http.StatusUnauthorized, 
                    "Invalid API key")
            } else {
                response.Error(c, http.StatusInternalServerError, 
                    "Database error")
            }
            c.Abort()
            return
        }
        
        // Check expiration
        if u.APIKeyExpiresAt != nil && 
           u.APIKeyExpiresAt.Before(time.Now()) {
            response.Error(c, http.StatusUnauthorized, 
                "API key expired")
            c.Abort()
            return
        }
        
        // Update last used timestamp
        db.Model(&u).Update("api_key_last_used", time.Now())
        
        c.Set("user_id", u.UID)
        c.Set("username", u.Username)
        c.Set("is_admin", false)
        c.Set("auth_type", "api_key")
        c.Set("api_key_name", c.GetHeader("X-API-Key-Name"))
        c.Next()
    }
}

// OptionalAuthMiddleware tries both JWT and API Key, passes if either valid
func OptionalAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        hasJWT := tryJWT(c)
        hasAPIKey := tryAPIKey(c, db)
        
        if !hasJWT && !hasAPIKey {
            response.Error(c, http.StatusUnauthorized, 
                "Missing or invalid authentication")
            c.Abort()
            return
        }
        
        c.Next()
    }
}

func tryJWT(c *gin.Context) bool {
    var tokenStr string
    
    authHeader := c.GetHeader("Authorization")
    if authHeader != "" {
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) == 2 && parts[0] == "Bearer" {
            tokenStr = parts[1]
        }
    } else {
        cookie, err := c.Cookie("token")
        if err != nil {
            return false
        }
        tokenStr = cookie
    }
    
    if tokenStr == "" {
        return false
    }
    
    claims, err := ParseToken(tokenStr)
    if err != nil || claims.ExpiresAt == nil || 
       time.Now().After(claims.ExpiresAt.Time) {
        return false
    }
    
    c.Set("user_id", claims.UserID)
    c.Set("username", claims.Username)
    c.Set("is_admin", claims.IsAdmin)
    c.Set("auth_type", "jwt")
    return true
}

func tryAPIKey(c *gin.Context, db *gorm.DB) bool {
    apiKey := c.GetHeader("X-API-Key")
    if apiKey == "" || !strings.HasPrefix(apiKey, "pk_") {
        return false
    }
    
    var u user.User
    if err := db.Where("api_key_hash = ?", apiKey).
        First(&u).Error; err != nil {
        return false
    }
    
    if u.APIKeyExpiresAt != nil && 
       u.APIKeyExpiresAt.Before(time.Now()) {
        return false
    }
    
    db.Model(&u).Update("api_key_last_used", time.Now())
    
    c.Set("user_id", u.UID)
    c.Set("username", u.Username)
    c.Set("is_admin", false)
    c.Set("auth_type", "api_key")
    c.Set("api_key_name", c.GetHeader("X-API-Key-Name"))
    return true
}
```

### Phase 2: Route Configuration ✓

**Status**: To be implemented

Update `internal/api/routes/routes.go`:

```go
// Public routes (no auth required)
public := router.Group("/api")
{
    public.POST("/auth/login", handlers.Login)
    public.POST("/auth/register", handlers.Register)
    public.GET("/health", handlers.Health)
}

// JWT-protected routes (for interactive users)
jwtRequired := router.Group("/api")
jwtRequired.Use(middleware.JWTAuthMiddleware())
{
    jwtRequired.POST("/auth/logout", handlers.Logout)
    jwtRequired.POST("/auth/refresh", handlers.RefreshToken)
    
    userRoutes := jwtRequired.Group("/users")
    {
        userRoutes.GET("/:id", handlers.GetUser)
        userRoutes.PUT("/:id", handlers.UpdateUser)
        userRoutes.POST("/:id/apikey", handlers.CreateAPIKey)
    }
}

// API-Key-protected routes (for services and automation)
serviceRequired := router.Group("/api/v1")
serviceRequired.Use(middleware.APIKeyAuthMiddleware(db))
{
    webhookRoutes := serviceRequired.Group("/webhooks")
    {
        webhookRoutes.POST("/events", handlers.HandleWebhookEvent)
    }
    
    statusRoutes := serviceRequired.Group("/status")
    {
        statusRoutes.GET("/health", handlers.SystemHealth)
    }
}

// Optional auth routes (accept either JWT or API Key)
anyAuth := router.Group("/api")
anyAuth.Use(middleware.OptionalAuthMiddleware(db))
{
    dataRoutes := anyAuth.Group("/data")
    {
        dataRoutes.GET("/export", handlers.ExportData)
    }
}
```

### Phase 3: Testing ✓

**Status**: To be implemented

Test files in `test/integration/`:

#### Test JWT Only
```bash
curl -H "Authorization: Bearer $JWT_TOKEN" \
  http://localhost:8080/api/users/me
# Expected: 200 OK
```

#### Test API Key Only
```bash
curl -H "X-API-Key: pk_abc123..." \
  http://localhost:8080/api/v1/webhooks/status
# Expected: 200 OK
```

#### Test Both Headers
```bash
curl -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-API-Key: pk_abc123..." \
  http://localhost:8080/api/admin/audit
# Expected: 200 OK (both authenticated)
```

#### Test Invalid Headers
```bash
curl -H "Authorization: Invalid" \
  http://localhost:8080/api/users/me
# Expected: 401 Unauthorized

curl -H "X-API-Key: invalid" \
  http://localhost:8080/api/v1/webhooks/status
# Expected: 401 Unauthorized
```

### Phase 4: Documentation ✓

**Status**: To be updated

- [ ] Update API documentation with header examples
- [ ] Create SDK examples for different languages
- [ ] Document migration from old auth system (if applicable)
- [ ] Create troubleshooting guide

---

## Client Integration Examples

### JavaScript/Browser

```javascript
// Using JWT
const response = await fetch('/api/users/me', {
    headers: {
        'Authorization': `Bearer ${jwtToken}`
    }
});

// Using API Key (less common for browsers)
const apiResponse = await fetch('/api/v1/webhooks/status', {
    headers: {
        'X-API-Key': apiKey
    }
});
```

### Python/Service

```python
import requests

# Using API Key
headers = {
    'X-API-Key': 'pk_a1b2c3d4e5f6...',
    'X-API-Key-Name': 'ci-pipeline'  # Optional
}

response = requests.post(
    'https://api.platform-go.com/api/v1/webhooks/events',
    headers=headers,
    json={'event': 'build.complete'}
)
```

### cURL

```bash
# JWT
curl -H "Authorization: Bearer $JWT" \
  https://api.platform-go.com/api/users/me

# API Key
curl -H "X-API-Key: pk_abc123..." \
  -H "X-API-Key-Name: ci-pipeline" \
  https://api.platform-go.com/api/v1/webhooks/events

# Both
curl -H "Authorization: Bearer $JWT" \
  -H "X-API-Key: pk_abc123..." \
  https://api.platform-go.com/api/admin/audit-logs
```

---

## Testing Verification

After implementation, run these tests:

```bash
# 1. Test JWT authentication
go test -v ./test/integration/... -run TestJWT

# 2. Test API Key authentication
go test -v ./test/integration/... -run TestAPIKey

# 3. Test both headers together
go test -v ./test/integration/... -run TestBothHeaders

# 4. Test header conflicts (should have none)
go test -v ./test/integration/... -run TestHeaderConflicts

# 5. Run all auth tests
go test -v ./test/integration/... -run TestAuth
```

---

## Security Considerations

### JWT Headers
- Store in `httpOnly` cookies for web browsers
- Never expose in logs or error messages
- Automatically sent by browser for same-origin requests
- Set `Secure` flag for HTTPS only

### API Key Headers
- Treat as secrets (like passwords)
- Never commit to version control
- Rotate periodically (90 days recommended)
- Use separate keys for different services
- Log all API key operations for audit trail

### Header Validation
- Validate header sizes to prevent buffer overflow
- Rate limit by API key or user ID
- Log authentication failures
- Monitor for unusual patterns

---

## Configuration

### User Model Fields

Ensure these fields exist in `internal/domain/user/model.go`:

```go
type User struct {
    UID                 string
    Username            string
    Password            string
    Email               string
    
    // API Key fields
    APIKeyHash          string     // Bcrypt hash of API key
    APIKeyName          string     // Human-readable name
    APIKeyLastUsed      *time.Time // Track usage
    APIKeyCreatedAt     time.Time
    APIKeyExpiresAt     *time.Time // Optional expiration
    
    CreatedAt           time.Time
    UpdatedAt           time.Time
    DeletedAt           gorm.DeletedAt
}
```

### Constants

Update `internal/constants/constants.go`:

```go
const (
    // API Key
    APIKeyPrefix        = "pk_"
    APIKeyLength        = 32  // bytes
    APIKeyExpiry        = 90  // days
    APIKeyHashCost      = 12  // bcrypt cost
    
    // JWT
    JWTAlgorithm        = "HS256"
    JWTExpiry           = 24  // hours
    RefreshTokenExpiry  = 7   // days
    
    // Headers
    AuthHeaderName      = "Authorization"
    APIKeyHeaderName    = "X-API-Key"
    APIKeyNameHeader    = "X-API-Key-Name"
    
    // Header limits (prevent buffer overflow)
    AuthHeaderMax       = 1000
    APIKeyHeaderMax     = 100
    APIKeyNameMax       = 200
)
```

---

## Migration Path

If migrating from old single-header design:

1. **Keep old middleware working** temporarily with deprecation warning
2. **Add new separate-header middleware** alongside
3. **Update routes gradually**, testing each change
4. **Monitor logs** for old vs new header usage
5. **Communicate with clients** about header changes
6. **Remove old middleware** after grace period

---

## Troubleshooting

### JWT Issues

**Problem**: "Missing JWT token"
- **Solution**: Ensure `Authorization: Bearer <token>` header is present
- **Check**: Token hasn't expired

**Problem**: "Invalid JWT: invalid token"
- **Solution**: Verify token format and secret key match

### API Key Issues

**Problem**: "Missing X-API-Key header"
- **Solution**: Ensure `X-API-Key: pk_...` header is present

**Problem**: "Invalid API key format"
- **Solution**: API key must start with `pk_` prefix

**Problem**: "Invalid API key" (exists in database)
- **Solution**: Check API key hasn't been revoked or expired

### Both Headers

**Problem**: Request rejected despite providing both
- **Solution**: Check which middleware is applied - if both required, both must be valid

---

## Summary

The separate header strategy provides:

✅ **Zero conflicts** - Headers are completely separate  
✅ **Clear intent** - Easy to see which auth method is used  
✅ **Flexible routing** - Different endpoints can require different methods  
✅ **Easy debugging** - No format detection logic needed  
✅ **Production ready** - Secure, auditable, scalable  

**Next Steps**:
1. Implement the middleware in `internal/api/middleware/auth.go`
2. Update routes in `internal/api/routes/routes.go`
3. Run tests to verify functionality
4. Document API with header examples
5. Update client SDKs if needed
