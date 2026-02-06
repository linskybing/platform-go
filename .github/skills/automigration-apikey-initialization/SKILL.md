---
name: automigration-apikey-initialization
description: Automatic database schema creation with GORM AutoMigration, API key-based authentication for all endpoints, default admin user initialization with secure password hashing, and complete security best practices for production systems.
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
---


# AutoMigration, API Key Authentication & Initialization

## Overview

This skill covers:
1. **GORM AutoMigration**: Automatic schema creation with foreign keys and constraints on first run
2. **API Key Authentication**: Mandatory API key-based access control for all endpoints (`Authorization: Bearer <key>`)
3. **Database Initialization**: Automatically seeds default admin user with generated API key and super group
4. **Security Best Practices**: Bcrypt hashing, key rotation, expiration, and comprehensive audit logging

**When to use this skill:**
- Setting up new platform-go instances from scratch
- Implementing API key-based authentication
- Ensuring database initialization follows security best practices
- Adding authentication middleware to routes
- Managing API keys (creation, rotation, revocation)


## Step 2: API Key Authentication Implementation

### User Model with API Key Fields

Add API key support to users:

```go
// internal/domain/user/model.go
package user

type User struct {
    UID         string    `gorm:"primaryKey;size:20"`
    Username    string    `gorm:"uniqueIndex;size:50"`
    Password    string    `gorm:"size:255"`
    Email       string    `gorm:"uniqueIndex;size:100"`
    FullName    string    `gorm:"size:50"`
    Type        string    `gorm:"type:user_type;default:'origin'"`
    Status      string    `gorm:"type:user_status;default:'offline'"`
    
    // API Key fields
    APIKey      string    `gorm:"uniqueIndex;size:64"` // Hashed API key
    APIKeyHash  string    `gorm:"size:128"`             // Bcrypt hash of API key
    APIKeyName  string    `gorm:"size:100"`             // Human-readable name
    APIKeyLastUsed *time.Time                           // Track usage
    APIKeyCreatedAt time.Time `gorm:"autoCreateTime"`    // When key was generated
    APIKeyExpiresAt *time.Time                          // Optional expiration
    
    CreatedAt   time.Time `gorm:"autoCreateTime"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime"`
    DeletedAt   gorm.DeletedAt `gorm:"index"`
}
```

### API Key Generation and Secure Hashing

```go
// pkg/utils/apikey.go
package utils

import (
    "crypto/rand"
    "encoding/base64"
    "golang.org/x/crypto/bcrypt"
)

const (
    APIKeyLength = 32 // 32 bytes = 43 chars when base64 encoded
    APIKeyPrefix = "pk_" // Prefix for identifying API keys
)

// GenerateAPIKey creates a new random API key
func GenerateAPIKey() (string, error) {
    b := make([]byte, APIKeyLength)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    
    key := APIKeyPrefix + base64.URLEncoding.EncodeToString(b)
    return key, nil
}

// HashAPIKey hashes an API key using bcrypt
func HashAPIKey(key string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
    return string(hash), err
}

// VerifyAPIKey checks if a provided key matches the hash
func VerifyAPIKey(providedKey, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(providedKey))
    return err == nil
}
```

### API Key Authentication Middleware (Using X-API-Key Header)

```go
// internal/api/middleware/apikey.go
package middleware

import (
    "errors"
    "net/http"
    "strings"
    "time"
    
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "platform-go/internal/domain/user"
    "platform-go/pkg/response"
)

// APIKeyAuth middleware validates API key from X-API-Key header
// Header format: X-API-Key: pk_<random-string>
// Optional: X-API-Key-Name: <friendly-name>
func APIKeyAuth(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Extract API key from X-API-Key header
        apiKey := c.GetHeader("X-API-Key")
        if apiKey == "" {
            response.Error(c, http.StatusUnauthorized, "Missing X-API-Key header")
            c.Abort()
            return
        }
        
        // Validate API key format (must start with pk_)
        if !strings.HasPrefix(apiKey, "pk_") {
            response.Error(c, http.StatusUnauthorized, "Invalid API key format")
            c.Abort()
            return
        }
        
        // Look up user by API key hash
        var u user.User
        if err := db.Where("api_key_hash = ?", apiKey).First(&u).Error; err != nil {
            if errors.Is(err, gorm.ErrRecordNotFound) {
                response.Error(c, http.StatusUnauthorized, "Invalid API key")
            } else {
                response.Error(c, http.StatusInternalServerError, "Database error")
            }
            c.Abort()
            return
        }
        
        // Verify API key is not expired
        if u.APIKeyExpiresAt != nil && u.APIKeyExpiresAt.Before(time.Now()) {
            response.Error(c, http.StatusUnauthorized, "API key has expired")
            c.Abort()
            return
        }
        
        // Update last used timestamp for audit trail
        db.Model(&u).Update("api_key_last_used", time.Now())
        
        // Store user in context
        c.Set("user", u)
        c.Set("user_id", u.UID)
        c.Set("auth_type", "api_key")
        c.Set("api_key_name", c.GetHeader("X-API-Key-Name"))
        
        c.Next()
    }
}

