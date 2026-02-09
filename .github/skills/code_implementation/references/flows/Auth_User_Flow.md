---
title: User Flow - Signup / Login / Access
description: Authentication flow and project access with validation, security considerations and tests.
---

# User Flow: Signup → Login → Access Project

## Endpoints
- POST `/api/v1/auth/signup` — create user
- POST `/api/v1/auth/login` — authenticate and return JWT
- GET `/api/v1/projects/:id` — retrieve project (requires auth)

## Steps (summary)
1. Signup: `internal/api/handlers/auth_handler.go:Signup` → `internal/application/user/service.go:RegisterUser` → `internal/repository/user.go:CreateUser`.
2. Login: `internal/api/handlers/auth_handler.go:Login` → `internal/application/user/service.go:Authenticate` → issue JWT (env secret) and optional refresh token.
3. Middleware: `internal/api/middleware/*` extracts and validates JWT or `X-API-Key`, attaches `user_id` to context.
4. Project access: `internal/api/handlers/project_handler.go:GetProjectByID` → `internal/application/project/service.go:GetProject` performs authorization checks.

## Security & Tests
- Use bcrypt for password hashing; rate-limit login endpoint; do not log tokens or passwords.
- Unit tests: table-driven tests for `Authenticate` mocking `UserRepo`.
- Integration tests: `test/integration/auth_test.go`.

## References
- `internal/api/handlers/auth_handler.go`
- `internal/application/user/service.go`
- `internal/repository/user.go`
- `internal/api/middleware/extractors.go`
