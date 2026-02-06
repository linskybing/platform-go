---
name: security-compliance
description: Security best practices, access control, authentication, API key management, and compliance for platform-go
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
  consolidated_from:
    - security-best-practices
    - access-control-best-practices
    - unified-authentication-strategy
    - automigration-apikey-initialization
---

# Security & Compliance Excellence

Comprehensive guidelines for authentication, authorization, access control, secure coding, and API key management.

## Authentication Strategy

### Unified Authentication Architecture
```
┌─────────────────────────────────────┐
│        Incoming Request             │
└────────────┬────────────────────────┘
             │
    ┌────────┴────────┐
    │                 │
    ▼                 ▼
JWT Token      API Key
(Bearer)      (X-API-Key)
    │                 │
    └────────┬────────┘
             │
    ┌────────▼────────┐
    │ Validate Token  │
    │ Extract Claims  │
    └────────┬────────┘
             │
    ┌────────▼────────────────────────┐
    │ Attach User to Context          │
    │ ctx = context.WithValue(...)    │
    └────────┬────────────────────────┘
             │
    ┌────────▼────────────────────────┐
    │ Pass to Handler                 │
    └────────────────────────────────┘
```

### JWT Token Handler (Web Clients)
```go
// Extract JWT from Authorization header
func extractToken(r *http.Request) (string, error) {
    authHeader := r.Header.Get("Authorization")
    if authHeader == "" {
        return "", ErrNoToken
    }
    
    parts := strings.Split(authHeader, " ")
    if len(parts) != 2 || parts[0] != "Bearer" {
        return "", ErrInvalidToken
    }
    
    return parts[1], nil
}

// Validate and extract claims
func ValidateToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected method: %v", token.Header["alg"])
        }
        return jwtSecret, nil
    })
    
    if err != nil {
        return nil, fmt.Errorf("parse token failed: %w", err)
    }
    
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, ErrInvalidToken
    }
    
    if claims.ExpiresAt.Before(time.Now()) {
        return nil, ErrTokenExpired
    }
    
    return claims, nil
}
```

### API Key Handler (Services/Automation)
```go
// Extract API key from X-API-Key header
func extractAPIKey(r *http.Request) (string, error) {
    apiKey := r.Header.Get("X-API-Key")
    if apiKey == "" {
        return "", ErrNoAPIKey
    }
    return apiKey, nil
}

// Validate API key and retrieve associated user
func ValidateAPIKey(ctx context.Context, apiKey string) (*User, error) {
    // Hash the provided key
    hashedKey := bcrypt.HashKey(apiKey)
    
    // Lookup in database
    user, err := db.GetUserByAPIKey(ctx, hashedKey)
    if err != nil {
        return nil, fmt.Errorf("invalid API key: %w", err)
    }
    
    // Check if key is still valid
    if user.APIKeyExpiredAt.Before(time.Now()) {
        return nil, ErrAPIKeyExpired
    }
    
    return user, nil
}
```

### Middleware Routing
```go
// Authentication middleware - detects JWT vs API Key
func AuthMiddleware(c *gin.Context) {
    // Try JWT first (web clients)
    token, _ := extractToken(c.Request)
    if token != "" {
        claims, err := ValidateToken(token)
        if err == nil {
            c.Set("user_id", claims.UserID)
            c.Set("user_type", "web")
            c.Next()
            return
        }
    }
    
    // Try API Key (services/automation)
    apiKey, _ := extractAPIKey(c.Request)
    if apiKey != "" {
        user, err := ValidateAPIKey(c.Request.Context(), apiKey)
        if err == nil {
            c.Set("user_id", user.ID)
            c.Set("user_type", "service")
            c.Next()
            return
        }
    }
    
    c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
    c.Abort()
}

// Route configuration
router.Use(AuthMiddleware)

// Web client routes
web := router.Group("/api/web")
web.POST("/projects", handlers.CreateProject)

// Service routes
service := router.Group("/api/service")
service.POST("/jobs", handlers.SubmitJob)
```

## Access Control & Authorization

### Role-Based Access Control (RBAC)
```
User → Role → Permissions
  ↓
admin     → [create_user, delete_user, manage_roles, ...]
user      → [view_profile, update_profile, ...]
guest     → [view_public_data, ...]
```

### Role Definition
```go
type Role string

const (
    Admin    Role = "admin"
    Manager  Role = "manager"
    User     Role = "user"
    Guest    Role = "guest"
)

type Permission string

const (
    CreateUser     Permission = "create_user"
    DeleteUser     Permission = "delete_user"
    ManageRoles    Permission = "manage_roles"
    ViewProject    Permission = "view_project"
    EditProject    Permission = "edit_project"
    DeleteProject  Permission = "delete_project"
)

var rolePermissions = map[Role][]Permission{
    Admin: {CreateUser, DeleteUser, ManageRoles, ViewProject, EditProject, DeleteProject},
    Manager: {ViewProject, EditProject},
    User: {ViewProject},
    Guest: {},
}
```

### Authorization Middleware
```go
func RequirePermission(requiredPerm Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID, _ := c.Get("user_id")
        
        // Get user role from database
        user, err := db.GetUser(c.Request.Context(), userID.(int))
        if err != nil {
            c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
            c.Abort()
            return
        }
        
        // Check if user has permission
        permissions := rolePermissions[user.Role]
        hasPermission := false
        for _, p := range permissions {
            if p == requiredPerm {
                hasPermission = true
                break
            }
        }
        
        if !hasPermission {
            c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Usage
router.DELETE("/api/users/:id", RequirePermission(DeleteUser), handlers.DeleteUser)
```

