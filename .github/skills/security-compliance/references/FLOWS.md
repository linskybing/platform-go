---
title: Security & Compliance Flows
description: Authentication, API key lifecycle, RBAC and secure initialization flows.
---

# Security Flows

## S1. API Key Generation & Revocation Flow

1. Admin calls API to create an API key (endpoint under admin/service routes). Handler calls `CreateAPIKey` service.
2. Service: `security-compliance` pattern — generate random key, hash with sha256, store `KeyHash` and expiry in DB (`internal/repository/apikeys.go`).
3. Return plain key only once to caller; subsequent lookups use hashed value.
4. Revoke: mark `ExpiresAt` or delete record; invalidate caches referencing API key.

Reference files:
- `internal/repository` for API key storage, `security-compliance/SKILL.md` guidance

## S2. JWT Issuance & Rotation Flow

1. On login, service signs JWT with current signing key and returns token.
2. Use `kid` in token header to identify signing key; store active keys in config or KMS.
3. Rotation: add new key, start issuing tokens with new `kid`, keep old key to validate until tokens expire, then retire old key.

Reference files:
- `internal/application/user/service.go` (token issuance), `internal/config` for secret management

## S3. Safe Auto-Migration & Default Admin Init

1. On startup, run GORM auto-migrate for domain models (see `internal/config/db/database.go`).
2. Check if admin exists; if not, create using `ADMIN_PASSWORD` env var. Fail startup if `ADMIN_PASSWORD` unset in sensitive environments.

Reference files:
- `internal/config/db/database.go`, `security-compliance/SKILL.md`
