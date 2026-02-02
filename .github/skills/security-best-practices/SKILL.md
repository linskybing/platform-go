---
name: security-best-practices
description: Security guidelines, authentication, authorization, input validation, and secure coding practices for platform-go
---

# Security Best Practices

This skill provides comprehensive security guidelines for the platform-go project.

## When to Use

Apply this skill when:
- Implementing authentication or authorization
- Handling user input or API requests
- Working with sensitive data (passwords, tokens, secrets)
- Accessing files or external resources
- Implementing access control
- Storing or transmitting credentials
- Validating user permissions
- Creating new API endpoints

## Authentication & Authorization

### 1. JWT Token Management

```go
// Use secure JWT implementation
type TokenClaims struct {
    UserID uint   `json:"user_id"`
    Role   string `json:"role"`
    jwt.StandardClaims
}

func GenerateToken(userID uint, role string, secret string) (string, error) {
    claims := TokenClaims{
        UserID: userID,
        Role:   role,
        StandardClaims: jwt.StandardClaims{
            ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
            IssuedAt:  time.Now().Unix(),
            NotBefore: time.Now().Unix(),
        },
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString([]byte(secret))
    if err != nil {
        return "", fmt.Errorf("failed to sign token: %w", err)
    }
    
    return tokenString, nil
}

func ValidateToken(tokenString string, secret string) (*TokenClaims, error) {
    claims := &TokenClaims{}
    token, err := jwt.ParseWithClaims(tokenString, claims,
        func(token *jwt.Token) (interface{}, error) {
            // Verify signing method
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(secret), nil
        })
    
    if err != nil {
        return nil, fmt.Errorf("failed to parse token: %w", err)
    }
    
    if !token.Valid {
        return nil, fmt.Errorf("invalid token")
    }
    
    return claims, nil
}
```

### 2. Password Security

```go
// Always hash passwords with bcrypt
func HashPassword(password string) (string, error) {
    // Cost of 12 is production-grade
    hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    return string(hash), nil
}

func VerifyPassword(hashedPassword, password string) error {
    return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// Validate password strength
func ValidatePasswordStrength(password string) error {
    if len(password) < 8 {
        return fmt.Errorf("password must be at least 8 characters")
    }
    
    hasUppercase := regexp.MustCompile("[A-Z]").MatchString(password)
    hasLowercase := regexp.MustCompile("[a-z]").MatchString(password)
    hasDigit := regexp.MustCompile("[0-9]").MatchString(password)
    hasSpecial := regexp.MustCompile("[!@#$%^&*()_+\\-=\\[\\]{};':\"\\\\|,.<>/?]").MatchString(password)
    
    if !hasUppercase || !hasLowercase || !hasDigit || !hasSpecial {
        return fmt.Errorf("password must contain uppercase, lowercase, digit, and special character")
    }
    
    return nil
}
```

### 3. RBAC Implementation

```go
// Define role-based permissions
const (
    RoleAdmin  = "admin"
    RoleUser   = "user"
    RoleGuest  = "guest"
)

type Permission int

const (
    PermRead    Permission = 1
    PermWrite   Permission = 2
    PermDelete  Permission = 4
    PermManage  Permission = 8
)

func HasPermission(userRole string, permission Permission) bool {
    rolePermissions := map[string][]Permission{
        RoleAdmin: {PermRead, PermWrite, PermDelete, PermManage},
        RoleUser:  {PermRead, PermWrite},
        RoleGuest: {PermRead},
    }
    
    for _, p := range rolePermissions[userRole] {
        if p == permission {
            return true
        }
    }
    return false
}

// Check authorization in handlers
func RequirePermission(permission Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
            c.Abort()
            return
        }
        
        if !HasPermission(role.(string), permission) {
            c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

## Input Validation & Sanitization

### 4. Input Validation

```go
// Validate all user inputs at API boundary
type UserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=32,alphanum"`
    Email    string `json:"email" binding:"required,email"`
    Age      int    `json:"age" binding:"omitempty,gt=0,lt=150"`
}

// Custom validation
func ValidateProjectName(name string) error {
    if len(name) < 1 || len(name) > 100 {
        return fmt.Errorf("project name must be 1-100 characters")
    }
    
    // Only allow alphanumeric, hyphen, underscore
    if !regexp.MustCompile("^[a-zA-Z0-9-_]+$").MatchString(name) {
        return fmt.Errorf("project name contains invalid characters")
    }
    
    return nil
}

// Sanitize file paths
func SafizeFilePath(basePath, userPath string) (string, error) {
    // Prevent directory traversal
    fullPath := filepath.Join(basePath, userPath)
    absPath, err := filepath.Abs(fullPath)
    if err != nil {
        return "", fmt.Errorf("invalid path: %w", err)
    }
    
    absBaPath, err := filepath.Abs(basePath)
    if err != nil {
        return "", fmt.Errorf("invalid base path: %w", err)
    }
    
    // Ensure path is within base directory
    if !strings.HasPrefix(absPath, absBasePath) {
        return "", fmt.Errorf("path traversal attempt detected")
    }
    
    return absPath, nil
}
```

### 5. SQL Injection Prevention

```go
// Always use parameterized queries (GORM does this by default)
// GOOD: Uses parameterized query
user, err := userRepo.GetByUsername(ctx, username)

// GORM: Always uses prepared statements
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*User, error) {
    var user User
    if err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error; err != nil {
        return nil, fmt.Errorf("user not found: %w", err)
    }
    return &user, nil
}

