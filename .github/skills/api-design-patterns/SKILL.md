---
name: api-design-patterns
description: RESTful API design patterns, Gin framework best practices, error handling, validation, and response formatting for platform-go
---

# API Design Patterns

This skill provides guidelines for designing and implementing RESTful APIs using Gin framework.

## When to Use

Apply this skill when:
- Creating new HTTP endpoints or handlers
- Designing request/response structures
- Implementing input validation for API requests
- Setting up middleware (authentication, logging, rate limiting)
- Handling errors in API responses
- Managing file uploads or streaming responses
- Implementing pagination, filtering, or sorting
- Adding new routes or versioning APIs

## Quick Start: Using API Scripts

This skill includes ready-to-use Swagger/OpenAPI scripts:

```bash
# Generate Swagger documentation from code comments
bash .github/skills/api-design-patterns/scripts/swagger-generate.sh

# Validate Swagger specification
bash .github/skills/api-design-patterns/scripts/swagger-validate.sh
```

## API Design Principles

### 1. Swagger/OpenAPI Documentation

All API endpoints must have complete Swagger documentation following OpenAPI 3.0 standard.

```go
// Install swag for Go
// go get -u github.com/swaggo/swag/cmd/swag
// go get -u github.com/swaggo/gin-swagger
// go get -u github.com/swaggo/files

// Generate Swagger docs
// swag init -g cmd/api/main.go

// main.go
package main

import (
    "github.com/gin-gonic/gin"
    swaggerfiles "github.com/swaggo/files"
    ginswagger "github.com/swaggo/gin-swagger"
    _ "github.com/linskybing/platform-go/docs" // Generated docs
)

// @title Platform-Go API
// @version 1.0
// @description REST API for platform-go project
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@example.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host api.example.com
// @basePath /api/v1
// @schemes https

func main() {
    r := gin.Default()
    
    // Swagger endpoint
    r.GET("/swagger/*any", ginswagger.WrapHandler(swaggerfiles.Handler))
    
    r.Run()
}
```

### 2. Handler Documentation with Swagger Annotations

Every handler must have complete Swagger annotations:

```go
// internal/api/handler/user_handler.go
package handler

import (
    "github.com/gin-gonic/gin"
    swaggerfiles "github.com/swaggo/files"
)

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with username and email
// @Tags users
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "Create user request"
// @Success 201 {object} UserResponse
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 409 {object} ErrorResponse "User already exists"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /users [post]
// @Security ApiKeyAuth
func (h *UserHandler) Create(c *gin.Context) {
    var req dto.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
        return
    }
    
    user, err := h.userService.CreateUser(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, dto.ToUserResponse(user))
}

// GetUser godoc
// @Summary Get user by ID
// @Description Retrieve a user by their ID
// @Tags users
// @Produce json
// @Param id path uint true "User ID"
// @Success 200 {object} UserResponse
// @Failure 404 {object} ErrorResponse "User not found"
// @Router /users/{id} [get]
// @Security ApiKeyAuth
func (h *UserHandler) Get(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    
    user, err := h.userService.GetUser(c.Request.Context(), uint(id))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
        return
    }
    
    c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

// ListUsers godoc
// @Summary List all users
// @Description List users with pagination and filtering
// @Tags users
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(20)
// @Param role query string false "Filter by role"
// @Param status query string false "Filter by status"
// @Success 200 {object} ListUsersResponse
// @Failure 500 {object} ErrorResponse
// @Router /users [get]
// @Security ApiKeyAuth
func (h *UserHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    
    users, total, err := h.userService.ListUsers(c.Request.Context(), &ListFilter{
        Page:  page,
        Limit: limit,
        Role:  c.Query("role"),
    })
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "data": users,
        "pagination": gin.H{
            "page":  page,
            "limit": limit,
            "total": total,
        },
    })
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user information
// @Tags users
// @Accept json
// @Produce json
// @Param id path uint true "User ID"
// @Param request body UpdateUserRequest true "Update request"
// @Success 200 {object} UserResponse
// @Failure 400 {object} ErrorResponse "Invalid input"
// @Failure 404 {object} ErrorResponse "User not found"
// @Router /users/{id} [put]
// @Security ApiKeyAuth
func (h *UserHandler) Update(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    
    var req dto.UpdateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
        return
    }
    
    user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, dto.ToUserResponse(user))
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete a user by ID
// @Tags users
// @Produce json
// @Param id path uint true "User ID"
// @Success 200 {object} SuccessResponse
// @Failure 404 {object} ErrorResponse "User not found"
// @Router /users/{id} [delete]
// @Security ApiKeyAuth
func (h *UserHandler) Delete(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
        return
    }
    
    if err := h.userService.DeleteUser(c.Request.Context(), uint(id)); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}
```