### Resource-Level Authorization
```go
func RequireResourceAccess(getUserID func(*gin.Context) int, getResourceOwner func(context.Context, int) (int, error)) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := getUserID(c)
        resourceID := c.Param("id")
        
        ownerID, err := getResourceOwner(c.Request.Context(), resourceID)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "resource not found"})
            c.Abort()
            return
        }
        
        if userID != ownerID {
            c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Usage
router.GET("/api/projects/:id", 
    RequireResourceAccess(
        func(c *gin.Context) int { uid, _ := c.Get("user_id"); return uid.(int) },
        func(ctx context.Context, projectID int) (int, error) {
            project, err := db.GetProject(ctx, projectID)
            return project.OwnerID, err
        },
    ),
    handlers.GetProject)
```

## API Key Management

### Key Generation & Storage
```go
// Generate secure API key
func GenerateAPIKey() string {
    return generateRandomString(32) // 256-bit entropy
}

// Hash before storage (never store plain keys)
func HashAPIKey(apiKey string) string {
    hash := sha256.Sum256([]byte(apiKey))
    return hex.EncodeToString(hash[:])
}

// Database schema
type APIKey struct {
    ID        int       `gorm:"primaryKey"`
    UserID    int       `gorm:"index"`
    KeyHash   string    `gorm:"uniqueIndex"`
    Name      string
    ExpiresAt time.Time
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Create API key
func CreateAPIKey(ctx context.Context, userID int, name string, expiresIn time.Duration) (string, error) {
    plainKey := GenerateAPIKey()
    hashedKey := HashAPIKey(plainKey)
    
    err := db.Create(&APIKey{
        UserID:    userID,
        KeyHash:   hashedKey,
        Name:      name,
        ExpiresAt: time.Now().Add(expiresIn),
    }).Error
    
    if err != nil {
        return "", fmt.Errorf("create API key failed: %w", err)
    }
    
    // Only return plain key once
    return plainKey, nil
}
```

## Secure Coding Practices

### Input Validation
```go
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required,min=1,max=100"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
}

func CreateUser(c *gin.Context) {
    var req CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    // Request already validated by binding tags
}
```

### Password Hashing
```go
import "golang.org/x/crypto/bcrypt"

const bcryptCost = 12

// Hash password before storage
func HashPassword(password string) (string, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
    return string(hash), err
}

// Compare password with hash
func VerifyPassword(hash, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
```

### SQL Injection Prevention
```go
// Use parameterized queries (GORM does this automatically)
var users []User
db.Where("email = ?", userEmail).Find(&users) // Safe ✓

// Not recommended
db.Where(fmt.Sprintf("email = '%s'", userEmail)).Find(&users) // Vulnerable ✗
```

### Environment Variables
```go
// Use environment variables for secrets
databaseURL := os.Getenv("DATABASE_URL")
if databaseURL == "" {
    log.Fatal("DATABASE_URL not set")
}

// Never log sensitive data
log.Info("Connecting to database") // ✓
log.Info("URL: " + databaseURL)    // ✗
```

## Default Admin Initialization

### Auto-Migration with Security
```go
func InitializeDatabase(db *gorm.DB) error {
    // Auto-migrate schema
    if err := db.AutoMigrate(
        &User{},
        &APIKey{},
        &Project{},
    ).Error; err != nil {
        return fmt.Errorf("auto migration failed: %w", err)
    }
    
    // Create default admin user if not exists
    adminExists := false
    db.Where("role = ?", Admin).Limit(1).Select("id").Row().Scan(&adminExists)
    
    if !adminExists {
        adminPassword := os.Getenv("ADMIN_PASSWORD")
        if adminPassword == "" {
            return errors.New("ADMIN_PASSWORD environment variable required")
        }
        
        hashedPassword, _ := HashPassword(adminPassword)
        
        admin := &User{
            Name:     "Admin",
            Email:    "admin@example.com",
            Password: hashedPassword,
            Role:     Admin,
        }
        
        if err := db.Create(admin).Error; err != nil {
            return fmt.Errorf("create admin user failed: %w", err)
        }
        
        log.Info("Default admin user created")
    }
    
    return nil
}
```

## Security Validation Checklist

- [ ] All passwords hashed with bcrypt
- [ ] API keys hashed before storage
- [ ] No sensitive data in logs
- [ ] HTTPS enforced in production
- [ ] Rate limiting configured
- [ ] Input validation on all endpoints
- [ ] CORS properly restricted
- [ ] SQL injection prevention (parameterized queries)
- [ ] CSRF tokens for state-changing operations
- [ ] Security headers set (Content-Security-Policy, X-Frame-Options, etc.)
- [ ] Regular security audits scheduled
- [ ] Incident response plan documented

## Tools & Scripts

### Security Checks
```bash
# Scan for vulnerabilities
bash .github/skills-consolidated/security-compliance/scripts/security-scan.sh

# Check for secrets in code
bash .github/skills-consolidated/security-compliance/scripts/check-secrets.sh

# Audit API keys
bash .github/skills-consolidated/security-compliance/scripts/audit-keys.sh
```

## References
- OWASP Top 10: https://owasp.org/www-project-top-ten/
- JWT Best Practices: https://tools.ietf.org/html/rfc8725
- Go Security: https://cheatsheetseries.owasp.org/cheatsheets/Go_Cheat_Sheet.html
- Bcrypt Information: https://en.wikipedia.org/wiki/Bcrypt
