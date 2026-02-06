# Authentication Header Strategy - Implementation Summary

**Date**: January 20, 2024  
**Status**: Design Complete - Ready for Implementation  
**Author**: Platform-Go Team

---

## Quick Reference

### Header Schema

```
JWT Authentication (Browser/Mobile):
  ┌─────────────────────────────────────────┐
  │ Authorization: Bearer <jwt-token>       │
  │ Example: Bearer eyJhbGc...             │
  └─────────────────────────────────────────┘

API Key Authentication (Service/CLI):
  ┌─────────────────────────────────────────┐
  │ X-API-Key: pk_<random-32-bytes>        │
  │ X-API-Key-Name: ci-pipeline (optional) │
  │ Example: pk_a1b2c3d4e5f6g7h8i9j0...   │
  └─────────────────────────────────────────┘

Zero Conflicts ✓
Separate Headers ✓
Clear Intent ✓
```

---

## Implementation Tasks

### 1. Middleware Implementation
**File**: `internal/api/middleware/auth.go`

- [ ] `JWTAuthMiddleware()` - Extract from `Authorization: Bearer`
- [ ] `APIKeyAuthMiddleware(db)` - Extract from `X-API-Key`
- [ ] `OptionalAuthMiddleware(db)` - Try both, pass if either valid
- [ ] `tryJWT()` - JWT validation helper
- [ ] `tryAPIKey()` - API key validation helper

**Lines of Code**: ~250-300 lines  
**Dependencies**: JWT parser, GORM, Gin

### 2. Route Configuration
**File**: `internal/api/routes/routes.go`

- [ ] Register public routes (no auth)
- [ ] Register JWT-required routes
- [ ] Register API-Key-required routes  
- [ ] Register optional-auth routes

**Example**:
```go
// JWT-protected
jwtRequired := router.Group("/api")
jwtRequired.Use(middleware.JWTAuthMiddleware())

// API-Key-protected
serviceRequired := router.Group("/api/v1")
serviceRequired.Use(middleware.APIKeyAuthMiddleware(db))

// Either/Or
anyAuth := router.Group("/api")
anyAuth.Use(middleware.OptionalAuthMiddleware(db))
```

### 3. Testing
**File**: `test/integration/auth_test.go`

- [ ] Test JWT authentication
- [ ] Test API Key authentication
- [ ] Test both headers simultaneously
- [ ] Test header conflicts (should have none)
- [ ] Test missing headers
- [ ] Test expired tokens/keys

**Test Coverage**: ~10-15 test functions

### 4. Documentation Update
**Files**: 
- `docs/API_STANDARDS.md` - Add header reference
- `README.md` - Quick start guide

- [ ] Header format documentation
- [ ] Client integration examples
- [ ] cURL command examples
- [ ] SDK usage patterns

---

## Validation Checklist

### Functional Requirements
- ✅ JWT uses `Authorization: Bearer` header
- ✅ API Key uses `X-API-Key` header
- ✅ `X-API-Key-Name` header is optional
- ✅ No header conflicts (separate headers)
- ✅ Both headers can be present simultaneously
- ✅ At least one auth method required for protected routes

### Non-Functional Requirements
- ✅ Secure header transmission (HTTPS only)
- ✅ No plaintext keys in logs
- ✅ Audit trail for all auth operations
- ✅ Performance (minimal overhead)
- ✅ Backwards compatible (if needed)

### Security Requirements
- ✅ Header size validation
- ✅ Rate limiting support
- ✅ Expiration validation
- ✅ Format validation (pk_ prefix for API keys)
- ✅ Token/key revocation support

---

## File Structure

```
internal/api/
├── middleware/
│   └── auth.go                    ← JWT + API Key middleware
├── routes/
│   └── routes.go                  ← Route registration
└── handlers/
    ├── auth.go                    ← Login, logout, refresh
    └── user.go                    ← User-related endpoints

internal/domain/
└── user/
    └── model.go                   ← User model (has API key fields)

docs/
├── SEPARATE_HEADERS_IMPLEMENTATION.md  ← Implementation guide
└── API_STANDARDS.md               ← API header documentation

test/integration/
└── auth_test.go                   ← Auth tests
```

---

## Configuration

### User Model (internal/domain/user/model.go)

Required fields:
```go
APIKeyHash          string      // Bcrypt hash
APIKeyName          string      // Friendly name
APIKeyLastUsed      *time.Time  // Audit trail
APIKeyCreatedAt     time.Time
APIKeyExpiresAt     *time.Time  // Optional
```

### Constants (internal/constants/constants.go)

```go
const (
    APIKeyPrefix    = "pk_"
    APIKeyLength    = 32  // bytes
    APIKeyHashCost  = 12  // bcrypt
)
```

---

## Request/Response Examples

### JWT Request
```http
GET /api/users/me HTTP/1.1
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response**:
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": "user123",
    "username": "john_doe"
  }
}
```

### API Key Request
```http
POST /api/v1/webhooks/events HTTP/1.1
X-API-Key: pk_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
X-API-Key-Name: ci-pipeline
Content-Type: application/json

{"event": "build.complete"}
```

**Response**:
```json
{
  "code": 200,
  "message": "Event processed",
  "data": {
    "event_id": "evt_123"
  }
}
```

