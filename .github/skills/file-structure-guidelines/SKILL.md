---
name: file-structure-guidelines
description: Code file organization, modular design, directory structure planning, and 200-line file limit implementation for maintainable platform-go codebase
---

# File Structure and Modularization Guidelines

This skill provides comprehensive guidelines for organizing code files into well-structured directories with 200-line file limits for optimal maintainability.

## When to Use

Apply this skill when:
- Creating new feature packages or modules
- Splitting large monolithic files
- Designing directory structure for new domain
- Refactoring existing code for maintainability
- Planning package-level code organization
- Implementing domain-driven design patterns
- Setting up new services or components
- Reviewing code organization for scalability

## Core Principle: 200-Line File Limit

### Why 200 Lines?

```
Rationale:
- Cognitive load: 200 lines is maximum for easy comprehension
- Review speed: Faster pull request reviews
- Testing: Easier to achieve high test coverage
- Reusability: More granular, focused components
- Maintenance: Clear responsibility per file
- Search: Easier to find code with specific behavior
```

### Benefits of File Splitting

```
Before (monolithic 800 lines):
- Hard to understand full responsibility
- Difficult test coverage
- Mixed concerns (business logic + data access)
- Merge conflicts likely
- Slow code reviews

After (split to 4 x 200 lines):
- Clear responsibility per file
- Easy to test (focused components)
- Separated concerns (clear layers)
- Fewer merge conflicts
- Fast code reviews
```

## Directory Structure Strategy

### 1. Feature-Based Directory Organization

```
internal/
├── user/                              # User domain
│   ├── domain/
│   │   ├── user.go                   # User entity (< 100 lines)
│   │   ├── errors.go                 # User domain errors (< 50 lines)
│   │   └── constants.go              # User constants (< 50 lines)
│   │
│   ├── repository/
│   │   ├── user_repository.go        # Interface (< 50 lines)
│   │   ├── postgres_user.go          # PostgreSQL implementation (< 200 lines)
│   │   └── user_queries.go           # Query builders (< 150 lines)
│   │
│   ├── service/
│   │   ├── user_service.go           # Interface (< 50 lines)
│   │   ├── create_user.go            # Create logic (< 150 lines)
│   │   ├── update_user.go            # Update logic (< 150 lines)
│   │   ├── get_user.go               # Get logic (< 100 lines)
│   │   ├── list_users.go             # List logic (< 150 lines)
│   │   └── delete_user.go            # Delete logic (< 100 lines)
│   │
│   ├── handler/
│   │   ├── user_handler.go           # Interface (< 50 lines)
│   │   ├── create_handler.go         # Create endpoint (< 100 lines)
│   │   ├── update_handler.go         # Update endpoint (< 100 lines)
│   │   ├── get_handler.go            # Get endpoint (< 80 lines)
│   │   ├── list_handler.go           # List endpoint (< 100 lines)
│   │   └── delete_handler.go         # Delete endpoint (< 80 lines)
│   │
│   ├── dto/
│   │   ├── request.go                # Request DTOs (< 150 lines)
│   │   └── response.go               # Response DTOs (< 150 lines)
│   │
│   └── router.go                     # Route registration (< 100 lines)
│
├── job/                               # Job domain (similar structure)
│   ├── domain/
│   ├── repository/
│   ├── service/
│   ├── handler/
│   ├── dto/
│   └── router.go
│
└── storage/                           # Storage domain
    ├── domain/
    ├── repository/
    ├── service/
    ├── handler/
    ├── dto/
    └── router.go
```

### 2. File Counting Example: User Domain

```
User domain total structure:
├── domain/
│   ├── user.go                       (95 lines)  - Entity definition
│   ├── errors.go                     (32 lines)  - Domain errors
│   └── constants.go                  (28 lines)  - Constants
│
├── repository/
│   ├── user_repository.go            (45 lines)  - Interface
│   ├── postgres_user.go              (180 lines) - Implementation
│   └── user_queries.go               (120 lines) - Query builders
│
├── service/
│   ├── user_service.go               (42 lines)  - Interface
│   ├── create_user.go                (145 lines) - Create operation
│   ├── update_user.go                (160 lines) - Update operation
│   ├── get_user.go                   (85 lines)  - Get operation
│   ├── list_users.go                 (140 lines) - List operation
│   └── delete_user.go                (92 lines)  - Delete operation
│
├── handler/
│   ├── user_handler.go               (48 lines)  - Interface
│   ├── create_handler.go             (92 lines)  - Create handler
│   ├── update_handler.go             (95 lines)  - Update handler
│   ├── get_handler.go                (75 lines)  - Get handler
│   ├── list_handler.go               (88 lines)  - List handler
│   └── delete_handler.go             (72 lines)  - Delete handler
│
├── dto/
│   ├── request.go                    (140 lines) - Request DTOs
│   └── response.go                   (135 lines) - Response DTOs
│
└── router.go                         (65 lines)  - Route setup

Total: ~1,900 lines organized in 17 files
Average: 112 lines per file
All files < 200 lines
```

