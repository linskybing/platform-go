---
name: unified-authentication-strategy
description: Comprehensive authentication strategy supporting both JWT tokens (Authorization header) and API keys (X-API-Key header) without conflicts, with clear routing rules for different client types (web browsers, mobile, automation, services).
license: Proprietary
metadata:
  author: platform-go
  version: "2.0"
  product: platform-go
  updated: "2024-01-20"
compatibility: Designed for platform-go API v1.0+
---

# Unified Authentication Strategy: Separate Headers for JWT & API Keys

## Overview

This skill provides a unified authentication approach that supports both JWT tokens and API keys using **separate, non-conflicting headers**. This design eliminates any ambiguity and allows simultaneous use of both authentication methods.

**When to use this skill:**
- Setting up authentication routes (login, logout, token refresh)
- Implementing API key management for service-to-service communication
- Protecting endpoints with authentication middleware
- Choosing appropriate authentication method for different client types
- Managing token expiration and key rotation
- Supporting simultaneous multi-factor authentication scenarios

---

## Step 1: Authentication Methods - Clear Separation

### JWT (JSON Web Tokens) - Browser & Mobile

**Use Case**: Interactive user sessions (web browsers, mobile apps)

**Header Format**: `Authorization: Bearer <jwt-token>`

**Advantages**:
- Stateless authentication (no server session storage)
- Contains user information in claims
- Automatic expiration built-in
- Can be stored in httpOnly cookies
- Standard OAuth2/OpenID Connect pattern

**Token Example**:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ1c2VyMTIzIn0.TJVA95OrM7E2cBab30RMHrHDcEfxjoYZgeFONFh7HgQ
```

### API Keys - Service-to-Service & Automation

**Use Case**: Machine-to-machine communication, automation tools, CLI

**Header Formats**:
```
X-API-Key: pk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
X-API-Key-Name: ci-pipeline (optional, for identification)
```

**Advantages**:
- Long-lived credentials for services
- No expiration by default (optional rotation policy)
- Audit trail tracking
- Can be per-service or per-user
- Completely separate from JWT headers

---

## Step 2: Header Design - Zero Conflicts

### Header Mapping

| Purpose | Header Name | Format | Example |
|---------|-------------|--------|---------|
| JWT Token | `Authorization` | `Bearer <token>` | `Bearer eyJhbGc...` |
| API Key | `X-API-Key` | `pk_<random>` | `pk_abc123...` |
| Key Name (optional) | `X-API-Key-Name` | text | `production-key` |

### Request Examples

**JWT in Header**:
```http
GET /api/users HTTP/1.1
Host: api.platform-go.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**JWT in Cookie**:
```http
GET /api/users HTTP/1.1
Host: api.platform-go.com
Cookie: token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**API Key in Header**:
```http
GET /api/webhooks HTTP/1.1
Host: api.platform-go.com
X-API-Key: pk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
X-API-Key-Name: ci-pipeline
```

**Both Simultaneously** (concurrent usage):
```http
GET /api/admin/audit HTTP/1.1
Host: api.platform-go.com
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
X-API-Key: pk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
X-API-Key-Name: ci-pipeline
```

---

## Step 3: Implementation Architecture

### Middleware Stack

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
        c.Next()
    }
}

// OptionalAuthMiddleware tries both JWT and API Key, passes if either valid
// Allows both to be present simultaneously
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
    return true
}
```

---

## Step 4: Route Configuration

### Public Routes (No Auth)

```go
// routes/routes.go - Routes accessible without authentication
public := router.Group("/api")
{
    public.POST("/auth/login", handlers.Login)
    public.POST("/auth/register", handlers.Register)
    public.GET("/health", handlers.Health)
    public.GET("/docs", handlers.ApiDocs)
}
```

### JWT-Protected Routes (Browsers, Mobile)

```go
// routes/routes.go - Routes requiring JWT
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
        userRoutes.DELETE("/:id/apikey", handlers.RevokeAPIKey)
    }
}
```

### API-Key-Protected Routes (Services, Automation)

```go
// routes/routes.go - Routes requiring API Key
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
        statusRoutes.GET("/metrics", handlers.GetMetrics)
    }
}
```

### Optional Auth (Either JWT or API Key)

```go
// routes/routes.go - Routes accepting either JWT or API Key
anyAuth := router.Group("/api")
anyAuth.Use(middleware.OptionalAuthMiddleware(db))
{
    dataRoutes := anyAuth.Group("/data")
    {
        dataRoutes.GET("/export", handlers.ExportData)
        dataRoutes.POST("/import", handlers.ImportData)
    }
}
```