### 3. DTO Swagger Model Tags

```go
// internal/api/handler/dto/request.go

// CreateUserRequest represents user creation request
// @Description User creation request with validation
type CreateUserRequest struct {
    // User username (3-32 alphanumeric characters)
    // Required: true
    // Example: john_doe
    Username string `json:"username" binding:"required,min=3,max=32,alphanum" example:"john_doe"`
    
    // User email address
    // Required: true
    // Format: email
    // Example: john@example.com
    Email string `json:"email" binding:"required,email" example:"john@example.com"`
    
    // User password (minimum 8 characters)
    // Required: true
    // MinLength: 8
    // Example: SecurePass123!
    Password string `json:"password" binding:"required,min=8" example:"SecurePass123!"`
    
    // User role (admin, manager, user)
    // Optional
    // Enum: [admin, manager, user]
    // Default: user
    Role string `json:"role" binding:"omitempty,oneof=admin manager user" example:"user"`
}

// UpdateUserRequest represents user update request
type UpdateUserRequest struct {
    // Updated username
    // Optional
    // Example: john_doe_updated
    Username string `json:"username,omitempty" example:"john_doe_updated"`
    
    // Updated email
    // Optional
    // Format: email
    Email string `json:"email,omitempty" example:"newemail@example.com"`
    
    // Updated password
    // Optional
    Password string `json:"password,omitempty"`
    
    // Updated role
    // Optional
    // Enum: [admin, manager, user]
    Role string `json:"role,omitempty" example:"manager"`
}

// internal/api/handler/dto/response.go

// UserResponse represents user data in responses
// @Description User response model with all public fields
type UserResponse struct {
    // User unique identifier
    // Example: 123
    ID uint `json:"id" example:"123"`
    
    // User username
    // Example: john_doe
    Username string `json:"username" example:"john_doe"`
    
    // User email address
    // Example: john@example.com
    Email string `json:"email" example:"john@example.com"`
    
    // User role
    // Example: user
    Role string `json:"role" example:"user"`
    
    // User creation timestamp
    // Format: date-time
    // Example: 2026-02-02T10:00:00Z
    CreatedAt time.Time `json:"created_at" example:"2026-02-02T10:00:00Z"`
    
    // User last update timestamp
    // Format: date-time
    // Example: 2026-02-02T10:00:00Z
    UpdatedAt time.Time `json:"updated_at" example:"2026-02-02T10:00:00Z"`
}

// ErrorResponse represents error response
// @Description Standard error response format
type ErrorResponse struct {
    // HTTP status code
    // Example: 400
    Code int `json:"code" example:"400"`
    
    // Error message
    // Example: Invalid input
    Error string `json:"error" example:"Invalid input"`
    
    // Additional error details
    // Example: {"field":"username","message":"minimum length is 3"}
    Details map[string]interface{} `json:"details,omitempty"`
}

// SuccessResponse represents success response
type SuccessResponse struct {
    // HTTP status code
    // Example: 200
    Code int `json:"code" example:"200"`
    
    // Success message
    // Example: Operation completed successfully
    Message string `json:"message" example:"Operation completed successfully"`
    
    // Response data
    Data interface{} `json:"data,omitempty"`
}

// ListUsersResponse represents paginated users response
type ListUsersResponse struct {
    Data []UserResponse `json:"data"`
    Pagination struct {
        Page       int   `json:"page" example:"1"`
        Limit      int   `json:"limit" example:"20"`
        Total      int64 `json:"total" example:"100"`
        TotalPages int   `json:"total_pages" example:"5"`
    } `json:"pagination"`
}
```

### 4. Route Registration with Swagger