## Large File Splitting Strategy

### 1. Service Layer Splitting (Monolithic 600 Lines)

```go
// BEFORE: user_service.go (600 lines - too large)
type UserService struct {
    userRepo repository.UserRepository
    mailRepo repository.MailRepository
    cacheRepo repository.CacheRepository
}

func (s *UserService) CreateUser(...) { ... }      // 120 lines
func (s *UserService) UpdateUser(...) { ... }      // 130 lines
func (s *UserService) GetUser(...) { ... }         // 80 lines
func (s *UserService) ListUsers(...) { ... }       // 110 lines
func (s *UserService) DeleteUser(...) { ... }      // 95 lines
// Plus helper methods, validators, etc.

// AFTER: Split into focused files
// user_service.go (45 lines)
type UserService interface {
    CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error)
    UpdateUser(ctx context.Context, id uint, req *UpdateUserRequest) (*User, error)
    GetUser(ctx context.Context, id uint) (*User, error)
    ListUsers(ctx context.Context, filter *ListFilter) ([]*User, int64, error)
    DeleteUser(ctx context.Context, id uint) error
}

type userService struct {
    userRepo  repository.UserRepository
    mailRepo  repository.MailRepository
    cacheRepo repository.CacheRepository
}

// create_user.go (145 lines)
func (s *userService) CreateUser(ctx context.Context, req *CreateUserRequest) (*User, error) {
    // Validate input
    if err := validateCreateRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Check duplicate
    existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
    if existing != nil {
        return nil, fmt.Errorf("user already exists")
    }
    
    // Hash password
    hashedPassword, err := hashPassword(req.Password)
    if err != nil {
        return nil, fmt.Errorf("password hashing failed: %w", err)
    }
    
    // Create user
    user := &User{
        Username: req.Username,
        Email:    req.Email,
        Password: hashedPassword,
        Role:     req.Role,
    }
    
    created, err := s.userRepo.Create(ctx, user)
    if err != nil {
        return nil, fmt.Errorf("user creation failed: %w", err)
    }
    
    // Send welcome email
    if err := s.sendWelcomeEmail(ctx, created); err != nil {
        // Log but don't fail
        log.Printf("failed to send welcome email: %v", err)
    }
    
    // Cache user
    _ = s.cacheRepo.Set(ctx, fmt.Sprintf("user:%d", created.ID), created)
    
    return created, nil
}

func (s *userService) sendWelcomeEmail(ctx context.Context, user *User) error {
    // Implementation
    return s.mailRepo.SendWelcomeEmail(ctx, user.Email, user.Username)
}

// update_user.go (160 lines)
func (s *userService) UpdateUser(ctx context.Context, id uint, req *UpdateUserRequest) (*User, error) {
    // Fetch existing user
    existing, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("user not found: %w", err)
    }
    
    // Validate update request
    if err := validateUpdateRequest(req); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }
    
    // Update fields
    if req.Username != "" && req.Username != existing.Username {
        existing.Username = req.Username
    }
    if req.Email != "" && req.Email != existing.Email {
        duplicate, _ := s.userRepo.GetByEmail(ctx, req.Email)
        if duplicate != nil && duplicate.ID != id {
            return nil, fmt.Errorf("email already in use")
        }
        existing.Email = req.Email
    }
    if req.Password != "" {
        hashedPassword, err := hashPassword(req.Password)
        if err != nil {
            return nil, fmt.Errorf("password hashing failed: %w", err)
        }
        existing.Password = hashedPassword
    }
    if req.Role != "" {
        existing.Role = req.Role
    }
    
    // Save changes
    updated, err := s.userRepo.Update(ctx, existing)
    if err != nil {
        return nil, fmt.Errorf("user update failed: %w", err)
    }
    
    // Invalidate cache
    s.cacheRepo.Delete(ctx, fmt.Sprintf("user:%d", id))
    
    return updated, nil
}

// get_user.go (85 lines)
func (s *userService) GetUser(ctx context.Context, id uint) (*User, error) {
    // Try cache first
    cacheKey := fmt.Sprintf("user:%d", id)
    if cached, err := s.cacheRepo.Get(ctx, cacheKey); err == nil {
        return cached.(*User), nil
    }
    
    // Fetch from database
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("user not found: %w", err)
    }
    
    // Cache result
    s.cacheRepo.Set(ctx, cacheKey, user)
    
    return user, nil
}

// list_users.go (140 lines)
func (s *userService) ListUsers(ctx context.Context, filter *ListFilter) ([]*User, int64, error) {
    // Validate filter
    if filter.Page < 1 {
        filter.Page = 1
    }
    if filter.Limit < 1 || filter.Limit > 100 {
        filter.Limit = 20
    }
    
    // Fetch from database
    users, total, err := s.userRepo.List(ctx, filter)
    if err != nil {
        return nil, 0, fmt.Errorf("list users failed: %w", err)
    }
    
    return users, total, nil
}

// delete_user.go (92 lines)
func (s *userService) DeleteUser(ctx context.Context, id uint) error {
    // Fetch user
    user, err := s.userRepo.GetByID(ctx, id)
    if err != nil {
        return fmt.Errorf("user not found: %w", err)
    }
    
    // Prevent admin deletion
    if user.Role == "admin" {
        return fmt.Errorf("cannot delete admin user")
    }
    
    // Delete from database
    if err := s.userRepo.Delete(ctx, id); err != nil {
        return fmt.Errorf("user deletion failed: %w", err)
    }
    
    // Invalidate cache
    s.cacheRepo.Delete(ctx, fmt.Sprintf("user:%d", id))
    
    return nil
}
```

