---
name: access-control-best-practices
description: Role-based access control (RBAC) patterns, authorization middleware, permission hierarchy, and security implementation for platform-go API endpoints
---

# Access Control & Authorization Best Practices

**Version**: 1.0  
**Date**: February 3, 2026  
**Status**: Active

---

## Overview

This skill document provides comprehensive guidance for implementing role-based access control (RBAC) and resource-level authorization throughout the platform-go project. It establishes the authorization architecture, permission hierarchy, and implementation patterns for all access-controlled endpoints.

### Key Principles

1. **Principle of Least Privilege** - Users get minimum required permissions
2. **Role-Based Access Control (RBAC)** - Permission hierarchy by role
3. **Resource-Level Authorization** - Fine-grained control per resource
4. **Explicit Denial** - Default deny, explicit allow
5. **Audit & Logging** - All access decisions logged

---

## When to Use This Skill

**Always reference this skill when**:

- [YES] **Creating new API endpoints** - Determine required authorization level
- [YES] **Implementing access control** - Select appropriate middleware pattern
- [YES] **Reviewing code with permissions** - Verify authorization is correct
- [YES] **Debugging authorization issues** - Understand decision flow and extractors
- [YES] **Adding role-based features** - Implement consistent patterns
- [YES] **Conducting security audits** - Verify all endpoints are protected
- [YES] **Onboarding new developers** - Learn authorization framework
- [YES] **Refactoring existing endpoints** - Migrate to standardized patterns
- [YES] **Writing authorization tests** - Follow established test strategies
- [YES] **Planning new features** - Design with proper access control

**Common Scenarios**:

| Scenario | Use This Section |
|----------|------------------|
| "What role should access this endpoint?" | [Authorization Hierarchy](#authorization-hierarchy) + [Permission Matrix](#complete-permission-matrix) |
| "How do I protect this route?" | [Common Authorization Patterns](#common-authorization-patterns) |
| "How to extract resource ID?" | [Resource-Level Authorization](#resource-level-authorization) |
| "How to test authorization?" | [Implementation Examples](#implementation-examples) |
| "What are common mistakes?" | [Common Mistakes & Fixes](#common-mistakes--fixes) |
| "How to migrate existing code?" | [Migration Guide](#migration-guide-adding-authorization-to-existing-endpoints) |

---

## Complete Permission Matrix

### Role-Based Permission Table

This table shows all permissions across all roles. Use this as a reference when determining what role should access a specific endpoint.

| Resource/Action | Super Admin | Group Admin | Group Manager | Group Member | User (Self) | Public |
|----------------|-------------|-------------|---------------|--------------|-------------|--------|
| **System Configuration** |
| System settings | [YES] Write | [NO] | [NO] | [NO] | [NO] | [NO] |
| View audit logs | [YES] All | [YES] Group only | [NO] | [NO] | [NO] | [NO] |
| **User Management** |
| Create user | [YES] | [NO] | [NO] | [NO] | [NO] | [YES] Register |
| Update own profile | [YES] | [YES] | [YES] | [YES] | [YES] Own only | [NO] |
| Update any user | [YES] | [NO] | [NO] | [NO] | [NO] | [NO] |
| Delete own account | [YES] | [YES] | [YES] | [YES] | [YES] Own only | [NO] |
| Delete any user | [YES] | [NO] | [NO] | [NO] | [NO] | [NO] |
| View user list | [YES] All | [YES] Group only | [NO] | [NO] | [NO] | [NO] |
| **Group Management** |
| Create group | [YES] | [NO] | [NO] | [NO] | [NO] | [NO] |
| Update group | [YES] | [YES] Own group | [NO] | [NO] | [NO] | [NO] |
| Delete group | [YES] | [YES] Own group | [NO] | [NO] | [NO] | [NO] |
| View groups | [YES] All | [YES] Own groups | [YES] Own groups | [YES] Own groups | [NO] | [NO] |
| Manage group members | [YES] | [YES] Own group | [NO] | [NO] | [NO] | [NO] |
| Assign group roles | [YES] | [YES] Own group | [NO] | [NO] | [NO] | [NO] |
| **Project Management** |
| Create project | [YES] | [YES] In group | [NO] | [NO] | [NO] | [NO] |
| Update project | [YES] | [YES] In group | [YES] Own project | [NO] | [NO] | [NO] |
| Delete project | [YES] | [YES] In group | [NO] | [NO] | [NO] | [NO] |
| View projects | [YES] All | [YES] In group | [YES] In group | [YES] In group | [NO] | [NO] |
| Manage project members | [YES] | [YES] In group | [YES] Own project | [NO] | [NO] | [NO] |
| **Storage/PVC Management** |
| Create storage | [YES] | [YES] In group | [NO] | [NO] | [NO] | [NO] |
| Update storage | [YES] | [YES] In group | [NO] | [NO] | [NO] | [NO] |
| Delete storage | [YES] | [YES] In group | [NO] | [NO] | [NO] | [NO] |
| View storage | [YES] All | [YES] In group | [YES] In group | [YES] In group | [NO] | [NO] |
| Set storage permissions | [YES] | [YES] In group | [NO] | [NO] | [NO] | [NO] |
| Bind PVC to project | [YES] | [YES] In group | [YES] Own project | [NO] | [NO] | [NO] |
| **Image Management** |
| Create image | [YES] | [YES] In group | [YES] In project | [NO] | [NO] | [NO] |
| Update image | [YES] | [YES] In group | [YES] Own image | [NO] | [NO] | [NO] |
| Delete image | [YES] | [YES] In group | [YES] Own image | [NO] | [NO] | [NO] |
| View images | [YES] All | [YES] In group | [YES] In group | [YES] In group | [NO] | [NO] |
| Approve image request | [YES] | [NO] | [NO] | [NO] | [NO] | [NO] |
| Add allowed image | [YES] | [YES] In group | [YES] In project | [NO] | [NO] | [NO] |
| Remove allowed image | [YES] | [YES] In group | [YES] In project | [NO] | [NO] | [NO] |
| **Config File Management** |
| Create config file | [YES] | [YES] In group | [YES] In project | [NO] | [NO] | [NO] |
| Update config file | [YES] | [YES] In group | [YES] Own file | [NO] | [NO] | [NO] |
| Delete config file | [YES] | [YES] In group | [YES] Own file | [NO] | [NO] | [NO] |
| View config files | [YES] All | [YES] In group | [YES] In group | [YES] In group | [NO] | [NO] |
| **Job/Instance Management** |
| Create job | [YES] | [YES] In group | [YES] In project | [YES] In project | [NO] | [NO] |
| Update job | [YES] | [YES] In group | [YES] In project | [YES] Own job | [NO] | [NO] |
| Delete job | [YES] | [YES] In group | [YES] In project | [YES] Own job | [NO] | [NO] |
| View jobs | [YES] All | [YES] In group | [YES] In project | [YES] In project | [NO] | [NO] |
| Execute in container | [YES] All | [YES] In group | [YES] In project | [YES] Own job | [NO] | [NO] |
| View job logs | [YES] All | [YES] In group | [YES] In project | [YES] In project | [NO] | [NO] |
| **Form/GPU Management** |
| Create form | [YES] | [YES] In group | [YES] In project | [NO] | [NO] | [NO] |
| Update form | [YES] | [YES] In group | [YES] Own form | [NO] | [NO] | [NO] |
| Delete form | [YES] | [YES] In group | [YES] Own form | [NO] | [NO] | [NO] |
| View forms | [YES] All | [YES] In group | [YES] In group | [YES] In group | [NO] | [NO] |
| Submit form response | [YES] | [YES] In group | [YES] In project | [YES] In project | [NO] | [NO] |
| **Authentication** |
| Login | [YES] | [YES] | [YES] | [YES] | [YES] | [YES] |
| Logout | [YES] | [YES] | [YES] | [YES] | [YES] | [NO] |
| Register | [YES] | [YES] | [YES] | [YES] | [YES] | [YES] |

### Middleware to Role Mapping

| Middleware Function | Allowed Roles | Use When |
|-------------------|---------------|----------|
| `Admin()` | Super Admin only | System configuration, global operations |
| `GroupAdmin(extractor)` | Super Admin, Group Admin | Group management, project deletion |
| `GroupManager(extractor)` | Super Admin, Group Admin, Group Manager | Project updates, resource configuration |
| `GroupMember(extractor)` | Super Admin, Group Admin, Group Manager, Group Member | User submissions, viewing resources |
| `UserOrAdmin()` | Super Admin, Own User | Profile updates, self-service |
| None (public) | All users (including unauthenticated) | Login, register, public docs |

**Note**: Super Admin always has access to all endpoints, regardless of middleware used.

---

## Authorization Hierarchy

### Global Roles

#### 1. **Super Admin** (Platform Level)
- **Definition**: System administrator with unrestricted access
- **Scope**: All projects, all groups, all users
- **Permissions**:
  - Create/Update/Delete any project
  - Create/Update/Delete any group
  - Manage all user accounts
  - Approve image requests
  - View all audit logs
  - System configuration

**Usage Pattern**:
```go
// Protect endpoint: admin only
route.POST("/admin/users", authMiddleware.Admin(), handler)
```

#### 2. **Group Admin** (Group Level)
- **Definition**: Administrator for a specific group
- **Scope**: All projects within their group
- **Permissions**:
  - Create/Update/Delete projects in group
  - Manage group members (roles)
  - Access all resources in group
  - Approve storage permissions

**Usage Pattern**:
```go
// Delete project: group admin required
route.DELETE("/:id", 
    authMiddleware.GroupAdmin(
        middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
    ), 
    handler
)
```

**Implementation**:
```go
// Extract group ID from resource ID and verify user is group admin
func (am *AuthMiddleware) GroupAdmin(extractor IDExtractor) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("user_id")
        resourceID := extractor(c)
        
        groupID := am.repos.Project.GetGroupIDByProjectID(resourceID)
        isAdmin := am.isGroupAdmin(userID, groupID)
        
        if !isAdmin {
            c.JSON(403, gin.H{"error": "forbidden"})
            c.Abort()
            return
        }
        c.Next()
    }
}
```

#### 3. **Group Manager** (Project Level)
- **Definition**: Manager/lead for a specific project
- **Scope**: Single project or resource
- **Permissions**:
  - Update project configuration
  - Manage project members
  - Create images/forms for project
  - Add/remove allowed images

**Usage Pattern**:
```go
// Update project: manager or admin
route.PUT("/:id", 
    authMiddleware.GroupManager(
        middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
    ), 
    handler
)
```

#### 4. **Group Member** (Workspace Level)
- **Definition**: Regular member of a group/project
- **Scope**: Own submissions and shared resources
- **Permissions**:
  - View project resources
  - Submit jobs to project
  - View allowed images
  - Create instances

**Usage Pattern**:
```go
// Create job: any group member
route.POST("/:id/jobs", 
    authMiddleware.GroupMember(
        middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
    ), 
    handler
)
```

#### 5. **User** (Self Only)
- **Definition**: Authenticated user
- **Scope**: Own resources only
- **Permissions**:
  - Update own profile
  - Delete own account
  - View own jobs/forms

**Usage Pattern**:
```go
// Update user: self or admin
route.PUT("/:id", 
    authMiddleware.UserOrAdmin(), 
    handler
)

// Implementation checks: userID == paramID || isAdmin
```

---

## Resource-Level Authorization

### Authorization Extractors

Pattern for extracting resource identifiers from request context:

#### 1. FromIDParam - Path Parameter
```go
// Extract ID from path parameter /:id
middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)

// Usage
route.PUT("/:id", 
    authMiddleware.GroupManager(middleware.FromIDParam(...)), 
    handler
)
```

#### 2. FromProjectIDInPayload - Request Body
```go
// Extract project_id from JSON body
middleware.FromProjectIDInPayload(configfile.CreateConfigFileInput{})

// Usage
route.POST("", 
    authMiddleware.GroupManager(
        middleware.FromProjectIDInPayload(configfile.CreateConfigFileInput{})
    ), 
    handler
)
```

#### 3. Custom Extractors
```go
// Create custom extractor for complex scenarios
customExtractor := func(c *gin.Context) uint {
    // Extract from multiple sources (body, params, query)
    userID := c.GetUint("user_id")
    resourceID := getResourceID(c)
    return resourceID
}

route.DELETE("", 
    authMiddleware.GroupManager(customExtractor), 
    handler
)
```

---

## Common Authorization Patterns

### Pattern 1: Role-Based Only (Global)
**Use Case**: System configuration, admin panels

```go
// Super admin only
route.GET("/admin/dashboard", authMiddleware.Admin(), handler)
route.POST("/admin/config", authMiddleware.Admin(), handler)
```

**When to use**:
- System-wide configuration
- User management (all users)
- Super-admin functions only

### Pattern 2: Group-Level Authorization
**Use Case**: Project management, resource deletion

```go
// Group admin for specific group
route.DELETE("/projects/:id", 
    authMiddleware.GroupAdmin(
        middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
    ), 
    handler
)
```

**When to use**:
- Destructive operations (delete)
- Group membership changes
- Policy changes

### Pattern 3: Resource-Level Authorization
**Use Case**: Project updates, image management

```go
// Group manager for specific project
route.PUT("/projects/:id", 
    authMiddleware.GroupManager(
        middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
    ), 
    handler
)

route.POST("/projects/:id/images", 
    authMiddleware.GroupManager(
        middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
    ), 
    handler
)
```

**When to use**:
- Update operations
- Add/configure resources
- Management tasks

### Pattern 4: Membership-Based Authorization
**Use Case**: Job creation, resource access

```go
// Any group member
route.POST("/projects/:id/jobs", 
    authMiddleware.GroupMember(
        middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
    ), 
    handler
)

route.GET("/projects/:id/images", 
    authMiddleware.GroupMember(
        middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
    ), 
    handler
)
```

**When to use**:
- Read operations on shared resources
- Create own submissions (jobs, forms)
- Access group resources

### Pattern 5: Self or Admin Authorization
**Use Case**: Profile updates, account management

```go
// Self or super admin
route.PUT("/users/:id", 
    authMiddleware.UserOrAdmin(), 
    handler
)

route.DELETE("/users/:id", 
    authMiddleware.UserOrAdmin(), 
    handler
)
```

**When to use**:
- User profile operations
- Account management
- Self-service with admin override

---

## Implementation Examples

### Example 1: Creating Project-Scoped Routes

```go
// internal/api/routes/project.go
func registerProjectRoutes(auth *gin.RouterGroup, handlers *handlers.Handlers, repos *repository.Repos) {
    projects := auth.Group("/projects")
    {
        // List: public to all authenticated users
        projects.GET("", handlers.Project.GetProjects)
        
        // Create: admin only
        projects.POST("", 
            authMiddleware.Admin(), 
            handlers.Project.CreateProject,
        )
        
        // Update: group manager for this project
        projects.PUT("/:id", 
            authMiddleware.GroupManager(
                middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
            ), 
            handlers.Project.UpdateProject,
        )
        
        // Delete: group admin for this project
        projects.DELETE("/:id", 
            authMiddleware.GroupAdmin(
                middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
            ), 
            handlers.Project.DeleteProject,
        )
        
        // Nested: config files under project
        projects.POST("/:id/config-files",
            authMiddleware.GroupMember(
                middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
            ),
            handlers.ConfigFile.CreateConfigFileHandler,
        )
    }
}
```

### Example 2: Resource Extraction Middleware

```go
// internal/api/middleware/extractors.go
package middleware

import (
    "github.com/gin-gonic/gin"
)

// IDExtractor extracts resource ID from request context
type IDExtractor func(*gin.Context) uint

// FromIDParam extracts ID from URL parameter /:id
func FromIDParam(fetcher func(uint) (uint, error)) IDExtractor {
    return func(c *gin.Context) uint {
        id := c.GetUint("id")  // Gin parses :id from path
        groupID, err := fetcher(id)
        if err != nil {
            c.JSON(404, gin.H{"error": "resource not found"})
            c.Abort()
            return 0
        }
        return groupID
    }
}

// FromProjectIDInPayload extracts project_id from request body
func FromProjectIDInPayload(dto interface{}) IDExtractor {
    return func(c *gin.Context) uint {
        var data map[string]interface{}
        c.BindJSON(&data)
        
        projectID, ok := data["project_id"].(float64)
        if !ok {
            c.JSON(400, gin.H{"error": "project_id required"})
            c.Abort()
            return 0
        }
        
        return uint(projectID)
    }
}
```

### Example 3: Complex Authorization Logic

```go
// internal/api/middleware/auth.go
func (am *AuthMiddleware) GroupManager(extractor IDExtractor) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetUint("user_id")
        claims := c.MustGet("claims").(*jwt.Claims)
        
        // Super admin bypasses all checks
        if claims.IsAdmin {
            c.Next()
            return
        }
        
        // Extract resource group
        groupID := extractor(c)
        
        // Check if user is group manager or admin
        userGroup, err := am.repos.UserGroup.GetUserGroup(
            c.Request.Context(), 
            userID, 
            groupID,
        )
        
        if err != nil || userGroup == nil || userGroup.Role != "manager" {
            c.JSON(403, gin.H{
                "error": "forbidden",
                "required": "group_manager",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

---

## Authorization Decision Tree

```
Request → Check JWT Token
    ↓
    ├─ No Token? → 401 Unauthorized
    │
    ├─ Token Valid?
    │   ├─ No → 401 Unauthorized
    │   ├─ Yes → Extract Claims (userID, isAdmin)
    │
    ├─ Is Super Admin (isAdmin == true)?
    │   ├─ Yes → Allow (skip resource checks)
    │   ├─ No → Continue to resource level
    │
    ├─ Required Role Level?
    │   ├─ Admin → 403 Forbidden
    │   ├─ GroupAdmin → Check group membership + role
    │   ├─ GroupManager → Check group membership + role
    │   ├─ GroupMember → Check group membership
    │   ├─ UserOrAdmin → Check userID == paramID
    │
    ├─ Authorization Check Result?
    │   ├─ Allowed → Call Handler
    │   ├─ Denied → Log, Return 403 Forbidden
```

---

## Security Checklist

### For Every Protected Endpoint

- [ ] **Authorization Level Defined**
  - What role is required? (Admin, GroupManager, GroupMember, etc.)
  - Is it global or resource-level?

- [ ] **Resource Extraction Correct**
  - Can user ID be verified from path/body?
  - Is group/project ID correctly extracted?
  - Are parameters validated?

- [ ] **Error Handling Secure**
  - 401 Unauthorized (missing/invalid token)
  - 403 Forbidden (insufficient permission)
  - 404 Not Found (resource doesn't exist or not accessible)
  - Never leak internal details in error messages

- [ ] **Logging & Audit**
  - Authorization decisions logged
  - Failed attempts recorded
  - User ID and resource ID tracked
  - Timestamp captured

- [ ] **SQL Injection Prevention**
  - Use parameterized queries (GORM handles this)
  - Never interpolate user input into SQL

- [ ] **Token Validation**
  - JWT token signature verified
  - Expiration checked
  - Claims extracted safely
  - userID present and valid

---

## Migration Guide: Adding Authorization to Existing Endpoints

### Step 1: Identify Required Role
```
Question: What's the minimum role to access this endpoint?

├─ Super Admin Functions? → authMiddleware.Admin()
├─ Group-Level Operation? → authMiddleware.GroupAdmin()
├─ Project Management? → authMiddleware.GroupManager()
├─ User Participation? → authMiddleware.GroupMember()
└─ Self-Service? → authMiddleware.UserOrAdmin()
```

### Step 2: Add Authorization Middleware
```go
// Before
route.PUT("/:id", handler)

// After
route.PUT("/:id", 
    authMiddleware.GroupManager(
        middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
    ),
    handler,
)
```

### Step 3: Update Handler Error Handling
```go
func (h *Handler) UpdateProject(c *gin.Context) {
    // If we reach here, authorization already passed
    userID := c.GetUint("user_id")  // Guaranteed valid
    projectID := c.GetUint("id")     // Guaranteed valid
    
    // Continue with business logic
}
```

### Step 4: Test Authorization
```go
// Test: User without permission
func TestUpdateProjectUnauthorized(t *testing.T) {
    // Create project in group A
    // Login as member of group B
    // PUT /projects/{groupA_projectID}
    // Expect: 403 Forbidden
}

// Test: User with permission
func TestUpdateProjectAuthorized(t *testing.T) {
    // Create project in group A
    // Login as manager of group A
    // PUT /projects/{projectID}
    // Expect: 200 OK or 204 No Content
}
```

---

## Common Mistakes & Fixes

### Mistake 1: Forgetting Resource-Level Authorization
```go
// WRONG: Only checks if user exists
route.PUT("/:id", middleware.JWTAuthMiddleware(), handler)

// CORRECT: Verifies user can access this specific resource
route.PUT("/:id", 
    authMiddleware.GroupManager(
        middleware.FromIDParam(repos.Project.GetGroupIDByProjectID)
    ), 
    handler
)
```

### Mistake 2: Trusting User Input for Authorization
```go
// WRONG: User can modify their own request
projectID := c.Query("project_id")  // User can change this!

// CORRECT: Extract from URL path or verified source
projectID := c.GetUint("id")  // From router
groupID := repos.Project.GetGroupIDByProjectID(projectID)  // Verified
```

### Mistake 3: Missing Admin Override
```go
// WRONG: Admin can't access when checking strict membership
isManager := userGroup.Role == "manager"

// CORRECT: Admin bypasses all checks
if claims.IsAdmin || userGroup.Role == "manager" {
    c.Next()
}
```

### Mistake 4: Inconsistent Error Codes
```go
// WRONG: Leaks whether resource exists
if user.ID != userID {
    return 404  // Resource doesn't exist?
}

// CORRECT: Same error for missing vs. unauthorized
if user.ID != userID {
    return 403  // Forbidden (could mean not yours)
}
```

### Mistake 5: No Audit Logging
```go
// WRONG: Silent authorization failure
if !authorized {
    c.Abort()
}

// CORRECT: Log all decisions
logger.Warn("authorization failed",
    "user_id", userID,
    "resource", resourceID,
    "action", c.Request.Method,
    "required_role", "manager",
)
if !authorized {
    c.Abort()
}
```

---

## File Organization

Authorization-related code should be organized as follows:

```
internal/
├── api/
│   ├── middleware/
│   │   ├── jwt.go              # JWT validation, claims extraction
│   │   ├── auth.go             # AuthMiddleware implementation
│   │   └── extractors.go       # IDExtractor implementations
│   └── routes/
│       ├── router.go           # Main router setup
│       ├── audit.go            # Audit routes
│       ├── auth.go             # Auth routes (public)
│       ├── project.go          # Project routes with authorization
│       └── ...                 # Feature-specific routes
└── domain/
    └── user/
        └── role.go             # Role definitions, constants
```

---

## References

- **JWT Claims Structure**: Defined in `pkg/types/claims.go`
- **User Group Roles**: Defined in `internal/domain/group/user_group.go`
- **Repository Methods**: See `internal/repository/` for data access patterns
- **Middleware Implementation**: See `internal/api/middleware/` directory

---

## Future Enhancements

1. **Attribute-Based Access Control (ABAC)**
   - Beyond roles: attribute-based decisions (department, project status, etc.)
   - More flexible than pure RBAC

2. **Dynamic Permissions**
   - Time-limited access (expiring memberships)
   - Conditional permissions (based on resource state)

3. **Fine-Grained Permissions**
   - Per-field authorization (hide sensitive data)
   - Operation-level permissions (read-only vs. read-write)

4. **Access Audit Trail**
   - Detailed logging of all authorization decisions
   - Reporting on who accessed what and when

5. **Role Inheritance**
   - Hierarchical roles (manager inherits member permissions)
   - Cleaner permission model

---

## Approval & Version History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-02-03 | Platform Team | Initial creation: defined RBAC hierarchy, patterns, implementation guide |

---

**Status**: Active & Maintained  
**Last Updated**: February 3, 2026  
**Apply To**: All new endpoints requiring authorization from this date forward