---

## Step 5: Client Usage Examples

### JavaScript/Browser (JWT)

```javascript
// Login and get JWT
const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username: 'user', password: 'pass' })
});

const { token } = await response.json();

// Store in httpOnly cookie (set by server)
// Subsequent requests automatically include cookie

// Or use manually in requests
const userData = await fetch('/api/users/me', {
    headers: { 'Authorization': `Bearer ${token}` }
});
```

### Python/CLI (API Key)

```python
import requests

# Use API key in requests
api_key = "pk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6"
headers = {
    "X-API-Key": api_key,
    "X-API-Key-Name": "ci-pipeline"  # Optional
}

response = requests.post(
    "https://api.platform-go.com/api/v1/webhooks/events",
    headers=headers,
    json={"event": "build.complete"}
)
```

### cURL Examples

```bash
# Using JWT
curl -X GET \
  -H "Authorization: Bearer eyJhbGc..." \
  https://api.platform-go.com/api/users/me

# Using API key
curl -X POST \
  -H "X-API-Key: pk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6" \
  -H "X-API-Key-Name: ci-pipeline" \
  -H "Content-Type: application/json" \
  -d '{"event":"build.complete"}' \
  https://api.platform-go.com/api/v1/webhooks/events

# Using both (multi-factor authentication)
curl -X GET \
  -H "Authorization: Bearer eyJhbGc..." \
  -H "X-API-Key: pk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6" \
  https://api.platform-go.com/api/admin/audit-logs
```

### Both Simultaneously (Admin Dashboard)

```python
# Request with both JWT and API Key for multi-factor authentication
headers = {
    "Authorization": f"Bearer {jwt_token}",
    "X-API-Key": api_key,
    "X-API-Key-Name": "admin-dashboard"
}
response = requests.get(
    "https://api.platform-go.com/api/admin/audit-logs",
    headers=headers
)
```

---

## Step 6: Security Implementation Details

### JWT Security

```go
const (
    JWTAlgorithm = "HS256"
    JWTExpiry = 24 * time.Hour
    RefreshTokenExpiry = 7 * 24 * time.Hour
    JWTIssuer = "platform-go"
)

// Secure cookie settings
func SetTokenCookie(c *gin.Context, token string) {
    c.SetCookie(
        "token",
        token,
        int((24 * time.Hour).Seconds()),
        "/",
        ".platform-go.com",
        true,  // Secure (HTTPS only)
        true,  // HttpOnly
    )
}
```

### API Key Security

```go
const (
    APIKeyPrefix = "pk_"
    APIKeyLength = 32  // bytes
    APIKeyExpiry = 90 * 24 * time.Hour  // 90 days
    APIKeyHashCost = 12  // bcrypt cost
)

// API key format: pk_<32-byte-base64>
// Example: pk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6

// Generate API key
func GenerateAPIKey() string {
    randomBytes := make([]byte, 32)
    if _, err := rand.Read(randomBytes); err != nil {
        return ""
    }
    return APIKeyPrefix + base64.StdEncoding.EncodeToString(randomBytes)
}

// Hash API key for storage
func HashAPIKey(plainKey string) (string, error) {
    return bcrypt.GenerateFromPassword([]byte(plainKey), APIKeyHashCost)
}
```

### Header Validation

```go
const (
    AuthHeaderMax = 1000        // Bearer token + JWT max length
    APIKeyHeaderMax = 100       // pk_<64chars>
    APIKeyNameMax = 200         // Human readable name
)

// Validate header sizes to prevent buffer overflow
func ValidateAPIKeyHeader(apiKey string) error {
    if len(apiKey) == 0 || len(apiKey) > APIKeyHeaderMax {
        return errors.New("invalid API key header size")
    }
    if !strings.HasPrefix(apiKey, APIKeyPrefix) {
        return errors.New("invalid API key prefix")
    }
    return nil
}
```

---

## Step 7: Key Features

### 1. No Header Conflicts

- JWT uses `Authorization: Bearer` header exclusively
- API Key uses `X-API-Key` header exclusively
- Zero ambiguity about which authentication method is used
- Clear, explicit intent in every request

### 2. Simultaneous Support

```go
// Both headers can be present and validated
// - JWT will be extracted from Authorization header
// - API Key will be extracted from X-API-Key header
// - Both contexts will be available in c.Get("auth_type")
```

### 3. Flexible Middleware Stack