## Repository Layer Splitting

### Large Repository File (400 Lines → Split)

```go
// BEFORE: user_repository.go (400 lines)
// Contains:
// - Interface definition
// - PostgreSQL implementation
// - Query helpers
// - Pagination logic
// - Transaction handling

// AFTER: Split into 3 files

// user_repository.go (45 lines) - Interface only
package repository

type UserRepository interface {
    GetByID(ctx context.Context, id uint) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
    Create(ctx context.Context, user *User) (*User, error)
    Update(ctx context.Context, user *User) (*User, error)
    Delete(ctx context.Context, id uint) error
    List(ctx context.Context, filter *ListFilter) ([]*User, int64, error)
}

// postgres_user.go (180 lines) - Core implementation
package repository

import "gorm.io/gorm"

type postgresUserRepository struct {
    db *gorm.DB
}

func NewPostgresUserRepository(db *gorm.DB) UserRepository {
    return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) GetByID(ctx context.Context, id uint) (*User, error) {
    var user User
    if err := r.db.WithContext(ctx).First(&user, "id = ?", id).Error; err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &user, nil
}

func (r *postgresUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
    var user User
    if err := r.db.WithContext(ctx).First(&user, "email = ?", email).Error; err != nil {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    return &user, nil
}

func (r *postgresUserRepository) Create(ctx context.Context, user *User) (*User, error) {
    if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }
    return user, nil
}

func (r *postgresUserRepository) Update(ctx context.Context, user *User) (*User, error) {
    if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
        return nil, fmt.Errorf("failed to update user: %w", err)
    }
    return user, nil
}

func (r *postgresUserRepository) Delete(ctx context.Context, id uint) error {
    if err := r.db.WithContext(ctx).Delete(&User{}, "id = ?", id).Error; err != nil {
        return fmt.Errorf("failed to delete user: %w", err)
    }
    return nil
}

func (r *postgresUserRepository) List(ctx context.Context, filter *ListFilter) ([]*User, int64, error) {
    var users []*User
    var total int64
    
    query := r.db.WithContext(ctx)
    query = r.applyFilters(query, filter)
    
    if err := query.Model(&User{}).Count(&total).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to count users: %w", err)
    }
    
    offset := (filter.Page - 1) * filter.Limit
    if err := query.Offset(offset).Limit(filter.Limit).
        Order("created_at DESC").
        Find(&users).Error; err != nil {
        return nil, 0, fmt.Errorf("failed to list users: %w", err)
    }
    
    return users, total, nil
}

// user_queries.go (120 lines) - Query builders and helpers
package repository

import "gorm.io/gorm"

func (r *postgresUserRepository) applyFilters(query *gorm.DB, filter *ListFilter) *gorm.DB {
    if filter.Role != "" {
        query = query.Where("role = ?", filter.Role)
    }
    
    if filter.Status != "" {
        query = query.Where("status = ?", filter.Status)
    }
    
    if filter.SearchTerm != "" {
        query = query.Where("username ILIKE ? OR email ILIKE ?", 
            "%"+filter.SearchTerm+"%",
            "%"+filter.SearchTerm+"%")
    }
    
    if !filter.CreatedAfter.IsZero() {
        query = query.Where("created_at >= ?", filter.CreatedAfter)
    }
    
    if !filter.CreatedBefore.IsZero() {
        query = query.Where("created_at <= ?", filter.CreatedBefore)
    }
    
    return query
}

func (r *postgresUserRepository) validateUser(user *User) error {
    if user.Username == "" {
        return fmt.Errorf("username required")
    }
    if user.Email == "" {
        return fmt.Errorf("email required")
    }
    if len(user.Password) < 60 {
        return fmt.Errorf("password must be hashed")
    }
    return nil
}
```

