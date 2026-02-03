# Access Control & Authorization Skills

**Latest Version**: 1.0  
**Status**: Active  
**Last Updated**: February 3, 2026

## Overview

This skills package provides comprehensive guidance for implementing role-based access control (RBAC) and authorization throughout the platform-go project. It establishes the permission hierarchy, middleware patterns, and implementation best practices for protecting endpoints.

### Key Documents

1. **[SKILL.md](SKILL.md)** - Complete authorization framework
   - 5-tier permission hierarchy (Super Admin → Group Admin → Manager → Member → User)
   - Authorization middleware patterns and extractors
   - Common authorization patterns with code examples
   - Security checklist and migration guide
   - Common mistakes and how to fix them

2. **[IMPLEMENTATION_CHECKLIST.md](IMPLEMENTATION_CHECKLIST.md)** - Step-by-step implementation guide
   - Pre-implementation review
   - Implementation phase checklist
   - Authorization logic verification
   - Testing strategy (unit and integration tests)
   - Code review checklist
   - Deployment verification

## Quick Start

### For New Endpoint Protection

1. **Identify the required role**:
   - System admin? → `authMiddleware.Admin()`
   - Group operations? → `authMiddleware.GroupAdmin(extractor)`
   - Project management? → `authMiddleware.GroupManager(extractor)`
   - User submissions? → `authMiddleware.GroupMember(extractor)`
   - Self-service? → `authMiddleware.UserOrAdmin()`

2. **Add middleware to route**:
   ```go
   route.PUT("/:id",
       authMiddleware.GroupManager(
           middleware.FromIDParam(repos.Project.GetGroupIDByProjectID),
       ),
       handler,
   )
   ```

3. **Follow [IMPLEMENTATION_CHECKLIST.md](IMPLEMENTATION_CHECKLIST.md)** for complete implementation

### For Understanding Current Architecture