```go
// internal/api/routes.go
package api

import (
    "github.com/gin-gonic/gin"
    "github.com/linskybing/platform-go/internal/api/handler"
)

func RegisterRoutes(r *gin.Engine, userHandler *handler.UserHandler) {
    // API v1 routes
    v1 := r.Group("/api/v1")
    
    // User endpoints
    users := v1.Group("/users")
    {
        // Create user (POST /api/v1/users)
        users.POST("", userHandler.Create)
        
        // List users (GET /api/v1/users)
        users.GET("", userHandler.List)
        
        // Get user (GET /api/v1/users/:id)
        users.GET("/:id", userHandler.Get)
        
        // Update user (PUT /api/v1/users/:id)
        users.PUT("/:id", userHandler.Update)
        
        // Delete user (DELETE /api/v1/users/:id)
        users.DELETE("/:id", userHandler.Delete)
    }
}
```

### 5. Swagger Security Configuration

```go
// main.go
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description JWT token with Bearer prefix
// Example: Authorization: Bearer eyJhbGciOiJIUzI1NiIs...

// @securityDefinitions.oauth2 OAuth2
// @flow implicit
// @authorizationUrl https://example.com/oauth/authorize
// @tokenUrl https://example.com/oauth/token
// @scopes.read Read access to protected resources
// @scopes.write Write access to protected resources

// Middleware to handle Bearer token
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, ErrorResponse{
                Code:  http.StatusUnauthorized,
                Error: "missing authorization token",
            })
            c.Abort()
            return
        }
        
        // Validate token
        if !strings.HasPrefix(token, "Bearer ") {
            c.JSON(http.StatusUnauthorized, ErrorResponse{
                Code:  http.StatusUnauthorized,
                Error: "invalid token format",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### 6. Swagger Generation and Verification

```makefile
# Makefile
.PHONY: swagger

# Generate Swagger documentation
swagger:
	swag init -g cmd/api/main.go

# Verify Swagger documentation
swagger-validate:
	swagger validate ./docs/swagger.json

# View Swagger documentation
swagger-ui:
	echo "Swagger UI available at http://localhost:8080/swagger/index.html"

# Clean generated swagger files
swagger-clean:
	rm -rf docs/
```

### 7. Swagger Best Practices Checklist

- [ ] All handlers have @Summary comment (max 120 characters)
- [ ] All handlers have @Description comment (clear and concise)
- [ ] @Tags used for logical grouping
- [ ] Request/response models documented with @Param and @Success
- [ ] All error cases documented with @Failure
- [ ] HTTP status codes match actual implementation
- [ ] Examples provided for all fields
- [ ] Data types correct and match implementation
- [ ] Security scheme defined (@Security)
- [ ] Pagination documented with page and limit parameters
- [ ] Enums documented with @Param type/enum
- [ ] Required fields marked in DTO struct tags
- [ ] Generated docs committed to version control
- [ ] Swagger endpoint available in running application
- [ ] API documentation updated with code changes

### 8. Swagger to Code Consistency

Ensure Swagger documentation always matches implementation:

- [ ] Response status codes match @Success/@Failure
- [ ] Request parameters match @Param declarations
- [ ] Response objects match DTO struct fields
- [ ] Error messages match @Failure descriptions
- [ ] Validation rules documented in field comments
- [ ] Examples are realistic and valid
- [ ] Data type conversions documented
- [ ] Deprecated endpoints marked with @Deprecated

### RESTful Resource Naming

```go
// Good: Use plural nouns for collections
GET    /api/v1/users
POST   /api/v1/users
GET    /api/v1/users/:id
PUT    /api/v1/users/:id
DELETE /api/v1/users/:id

// Good: Use nested resources for relationships
GET    /api/v1/projects/:id/members
POST   /api/v1/projects/:id/members
DELETE /api/v1/projects/:id/members/:user_id

// Good: Use query parameters for filtering
GET /api/v1/jobs?status=running&user_id=123&page=2&limit=20

// Bad: Using verbs in URLs
POST /api/v1/createUser
GET  /api/v1/getUser/:id
```

### 2. Request/Response DTOs

```go
// Define clear input DTOs with validation tags
type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=32,alphanum"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    Role     string `json:"role" binding:"omitempty,oneof=user admin"`
}