## Handler Layer Splitting

### Split Handlers by Operation

```go
// BEFORE: user_handler.go (400 lines)
// - Create handler
// - Get handler
// - Update handler
// - List handler
// - Delete handler

// AFTER: Split by operation

// user_handler.go (45 lines) - Interface
package handler

type UserHandler interface {
    Create(c *gin.Context)
    Get(c *gin.Context)
    Update(c *gin.Context)
    List(c *gin.Context)
    Delete(c *gin.Context)
}

type userHandler struct {
    userService service.UserService
}

// create_handler.go (95 lines)
package handler

func (h *userHandler) Create(c *gin.Context) {
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

// get_handler.go (75 lines)
package handler

func (h *userHandler) Get(c *gin.Context) {
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

// update_handler.go (98 lines)
package handler

func (h *userHandler) Update(c *gin.Context) {
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

// list_handler.go (88 lines)
package handler

func (h *userHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    
    filter := &service.ListFilter{
        Page:  page,
        Limit: limit,
        Role:  c.Query("role"),
    }
    
    users, total, err := h.userService.ListUsers(c.Request.Context(), filter)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    
    responses := make([]dto.UserResponse, 0, len(users))
    for _, user := range users {
        responses = append(responses, *dto.ToUserResponse(user))
    }
    
    c.JSON(http.StatusOK, gin.H{
        "data": responses,
        "pagination": gin.H{
            "page":  page,
            "limit": limit,
            "total": total,
        },
    })
}

// delete_handler.go (72 lines)
package handler

func (h *userHandler) Delete(c *gin.Context) {
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

## File Naming Conventions

```
Naming by Operation/Function:
create_user.go       - User creation logic
update_user.go       - User update logic
get_user.go          - User retrieval logic
list_users.go        - User listing logic
delete_user.go       - User deletion logic

Naming by Responsibility:
user.go              - Entity definition only
user_repository.go   - Repository interface
postgres_user.go     - PostgreSQL implementation
user_service.go      - Service interface
user_handler.go      - Handler interface
user_queries.go      - Query builders
user_validator.go    - Validation logic

Avoid Generic Names:
helpers.go           - Too generic (use specific names)
utils.go             - Too generic (use specific names)
types.go             - Too generic (use domain-specific names)
```

## Directory Benefits for 200-Line Limit

### Problem: Too Many Files Without Structure

```
Without directories (chaos):
internal/
├── create_user.go
├── update_user.go
├── get_user.go
├── delete_user.go
├── list_users.go
├── create_job.go
├── update_job.go
├── get_job.go
├── delete_job.go
├── list_jobs.go
├── user_repository.go
├── job_repository.go
├── user_service.go
├── job_service.go
├── user_handler.go
├── job_handler.go
└── 30+ more files...
```

### Solution: Organized Directories

```
With directories (organized):
internal/
├── user/
│   ├── service/
│   │   ├── create_user.go
│   │   ├── update_user.go
│   │   ├── get_user.go
│   │   ├── delete_user.go
│   │   └── list_users.go
│   ├── repository/
│   │   ├── user_repository.go
│   │   ├── postgres_user.go
│   │   └── user_queries.go
│   ├── handler/
│   │   ├── create_handler.go
│   │   ├── update_handler.go
│   │   ├── get_handler.go
│   │   ├── delete_handler.go
│   │   └── list_handler.go
│   ├── domain/
│   │   ├── user.go
│   │   └── errors.go
│   └── router.go
│
└── job/
    ├── service/
    ├── repository/
    ├── handler/
    ├── domain/
    └── router.go
