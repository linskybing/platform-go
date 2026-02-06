---
name: access-control-best-practices
description: Role-based access control (RBAC) patterns, authorization middleware, permission hierarchy, and security implementation for platform-go API endpoints
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
---


# Access Control & Authorization Best Practices

**Version**: 1.0  
**Date**: February 3, 2026  
**Status**: Active


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


**Status**: Active & Maintained  
**Last Updated**: February 3, 2026  
**Apply To**: All new endpoints requiring authorization from this date forward