// GetUserFromContext retrieves authenticated user from context
func GetUserFromContext(c *gin.Context) (*user.User, error) {
    val, exists := c.Get("user")
    if !exists {
        return nil, errors.New("user not found in context")
    }
    
    u, ok := val.(user.User)
    if !ok {
        return nil, errors.New("invalid user type in context")
    }
    
    return &u, nil
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(c *gin.Context) (string, error) {
    val, exists := c.Get("user_id")
    if !exists {
        return "", errors.New("user_id not found in context")
    }
    
    uid, ok := val.(string)
    if !ok {
        return "", errors.New("invalid user_id type in context")
    }
    
    return uid, nil
}
```

### Complete API Key Management Service

```go
// internal/application/user/apikey_service.go
package user_service

import (
    "time"
    "gorm.io/gorm"
    "platform-go/internal/domain/user"
    "platform-go/pkg/utils"
)

type APIKeyService struct {
    db *gorm.DB
}

func NewAPIKeyService(db *gorm.DB) *APIKeyService {
    return &APIKeyService{db: db}
}

// CreateAPIKey generates a new API key for a user
func (s *APIKeyService) CreateAPIKey(uid string, keyName string) (string, error) {
    // Generate new API key
    apiKey, err := utils.GenerateAPIKey()
    if err != nil {
        return "", err
    }
    
    // Hash the API key
    hash, err := utils.HashAPIKey(apiKey)
    if err != nil {
        return "", err
    }
    
    // Store in database
    now := time.Now()
    if err := s.db.Model(&user.User{}).
        Where("u_id = ?", uid).
        Updates(map[string]interface{}{
            "api_key_hash": hash,
            "api_key_name": keyName,
            "api_key_created_at": now,
        }).Error; err != nil {
        return "", err
    }
    
    // Return unhashed key to user (only time it's visible)
    return apiKey, nil
}

// RevokeAPIKey removes an API key
func (s *APIKeyService) RevokeAPIKey(uid string) error {
    return s.db.Model(&user.User{}).
        Where("u_id = ?", uid).
        Updates(map[string]interface{}{
            "api_key_hash": nil,
            "api_key_name": nil,
        }).Error
}

// RotateAPIKey generates a new key and revokes the old one
func (s *APIKeyService) RotateAPIKey(uid string, keyName string) (string, error) {
    // Create new key (this will overwrite old one)
    return s.CreateAPIKey(uid, keyName)
}

// SetAPIKeyExpiration sets expiration date for an API key
func (s *APIKeyService) SetAPIKeyExpiration(uid string, expiresAt time.Time) error {
    return s.db.Model(&user.User{}).
        Where("u_id = ?", uid).
        Update("api_key_expires_at", expiresAt).Error
}

// GetAPIKeyInfo retrieves API key metadata (not the key itself)
func (s *APIKeyService) GetAPIKeyInfo(uid string) (*APIKeyInfo, error) {
    var u user.User
    if err := s.db.Where("u_id = ?", uid).First(&u).Error; err != nil {
        return nil, err
    }
    
    return &APIKeyInfo{
        KeyName: u.APIKeyName,
        CreatedAt: u.APIKeyCreatedAt,
        LastUsed: u.APIKeyLastUsed,
        ExpiresAt: u.APIKeyExpiresAt,
    }, nil
}

type APIKeyInfo struct {
    KeyName string
    CreatedAt time.Time
    LastUsed *time.Time
    ExpiresAt *time.Time
}
```


## Step 4: API Routes with Authentication

### Route Registration with API Key Middleware

```go
// internal/api/routes/routes.go
package routes

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "platform-go/internal/api/handlers"
    "platform-go/internal/api/middleware"
)