// Define clear output DTOs (never expose internal models directly)
type UserResponse struct {
    ID        uint      `json:"id"`
    Username  string    `json:"username"`
    Email     string    `json:"email"`
    Role      string    `json:"role"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// Convert domain model to DTO
func ToUserResponse(user *domain.User) *UserResponse {
    return &UserResponse{
        ID:        user.ID,
        Username:  user.Username,
        Email:     user.Email,
        Role:      user.Role,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
        // Note: Never include Password in response
    }
}
```

### 3. Standardized Response Format

```go
// Success response wrapper
type SuccessResponse struct {
    Code    int         `json:"code"`
    Message string      `json:"message,omitempty"`
    Data    interface{} `json:"data,omitempty"`
}

// Error response wrapper
type ErrorResponse struct {
    Code    int                    `json:"code"`
    Error   string                 `json:"error"`
    Details map[string]interface{} `json:"details,omitempty"`
}

// Paginated response
type PaginatedResponse struct {
    Code       int         `json:"code"`
    Data       interface{} `json:"data"`
    Pagination struct {
        Page       int `json:"page"`
        Limit      int `json:"limit"`
        Total      int `json:"total"`
        TotalPages int `json:"total_pages"`
    } `json:"pagination"`
}

// Usage in handlers
func (h *UserHandler) GetUser(c *gin.Context) {
    id := c.Param("id")
    
    user, err := h.userService.GetUser(id)
    if err != nil {
        c.JSON(http.StatusNotFound, ErrorResponse{
            Code:  http.StatusNotFound,
            Error: "User not found",
        })
        return
    }
    
    c.JSON(http.StatusOK, SuccessResponse{
        Code: http.StatusOK,
        Data: ToUserResponse(user),
    })
}
```

### 4. Input Validation

```go
// Validate at API boundary
func (h *UserHandler) CreateUser(c *gin.Context) {
    var req CreateUserRequest
    
    // Binding automatically validates using struct tags
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code:  http.StatusBadRequest,
            Error: "Invalid input",
            Details: map[string]interface{}{
                "validation_errors": err.Error(),
            },
        })
        return
    }
    
    // Additional business logic validation
    if err := h.validateBusinessRules(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code:  http.StatusBadRequest,
            Error: err.Error(),
        })
        return
    }
    
    // Process request
    user, err := h.userService.CreateUser(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Code:  http.StatusInternalServerError,
            Error: "Failed to create user",
        })
        return
    }
    
    c.JSON(http.StatusCreated, SuccessResponse{
        Code:    http.StatusCreated,
        Message: "User created successfully",
        Data:    ToUserResponse(user),
    })
}

// Custom validation functions
func validateUsername(fl validator.FieldLevel) bool {
    username := fl.Field().String()
    matched, _ := regexp.MatchString("^[a-z0-9-]+$", username)
    return matched
}

// Register custom validator
func RegisterCustomValidators(v *validator.Validate) {
    v.RegisterValidation("username", validateUsername)
}
```

### 5. Error Handling & HTTP Status Codes

```go
// Use appropriate HTTP status codes
func (h *Handler) HandleError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, ErrNotFound):
        c.JSON(http.StatusNotFound, ErrorResponse{
            Code:  http.StatusNotFound,
            Error: "Resource not found",
        })
        
    case errors.Is(err, ErrUnauthorized):
        c.JSON(http.StatusUnauthorized, ErrorResponse{
            Code:  http.StatusUnauthorized,
            Error: "Unauthorized access",
        })
        
    case errors.Is(err, ErrForbidden):
        c.JSON(http.StatusForbidden, ErrorResponse{
            Code:  http.StatusForbidden,
            Error: "Access forbidden",
        })
        
    case errors.Is(err, ErrConflict):
        c.JSON(http.StatusConflict, ErrorResponse{
            Code:  http.StatusConflict,
            Error: "Resource already exists",
        })
        
    case errors.Is(err, ErrValidation):
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code:  http.StatusBadRequest,
            Error: err.Error(),
        })
        
    default:
        log.Printf("[ERROR] Unexpected error: %v", err)
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Code:  http.StatusInternalServerError,
            Error: "Internal server error",
        })
    }
}

// HTTP Status Code Guide:
// 200 OK - Successful GET, PUT, PATCH
// 201 Created - Successful POST
// 204 No Content - Successful DELETE
// 400 Bad Request - Validation errors
// 401 Unauthorized - Authentication required
// 403 Forbidden - Authenticated but not authorized
// 404 Not Found - Resource doesn't exist
// 409 Conflict - Resource already exists
// 422 Unprocessable Entity - Semantic errors
// 500 Internal Server Error - Server-side errors
```

### 6. Middleware Patterns

```go
// Authentication middleware
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            c.JSON(http.StatusUnauthorized, ErrorResponse{
                Code:  http.StatusUnauthorized,
                Error: "Missing authorization token",
            })
            c.Abort()
            return
        }
        
        claims, err := ValidateToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, ErrorResponse{
                Code:  http.StatusUnauthorized,
                Error: "Invalid token",
            })
            c.Abort()
            return
        }
        
        c.Set("userID", claims.UserID)
        c.Set("role", claims.Role)
        c.Next()
    }
}

// Role-based authorization middleware
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusForbidden, ErrorResponse{
                Code:  http.StatusForbidden,
                Error: "Access denied",
            })
            c.Abort()
            return
        }
        
        for _, role := range roles {
            if userRole == role {
                c.Next()
                return
            }
        }
        
        c.JSON(http.StatusForbidden, ErrorResponse{
            Code:  http.StatusForbidden,
            Error: "Insufficient permissions",
        })
        c.Abort()
    }
}

// Request logging middleware
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        method := c.Request.Method
        
        c.Next()
        
        duration := time.Since(start)
        statusCode := c.Writer.Status()
        
        log.Printf("[%s] %s %s - %d (%v)",
            method, path, c.ClientIP(), statusCode, duration)
    }
}

// Rate limiting middleware
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
    limiter := rate.NewLimiter(rate.Every(window/time.Duration(limit)), limit)
    
    return func(c *gin.Context) {
        if !limiter.Allow() {
            c.JSON(http.StatusTooManyRequests, ErrorResponse{
                Code:  http.StatusTooManyRequests,
                Error: "Rate limit exceeded",
            })
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 7. Query Parameter Handling

```go
// Parse and validate query parameters
func (h *JobHandler) ListJobs(c *gin.Context) {
    // Pagination
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 20
    }
    
    // Filtering
    status := c.Query("status")
    userID := c.Query("user_id")
    
    // Sorting
    sortBy := c.DefaultQuery("sort_by", "created_at")
    sortOrder := c.DefaultQuery("sort_order", "desc")
    
    jobs, total, err := h.jobService.ListJobs(c.Request.Context(), &ListJobsParams{
        Page:      page,
        Limit:     limit,
        Status:    status,
        UserID:    userID,
        SortBy:    sortBy,
        SortOrder: sortOrder,
    })
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Code:  http.StatusInternalServerError,
            Error: "Failed to list jobs",
        })
        return
    }
    
    c.JSON(http.StatusOK, PaginatedResponse{
        Code: http.StatusOK,
        Data: jobs,
        Pagination: struct {
            Page       int `json:"page"`
            Limit      int `json:"limit"`
            Total      int `json:"total"`
            TotalPages int `json:"total_pages"`
        }{
            Page:       page,
            Limit:      limit,
            Total:      total,
            TotalPages: (total + limit - 1) / limit,
        },
    })
}
```

### 8. File Upload Handling

```go
// Handle file uploads with validation
func (h *Handler) UploadFile(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code:  http.StatusBadRequest,
            Error: "No file uploaded",
        })
        return
    }
    
    // Validate file size (max 10MB)
    if file.Size > 10*1024*1024 {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code:  http.StatusBadRequest,
            Error: "File size exceeds 10MB limit",
        })
        return
    }
    
    // Validate file type
    allowedTypes := map[string]bool{
        "image/jpeg": true,
        "image/png":  true,
        "application/pdf": true,
    }
    
    contentType := file.Header.Get("Content-Type")
    if !allowedTypes[contentType] {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code:  http.StatusBadRequest,
            Error: "Invalid file type",
        })
        return
    }
    
    // Save file
    dst := filepath.Join("/uploads", file.Filename)
    if err := c.SaveUploadedFile(file, dst); err != nil {
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Code:  http.StatusInternalServerError,
            Error: "Failed to save file",
        })
        return
    }
    
    c.JSON(http.StatusOK, SuccessResponse{
        Code:    http.StatusOK,
        Message: "File uploaded successfully",
        Data: map[string]interface{}{
            "filename": file.Filename,
            "size":     file.Size,
        },
    })
}
```

### 9. Context & Timeout Management

```go
// Set timeouts for long operations
func (h *Handler) ProcessLongOperation(c *gin.Context) {
    ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
    defer cancel()
    
    result, err := h.service.LongOperation(ctx)
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            c.JSON(http.StatusGatewayTimeout, ErrorResponse{
                Code:  http.StatusGatewayTimeout,
                Error: "Operation timeout",
            })
            return
        }
        
        c.JSON(http.StatusInternalServerError, ErrorResponse{
            Code:  http.StatusInternalServerError,
            Error: "Operation failed",
        })
        return
    }
    
    c.JSON(http.StatusOK, SuccessResponse{
        Code: http.StatusOK,
        Data: result,
    })
}
```

### 10. API Versioning

```go
// Version APIs in URL path
func RegisterRoutes(r *gin.Engine) {
    v1 := r.Group("/api/v1")
    {
        v1.GET("/users", handlers.ListUsers)
        v1.POST("/users", handlers.CreateUser)
    }
    
    v2 := r.Group("/api/v2")
    {
        v2.GET("/users", handlers.ListUsersV2) // New implementation
        v2.POST("/users", handlers.CreateUserV2)
    }
}
```

## Handler Template

```go
type ResourceHandler struct {
    service *ResourceService
}