```

## Implementation Checklist

- [ ] Max file size is 200 lines (not counting imports/blank lines)
- [ ] Each file has single, clear responsibility
- [ ] Directory structure mirrors domain/feature boundaries
- [ ] Related files grouped in same directory
- [ ] Files named by operation or function (verb_noun.go)
- [ ] No generic package names (helpers, utils, types)
- [ ] Interfaces defined in focused files (< 50 lines)
- [ ] Implementation split by operation/concern
- [ ] Query builders in separate files
- [ ] Handlers split by endpoint/operation
- [ ] DTOs organized by request/response
- [ ] No circular dependencies between files
- [ ] Easy to navigate and understand structure
- [ ] Quick lookup of specific functionality
- [ ] Minimal blast radius for changes

## Migration Strategy

When refactoring existing large files:

1. Identify distinct operations/responsibilities
2. Group related functionality
3. Create new directory if needed
4. Extract each operation to separate file
5. Ensure all < 200 lines
6. Update imports
7. Test thoroughly
8. Update documentation

## Production Quality Checklist

- [ ] All files have documentation (package comment)
- [ ] File purpose is clear from name
- [ ] Functions are exported or unexported appropriately
- [ ] No cross-package circular dependencies
- [ ] Each file is independently testable
- [ ] Related tests in mirrored directory structure
- [ ] File organization enables easy navigation
- [ ] New developers can quickly understand structure
- [ ] Code search yields relevant results
- [ ] Reasonable compile times (< 10 seconds)

## Anti-Patterns (FORBIDDEN)

- Never exceed 200 lines per file without strong justification
- Never mix multiple domains in one file
- Never create generic utility files
- Never have unclear file responsibilities
- Never skip directory organization to keep flat structure
- Never have inter-file dependencies that create circles
- Never use ambiguous file names
- Never put business logic and data access in same file
- Never skip package-level documentation
- Never organize only by file type (all handlers in one dir)

## Handler Organization by Domain (K8s Example)

### When to Use Subdirectories for Handlers

Create subdirectories for handler groups when you have:
- 6+ related handlers for the same domain/service
- Clear separation by business concern (e.g., storage, compute, networking)
- Need to simplify package imports and reduce naming verbosity

### Kubernetes Handlers Organization Example

```
internal/api/handlers/
├── container.go                    # Main handler container (registry)
├── user_handler.go                 # User CRUD queries
├── user_auth_handler.go            # User authentication
├── user_group_handler.go           # User-group queries
├── user_group_mutate_handler.go    # User-group mutations
├── project_handler.go              # Project queries
├── project_mutate_handler.go       # Project mutations
├── image_handler.go                # Image management
├── image_request_handler.go        # Image requests/approval
├── image_pull_handler.go           # Image pull jobs
├── job_handler.go                  # Job management
├── ws_handler.go                   # WebSocket handlers
│
├── k8s/                            # K8s domain subdirectory
│   ├── base.go                     # Base K8s operations (171 lines)
│   ├── job.go                      # K8s job logging (161 lines)
│   ├── user_storage.go             # User storage management (125 lines)
│   ├── project_storage.go          # Project storage (180 lines)
│   ├── user_storage_browse.go      # Storage file browser (125 lines)
│   └── project_storage_proxy.go    # Storage proxy (138 lines)
│
└── other_handlers...

Benefits of k8s/ subdirectory:
- Cleaner naming (base.go vs k8s_handler.go)
- Clear domain grouping (6 related files together)
- Easy to navigate (separate pkg for K8s)
- Reduced import paths (handlers.k8s.PodLogHandler)
- Scalability (easily add more K8s handlers later)
```

### Naming Conventions for Handler Subdirectories

```
✓ Good:  k8s/base.go, k8s/job.go              # Clear, concise
✓ Good:  auth/login.go, auth/register.go      # Business concern
✓ Good:  storage/user.go, storage/project.go  # Domain separation

✗ Bad:   k8s_handler.go, k8s_job_handler.go   # Repetitive suffix
✗ Bad:   handlers/k8s/handlers.go             # Redundant naming
✗ Bad:   handlers/kubernetes/                 # Overly verbose
```

### Container Registry Pattern

When using handler subdirectories, register them in container.go:

```go
// container.go - Handler registry
package handlers

import (
    "github.com/linskybing/platform-go/internal/api/handlers/k8s"
    // ... other imports
)

type Handlers struct {
    // Root-level handlers
    User *UserHandler
    Project *ProjectHandler
    
    // Subdirectory handlers
    K8s *k8s.K8sHandler  // Fully qualified package name
}

func New(...) *Handlers {
    return &Handlers{
        User:   NewUserHandler(...),
        K8s:    k8s.NewK8sHandler(...),  // Subdirectory initialization
    }
}
```