### Error Cases

**Missing JWT**:
```json
{
  "code": 401,
  "message": "Missing JWT token",
  "error": "unauthorized"
}
```

**Expired API Key**:
```json
{
  "code": 401,
  "message": "API key expired",
  "error": "unauthorized"
}
```

---

## Context Injection

Both middleware inject into Gin context:

```go
// From JWT
c.Get("user_id")      // string
c.Get("username")     // string
c.Get("is_admin")     // bool
c.Get("auth_type")    // "jwt"

// From API Key
c.Get("user_id")      // string
c.Get("username")     // string
c.Get("is_admin")     // false (for API keys)
c.Get("auth_type")    // "api_key"
c.Get("api_key_name") // string or nil
```

**Handler Usage**:
```go
func GetUser(c *gin.Context) {
    userID, _ := c.Get("user_id")
    authType, _ := c.Get("auth_type")
    
    if authType == "jwt" {
        // Handle JWT-authenticated request
    } else if authType == "api_key" {
        // Handle API-key-authenticated request
    }
}
```

---

## Testing Strategy

### Unit Tests
```bash
go test ./internal/api/middleware -run TestJWTAuth
go test ./internal/api/middleware -run TestAPIKeyAuth
```

### Integration Tests
```bash
# Test JWT flow
go test ./test/integration -run TestJWTFlow

# Test API Key flow
go test ./test/integration -run TestAPIKeyFlow

# Test both headers
go test ./test/integration -run TestBothHeaders

# Test all auth scenarios
go test ./test/integration -run TestAuth
```

### Manual Testing
```bash
# Generate test JWT
JWT_TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}' \
  | jq -r '.data.token')

# Test JWT
curl -H "Authorization: Bearer $JWT_TOKEN" \
  http://localhost:8080/api/users/me

# Test API Key
curl -H "X-API-Key: pk_abc123..." \
  http://localhost:8080/api/v1/webhooks/status

# Test both
curl -H "Authorization: Bearer $JWT_TOKEN" \
  -H "X-API-Key: pk_abc123..." \
  http://localhost:8080/api/admin/audit
```

---

## Comparison: Before vs After

### Before (Shared Header with Token Detection)
```
Authorization: Bearer <token-or-key>

Problem:
- Need to detect token type by format (ey... vs pk_...)
- Ambiguous intent
- Error handling complex
- Format detection logic required
```

### After (Separate Headers)
```
Authorization: Bearer <jwt-token>
X-API-Key: pk_<key>

Benefits:
✓ Clear intent - no detection needed
✓ Separate header namespaces
✓ Simpler implementation
✓ Easier debugging
✓ Supports both simultaneously
✓ No header conflicts possible
```

---

## Performance Impact

- **JWT Middleware**: ~0.5ms per request (JWT parsing)
- **API Key Middleware**: ~1-2ms per request (DB lookup)
- **Optional Auth Middleware**: ~2-3ms per request (tries both)

**Optimization**: Use caching for API keys with Redis if needed.

---

## Migration Path (if applicable)

For existing systems using old auth:

1. **Phase 1**: Deploy new middleware alongside old
2. **Phase 2**: Route new clients to new middleware
3. **Phase 3**: Log usage of old vs new
4. **Phase 4**: Deprecate old middleware with warnings
5. **Phase 5**: Remove old middleware after grace period

---

## Support & Troubleshooting

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| 401 Missing JWT | No Authorization header | Add `Authorization: Bearer <token>` |
| 401 Invalid JWT | Token expired/malformed | Regenerate token via login endpoint |
| 401 Missing API Key | No X-API-Key header | Add `X-API-Key: pk_...` header |
| 401 Expired API Key | Key past expiration date | Regenerate or rotate API key |
| Header conflict errors | Should not occur | No conflicts with separate headers |

### Debug Commands

```bash
# Check header is being sent
curl -v -H "Authorization: Bearer $JWT" http://localhost:8080/api/users/me

# Check API key format
curl -v -H "X-API-Key: pk_abc123..." http://localhost:8080/api/v1/webhooks/status

# View request headers in logs
export DEBUG=1 && curl -H "Authorization: Bearer $JWT" http://localhost:8080/api/users/me
```

---

## Related Skills

- `unified-authentication-strategy/SKILL.md` - Complete design documentation
- `automigration-apikey-initialization/SKILL.md` - Database setup
- `security-best-practices/SKILL.md` - Security guidelines

---

## Next Steps

1. ✅ Design complete (this document)
2. ⏳ Implement middleware (internal/api/middleware/auth.go)
3. ⏳ Update routes (internal/api/routes/routes.go)
4. ⏳ Add tests (test/integration/auth_test.go)
5. ⏳ Update documentation (docs/)
6. ⏳ Client SDKs update (if applicable)
7. ⏳ Deploy to production

---

## Questions & Support

For questions about this authentication strategy:
1. Check `docs/SEPARATE_HEADERS_IMPLEMENTATION.md` for detailed guide
2. Review `unified-authentication-strategy/SKILL.md` for design rationale
3. Check test files for usage examples
4. Consult security team for security requirements

---

**Status**: ✅ Design Complete - Ready for Development Team