func NewResourceHandler(service *ResourceService) *ResourceHandler {
    return &ResourceHandler{service: service}
}

// GET /api/v1/resources/:id
func (h *ResourceHandler) Get(c *gin.Context) {
    id := c.Param("id")
    
    resource, err := h.service.Get(c.Request.Context(), id)
    if err != nil {
        h.handleError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, SuccessResponse{
        Code: http.StatusOK,
        Data: ToResourceResponse(resource),
    })
}

// POST /api/v1/resources
func (h *ResourceHandler) Create(c *gin.Context) {
    var req CreateResourceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code:  http.StatusBadRequest,
            Error: "Invalid input",
        })
        return
    }
    
    resource, err := h.service.Create(c.Request.Context(), &req)
    if err != nil {
        h.handleError(c, err)
        return
    }
    
    c.JSON(http.StatusCreated, SuccessResponse{
        Code:    http.StatusCreated,
        Message: "Resource created successfully",
        Data:    ToResourceResponse(resource),
    })
}

// PUT /api/v1/resources/:id
func (h *ResourceHandler) Update(c *gin.Context) {
    id := c.Param("id")
    
    var req UpdateResourceRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, ErrorResponse{
            Code:  http.StatusBadRequest,
            Error: "Invalid input",
        })
        return
    }
    
    resource, err := h.service.Update(c.Request.Context(), id, &req)
    if err != nil {
        h.handleError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, SuccessResponse{
        Code:    http.StatusOK,
        Message: "Resource updated successfully",
        Data:    ToResourceResponse(resource),
    })
}

// DELETE /api/v1/resources/:id
func (h *ResourceHandler) Delete(c *gin.Context) {
    id := c.Param("id")
    
    err := h.service.Delete(c.Request.Context(), id)
    if err != nil {
        h.handleError(c, err)
        return
    }
    
    c.JSON(http.StatusOK, SuccessResponse{
        Code:    http.StatusOK,
        Message: "Resource deleted successfully",
    })
}

func (h *ResourceHandler) handleError(c *gin.Context, err error) {
    // Implement error handling logic
}
```

## Performance Tips

- Use `c.ShouldBind` instead of `c.Bind` to avoid automatic 400 responses
- Cache expensive computations in middleware
- Use connection pooling for database
- Enable Gzip compression for responses
- Set appropriate timeouts
- Use streaming for large responses

## Security Checklist

- [ ] Validate all user input
- [ ] Sanitize file paths and names
- [ ] Use HTTPS in production
- [ ] Implement rate limiting
- [ ] Add CORS configuration
- [ ] Never expose internal errors to clients
- [ ] Use secure headers (helmet middleware)
- [ ] Implement request size limits
- [ ] Log security events