Refer to [SKILL.md](SKILL.md) sections:
- [Authorization Hierarchy](#authorization-hierarchy) - Role definitions
- [Resource-Level Authorization](#resource-level-authorization) - Extractors and patterns
- [Implementation Examples](#implementation-examples) - Real code patterns
- [Authorization Decision Tree](#authorization-decision-tree) - How decisions are made

## Permission Hierarchy

```
┌─────────────────────────────────────────┐
│         Super Admin (Highest)           │
│  - System configuration                 │
│  - All projects, all groups             │
│  - Unlimited access                     │
└────────────┬────────────────────────────┘
             │
┌────────────┴────────────────────────────┐
│      Group Admin (Per-Group Level)      │
│  - Manage projects in group             │
│  - Manage group members                 │
│  - Approve permissions                  │
└────────────┬────────────────────────────┘
             │
┌────────────┴────────────────────────────┐
│     Group Manager (Per-Project Level)   │
│  - Update project config                │
│  - Manage project members               │
│  - Create images/forms                  │
└────────────┬────────────────────────────┘
             │
┌────────────┴────────────────────────────┐
│    Group Member (Workspace Level)       │
│  - Submit jobs/forms                    │
│  - View resources                       │
│  - Create instances                     │
└──────────────────────────────────────────┘
```

## Middleware Patterns Quick Reference

| Pattern | Usage | Example |
|---------|-------|---------|
| `Admin()` | Super admin only | System config, global operations |
| `GroupAdmin(extractor)` | Group-level destructive ops | Delete project, change group policies |
| `GroupManager(extractor)` | Project management | Update project, add images |
| `GroupMember(extractor)` | User submissions | Create jobs, submit forms |
| `UserOrAdmin()` | Self-service with admin override | User profile updates |

## Authorization Decision Flow

Every protected endpoint follows this flow:

1. **JWT Validation** - Token present and valid?
2. **Super Admin Check** - Is user super admin? → Allow (skip remaining checks)
3. **Role Verification** - Does user have required role?
4. **Resource Verification** - Does user have access to this resource?
5. **Decision** - Allow or Deny with appropriate status code

## File Organization

```
internal/
├── api/
│   ├── middleware/
│   │   ├── jwt.go          # JWT token validation
│   │   ├── auth.go         # Authorization middleware
│   │   └── extractors.go   # Resource ID extractors
│   └── routes/
│       ├── router.go       # Main router setup
│       ├── admin.go        # Admin routes
│       ├── project.go      # Project routes (authorization examples)
│       └── ...
└── domain/
    └── user/
        └── role.go         # Role definitions
```

## Implementation Rules

### DO ✅

- ✅ Use consistent middleware across similar endpoints
- ✅ Add authorization middleware to all protected endpoints
- ✅ Verify resource extraction is correct
- ✅ Test with users having different permission levels
- ✅ Log authorization decisions for audit
- ✅ Use standard HTTP status codes (401, 403, 404)
- ✅ Let middleware handle authorization, not the handler

### DON'T ❌

- ❌ Trust user-provided IDs without verification
- ❌ Re-check authorization in the handler
- ❌ Use only authentication without authorization
- ❌ Mix multiple authorization patterns on same endpoint
- ❌ Leak information in error messages
- ❌ Bypass middleware for "fast" endpoints

## Testing Strategy

Every protected endpoint should have tests for:

1. **Authentication Tests**
   - [ ] No token → 401 Unauthorized
   - [ ] Invalid token → 401 Unauthorized

2. **Authorization Tests**
   - [ ] Insufficient role → 403 Forbidden
   - [ ] Wrong resource group → 403 Forbidden
   - [ ] Sufficient role → Success (200-299)
   - [ ] Admin user → Success (200-299)

3. **Edge Cases**
   - [ ] Non-existent resource → 404 Not Found
   - [ ] Malformed input → 400 Bad Request
   - [ ] Database error → 500 Internal Server Error

See [IMPLEMENTATION_CHECKLIST.md](IMPLEMENTATION_CHECKLIST.md) for detailed test examples.

## Common Mistakes & Fixes

| Mistake | Problem | Solution |
|---------|---------|----------|
| Missing authorization middleware | Anyone can access endpoint | Add `authMiddleware.RoleLevel(extractor)` |
| No resource extraction | Can't verify resource ownership | Add appropriate extractor |
| Trusting user input | User can modify their own IDs | Extract ID from URL, verify in DB |
| Wrong HTTP status | Confusing error messages | 401 (auth), 403 (authz), 404 (not found) |
| No logging | Can't audit access | Log authorization decisions |

For more details, see [SKILL.md - Common Mistakes & Fixes](SKILL.md#common-mistakes--fixes)

## Integration with Other Skills

This skill relates to and should be used in conjunction with:

- **[api-design-patterns](../api-design-patterns/SKILL.md)** - RESTful endpoint design
- **[golang-production-standards](../golang-production-standards/SKILL.md)** - Error handling and logging
- **[security-best-practices](../security-best-practices/SKILL.md)** - General security guidelines
- **[testing-best-practices](../testing-best-practices/SKILL.md)** - Authorization testing strategies

## Usage in Code Reviews

When reviewing code with authorization requirements:

1. **Check middleware presence** - Is endpoint protected? ([SKILL.md](SKILL.md#authorization-hierarchy))
2. **Verify role level** - Is correct role required? ([SKILL.md](SKILL.md#common-authorization-patterns))
3. **Validate resource extraction** - Is ID extracted correctly? ([SKILL.md](SKILL.md#resource-level-authorization))
4. **Review error handling** - Are status codes appropriate? ([SKILL.md](SKILL.md#authorization-decision-tree))
5. **Check testing** - Are there authorization tests? ([IMPLEMENTATION_CHECKLIST.md](IMPLEMENTATION_CHECKLIST.md#testing-phase))

## Current Implementation Reference

The following files demonstrate this skill in practice:

- **Router setup**: `internal/api/routes/router.go` - Authorization middleware initialization
- **Middleware**: `internal/api/middleware/auth.go` - Authorization implementation
- **Extractors**: `internal/api/middleware/extractors.go` - Resource ID extraction
- **Example routes**: 
  - `internal/api/routes/project.go` - Group-level authorization
  - `internal/api/routes/admin.go` - Admin-only endpoints

## Checklist for New Developers

Starting a new endpoint requiring authorization?

1. [ ] Read [SKILL.md - Authorization Hierarchy](SKILL.md#authorization-hierarchy)
2. [ ] Identify which role should access this endpoint
3. [ ] Look at similar endpoint in current code for pattern
4. [ ] Follow [IMPLEMENTATION_CHECKLIST.md](IMPLEMENTATION_CHECKLIST.md)
5. [ ] Write unit tests for all role levels
6. [ ] Get code review using [IMPLEMENTATION_CHECKLIST.md - Code Review Checklist](IMPLEMENTATION_CHECKLIST.md#code-review-checklist)

## FAQ

**Q: When should I use `GroupAdmin` vs `GroupManager`?**  
A: Use `GroupAdmin` for destructive operations (delete, revoke), `GroupManager` for update/configuration operations. See [SKILL.md - Common Authorization Patterns](SKILL.md#common-authorization-patterns).

**Q: How do I extract the resource ID for a nested route?**  
A: Use appropriate extractor: `FromIDParam` for URL, `FromProjectIDInPayload` for JSON body. See [SKILL.md - Authorization Extractors](SKILL.md#authorization-extractors).

**Q: Can an admin bypass these checks?**  
A: Yes, but only super admin. See [SKILL.md - Implementation Examples - Example 3](SKILL.md#example-3-complex-authorization-logic) for how admin bypass is implemented.

**Q: What if I need custom authorization logic?**  
A: Create custom middleware following the pattern in [SKILL.md - Example 3](SKILL.md#example-3-complex-authorization-logic).

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-02-03 | Initial release: RBAC hierarchy, middleware patterns, implementation guide, checklists |

## Maintenance

This skill is maintained as part of the platform-go project. For updates or improvements:

1. Review [SKILL.md](SKILL.md) for completeness
2. Check [IMPLEMENTATION_CHECKLIST.md](IMPLEMENTATION_CHECKLIST.md) for accuracy
3. Ensure all examples compile and work
4. Update version history when changes made

---

**Status**: ✅ Active  
**Last Reviewed**: February 3, 2026  
**Next Review**: As needed when new authorization patterns emerge