// BAD: String concatenation (FORBIDDEN)
// DO NOT DO THIS:
query := "SELECT * FROM users WHERE username = '" + username + "'"
```

## Secrets & Credentials Management

### 6. Environment Variables & Secrets

```go
// Load secrets from environment, not config files
type Config struct {
    DatabaseURL  string `envconfig:"DATABASE_URL" required:"true"`
    JWTSecret    string `envconfig:"JWT_SECRET" required:"true"`
    MinioKey     string `envconfig:"MINIO_KEY" required:"true"`
    MinioSecret  string `envconfig:"MINIO_SECRET" required:"true"`
}

func LoadConfig() (*Config, error) {
    var cfg Config
    if err := envconfig.Process("APP", &cfg); err != nil {
        return nil, fmt.Errorf("failed to load config: %w", err)
    }
    return &cfg, nil
}

// Never log secrets
func (s *Service) LogOperation(operation string, result interface{}) {
    // GOOD: Log sanitized information
    log.Printf("Operation: %s, Result: success", operation)
    
    // BAD: NEVER log secrets (FORBIDDEN)
    // log.Printf("Using secret: %s", s.jwtSecret)
}

// Use context to pass secrets
func (h *Handler) HandleRequest(c *gin.Context) {
    // GOOD: Get secret from secure context
    secret := c.Request.Context().Value("jwt_secret").(string)
    
    // Use secret securely
    token, err := GenerateToken(userID, role, secret)
}
```

### 7. File Upload Security

```go
// Validate file uploads
func ValidateFileUpload(file *multipart.FileHeader, maxSize int64, allowedTypes []string) error {
    // Check file size
    if file.Size > maxSize {
        return fmt.Errorf("file size exceeds maximum: %d bytes", maxSize)
    }
    
    // Verify MIME type
    mimeType := file.Header.Get("Content-Type")
    allowed := false
    for _, t := range allowedTypes {
        if mimeType == t {
            allowed = true
            break
        }
    }
    if !allowed {
        return fmt.Errorf("file type not allowed: %s", mimeType)
    }
    
    // Sanitize filename
    filename := filepath.Base(file.Filename)
    filename = strings.TrimSpace(filename)
    if filename == "" {
        return fmt.Errorf("invalid filename")
    }
    
    // Prevent directory traversal
    if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
        return fmt.Errorf("invalid filename contains path separators")
    }
    
    return nil
}

// Save file with secure naming
func SaveUploadedFile(file *multipart.FileHeader, uploadDir string) (string, error) {
    // Generate unique filename
    uuid := uuid.New().String()
    ext := filepath.Ext(file.Filename)
    savedName := fmt.Sprintf("%s%s", uuid, ext)
    
    savePath := filepath.Join(uploadDir, savedName)
    
    // Verify path is safe
    absPath, err := filepath.Abs(savePath)
    if err != nil {
        return "", fmt.Errorf("invalid path: %w", err)
    }
    
    absUploadDir, err := filepath.Abs(uploadDir)
    if err != nil {
        return "", fmt.Errorf("invalid upload directory: %w", err)
    }
    
    if !strings.HasPrefix(absPath, absUploadDir) {
        return "", fmt.Errorf("path traversal attempt detected")
    }
    
    if err := ioutil.WriteFile(absPath, fileData, 0644); err != nil {
        return "", fmt.Errorf("failed to save file: %w", err)
    }
    
    return savedName, nil
}
```

## Security Headers & Configuration

### 8. API Security

```go
// Add security headers middleware
func SecurityHeadersMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Content-Security-Policy", "default-src 'self'")
        
        c.Next()
    }
}

// CORS configuration
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", os.Getenv("CORS_ALLOWED_ORIGINS"))
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}

// Rate limiting
func RateLimitMiddleware(limit rate.Limit, burst int) gin.HandlerFunc {
    limiter := rate.NewLimiter(limit, burst)
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

## Security Checklist

- [ ] All passwords are hashed with bcrypt (cost >= 12)
- [ ] All database queries use parameterized statements (no string concatenation)
- [ ] Input validation is performed on all user inputs
- [ ] File paths are validated to prevent directory traversal
- [ ] Sensitive data (secrets, tokens) are never logged
- [ ] JWT tokens have expiration times
- [ ] HTTPS is enforced in production
- [ ] CORS is properly configured
- [ ] Authentication middleware is applied to protected routes
- [ ] Role-based access control is implemented
- [ ] Security headers are set on all responses
- [ ] API rate limiting is enabled
- [ ] File uploads are validated (size, type, content)
- [ ] Secrets are loaded from environment variables, not config files
- [ ] SQL injections are prevented through parameterized queries

## Security Anti-Patterns (FORBIDDEN)

- Never use md5 or sha1 for passwords
- Never log user passwords or tokens
- Never use rand for security-critical operations
- Never skip TLS certificate validation
- Never hardcode secrets in code
- Never trust user input without validation
- Never allow SQL injection through string concatenation
- Never use weak encryption algorithms (avoid DES, RC4)
- Never return sensitive error details to clients
- Never store passwords in plaintext

## Security Tools & Commands

```bash
# Check for security vulnerabilities
go list -json -m all | nancy sleuth

# Check for hardcoded secrets
truffleHog filesystem --json

# Run security linter
gosec ./...

# Check OWASP vulnerabilities
go test -run TestSecurity

# TLS certificate generation (development only)
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 365 -nodes
```