func RegisterRoutes(router *gin.Engine, db *gorm.DB) {
    // Health check (no auth required)
    router.GET("/health", handlers.Health)
    
    // Authentication routes (no auth required)
    authRoutes := router.Group("/api/auth")
    {
        authRoutes.POST("/login", handlers.Login)
        authRoutes.POST("/register", handlers.Register)
    }
    
    // Protected API routes (require API key)
    protectedRoutes := router.Group("/api")
    protectedRoutes.Use(middleware.APIKeyAuth(db))
    {
        // User routes
        userRoutes := protectedRoutes.Group("/users")
        {
            userRoutes.GET("/:id", handlers.GetUser)
            userRoutes.PUT("/:id", handlers.UpdateUser)
            userRoutes.POST("/:id/apikey", handlers.CreateAPIKey)
            userRoutes.DELETE("/:id/apikey", handlers.RevokeAPIKey)
        }
        
        // Group routes
        groupRoutes := protectedRoutes.Group("/groups")
        {
            groupRoutes.GET("", handlers.ListGroups)
            groupRoutes.GET("/:id", handlers.GetGroup)
            groupRoutes.POST("", handlers.CreateGroup)
            groupRoutes.PUT("/:id", handlers.UpdateGroup)
            groupRoutes.DELETE("/:id", handlers.DeleteGroup)
        }
        
        // Project routes
        projectRoutes := protectedRoutes.Group("/projects")
        {
            projectRoutes.GET("", handlers.ListProjects)
            projectRoutes.GET("/:id", handlers.GetProject)
            projectRoutes.POST("", handlers.CreateProject)
            projectRoutes.PUT("/:id", handlers.UpdateProject)
            projectRoutes.DELETE("/:id", handlers.DeleteProject)
        }
        
        // ... Add more routes as needed
    }
}
```

### Login Handler with API Key Response

```go
// internal/api/handlers/auth.go
package handlers

import (
    "net/http"
    "golang.org/x/crypto/bcrypt"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "platform-go/internal/domain/user"
    "platform-go/internal/application/user_service"
    "platform-go/pkg/response"
    "platform-go/pkg/utils"
)

type LoginRequest struct {
    Username string `json:"username" binding:"required"`
    Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
    UserID  string `json:"user_id"`
    Username string `json:"username"`
    APIKey  string `json:"api_key"` // Only shown at login
    Message string `json:"message"`
}

// Login authenticates user and returns API key
func Login(c *gin.Context, db *gorm.DB) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid request")
        return
    }
    
    // Find user by username
    var u user.User
    if err := db.Where("username = ?", req.Username).First(&u).Error; err != nil {
        response.Error(c, http.StatusUnauthorized, "Invalid credentials")
        return
    }
    
    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)); err != nil {
        response.Error(c, http.StatusUnauthorized, "Invalid credentials")
        return
    }
    
    // Generate new API key
    apiKeyService := user_service.NewAPIKeyService(db)
    apiKey, err := apiKeyService.CreateAPIKey(u.UID, "Login "+time.Now().Format("2006-01-02 15:04:05"))
    if err != nil {
        response.Error(c, http.StatusInternalServerError, "Failed to generate API key")
        return
    }
    
    response.Success(c, http.StatusOK, LoginResponse{
        UserID:   u.UID,
        Username: u.Username,
        APIKey:   apiKey,
        Message:  "Login successful. Save your API key securely.",
    })
}
```


## Step 6: Implementation and Testing Checklist

### Database Setup
- [x] Update User model with API key fields
- [x] Create AutoMigration function
- [x] Create foreign key constraint function
- [x] Create index optimization function
- [x] Create default data initialization function

### API Key Management
- [x] Implement API key generation
- [x] Implement API key hashing
- [x] Implement API key verification
- [x] Create APIKeyService
- [x] Create APIKeyAuth middleware

### Integration
- [x] Update routes with middleware
- [x] Implement login handler
- [x] Implement API key creation endpoint
- [x] Implement API key revocation endpoint
- [x] Add audit logging for API key operations

### Security
- [x] Use bcrypt for password hashing
- [x] Use bcrypt for API key hashing
- [x] Never store plain API keys
- [x] Log API key operations
- [x] Implement key expiration

### Testing
- [ ] Test AutoMigration with fresh database
- [ ] Test API key generation and verification
- [ ] Test cascade delete behavior
- [ ] Test API key middleware on protected routes
- [ ] Test concurrent API key creation
- [ ] Test API key expiration logic