```
Public Endpoints:
  No middleware
  
JWT-Only Endpoints:
  JWTAuthMiddleware() 
  
API-Key-Only Endpoints:
  APIKeyAuthMiddleware(db)
  
Optional Auth Endpoints:
  OptionalAuthMiddleware(db)
  → Tries both, passes if either valid
  
Multi-Factor Endpoints:
  JWTAuthMiddleware() + APIKeyAuthMiddleware(db)
  → Requires both to be present and valid
```

### 4. Audit Trail

```go
// Every API key operation is logged
LogAPIKeyCreated(db, userID, keyName)
LogAPIKeyRevoked(db, userID, keyName)
LogAPIKeyLastUsed(db, userID, timestamp)
LogAPIKeyExpired(db, userID, keyName)
```

---

## Step 8: Implementation Checklist

### Core Implementation
- [ ] JWT middleware with token parsing
- [ ] API Key model with fields (hash, expiry, last_used)
- [ ] JWTAuthMiddleware function (separate headers)
- [ ] APIKeyAuthMiddleware function (X-API-Key header)
- [ ] OptionalAuthMiddleware for either/or scenarios
- [ ] Protected routes registration
- [ ] Public routes registration

### API Key Management
- [ ] API Key generation service
- [ ] API Key revocation endpoint
- [ ] API Key rotation endpoint
- [ ] API Key expiration notifications
- [ ] Audit logging for key operations

### Security
- [ ] Bcrypt hashing for API keys
- [ ] httpOnly cookies for JWT
- [ ] Header size validation
- [ ] Rate limiting per API key
- [ ] Audit logs for all auth operations

### Testing
- [ ] JWT authentication tests
- [ ] API Key authentication tests
- [ ] Token expiration tests
- [ ] **Separate header tests** (no conflicts)
- [ ] Concurrent auth methods tests
- [ ] Both headers present tests

### Documentation
- [ ] API authentication guide
- [ ] Client integration examples
- [ ] Header reference documentation
- [ ] Security considerations
- [ ] Troubleshooting guide

---

## Step 9: Quick Reference

### Header Usage Matrix

| Scenario | Headers | Example |
|----------|---------|---------|
| Browser/Mobile User | `Authorization: Bearer` | JWT token |
| Service/CLI | `X-API-Key` | `pk_...` |
| Optional ID | `X-API-Key-Name` | `ci-pipeline` |
| Multi-Auth | Both | Both present |

### Middleware Selection

```
Endpoint Type          | Middleware
-----------------------+-----------------------------------
User Dashboard         | JWTAuthMiddleware()
API Client             | APIKeyAuthMiddleware(db)
Webhook Handler        | APIKeyAuthMiddleware(db)
Data Export            | OptionalAuthMiddleware(db)
Admin Panel            | JWTAuthMiddleware() + role check
Service Health         | APIKeyAuthMiddleware(db)
Multi-Factor Admin     | Both middlewares + validation
```

### Common cURL Commands

```bash
# Login with JWT
curl -X POST -d '{"username":"user","password":"pass"}' \
  https://api.platform-go.com/api/auth/login

# Create API key (requires JWT)
curl -X POST \
  -H "Authorization: Bearer $JWT" \
  https://api.platform-go.com/api/users/$USER_ID/apikey

# Use API key
curl -H "X-API-Key: pk_abc123..." \
  https://api.platform-go.com/api/v1/webhooks/status

# Use both headers
curl -H "Authorization: Bearer $JWT" \
  -H "X-API-Key: pk_abc123..." \
  https://api.platform-go.com/api/admin/audit
```

---

## Summary

This unified authentication strategy provides:

1. **Zero Header Conflicts** ✅
   - JWT: `Authorization: Bearer <token>`
   - API Key: `X-API-Key: <key>`
   - Completely separate, explicit intent

2. **Clear Separation of Concerns** ✅
   - Interactive users: JWT (stateless, expiring)
   - Services: API Keys (long-lived, trackable)
   - Admins: Both for multi-factor auth

3. **Flexible Middleware Stack** ✅
   - JWT-only for user features
   - API-Key-only for automation
   - Optional for flexible endpoints
   - Both for multi-factor scenarios

4. **Production Ready** ✅
   - Bcrypt hashing for API keys
   - Optional expiration
   - Comprehensive audit trail
   - Rate limiting capable

5. **Easy Integration** ✅
   - Clean separation prevents confusion
   - Clear header documentation
   - Standardized implementation
   - Easy client integration

**Key Principle**: One authentication strategy, multiple authentication methods, **zero conflicts, complete flexibility**.
