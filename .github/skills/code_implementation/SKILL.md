---
name: code_implementation
description: Consolidated skills for platform-go: architecture, code-quality, documentation, operations, security-compliance
license: Proprietary
metadata:
  author: platform-go
  version: "1.0"
  consolidated_from:
    - architecture
    - code-quality
    - documentation
    - operations
    - security-compliance
---

# Platform-Go Consolidated Skills

This file consolidates the key backend-design and integration guidance for platform-go to serve as a single reference for backend implementers and integrators.

## 1. Architecture

- System layering: API Handlers → Services/Application → Domain Models → Repository → External Services (K8s, MinIO).
- RESTful patterns, standard response envelope, and explicit status codes (200/201/204/400/401/403/404/409/500).
- Database design: primary keys (ID/UUID), created_at/updated_at, indexes for hot queries, foreign keys, GORM migration and transaction patterns.
- Scalability: stateless services, horizontal scaling, Redis for sessions/cache, CDN for static assets, resource and connection pooling.
- Security basics: env var secrets, API key rotation, bcrypt for passwords, HTTPS in production.
- Workflow & scheduler APIs: endpoints for workflow submission/lifecycle, queue management, WebSocket for real-time progress, integration notes for Argo/Volcano/K8s.

## 2. Code Quality

- Clean architecture & single responsibility per package; prefer small files (target < 200 lines).
- Error handling: wrap errors with context, propagate `context.Context` as first parameter.
- Testing: table-driven unit tests, integration tests for external services, target >= 70% coverage.
- Concurrency & performance: use semaphores or worker pools to limit concurrency; preallocate slices and use connection pooling.
- Database best practices: use transactions for multi-step operations, paginate large queries, avoid loading all records.
- Validation checklist: no unused imports, errors wrapped, file size limits, and coverage verification.

## 3. Documentation

- Project docs layout under `docs/` with README, ARCHITECTURE, API, INSTALLATION, DEPLOYMENT, TROUBLESHOOTING, CONTRIBUTING, CHANGELOG, and guides.
- Markdown standards: lowercase hyphen filenames, heading hierarchy, code fences by language, tables for config and env vars.
- README template: brief description, prerequisites (Go 1.20+, Docker, Postgres), quick start commands, env setup, build/run.
- API docs: base URL, auth examples (JWT Bearer and `X-API-Key`), endpoint request/response examples, pagination and query param specs.
- Architecture docs: component diagrams, data flow, responsibilities per layer.

## 4. Operations

- CI/CD: GitHub Actions example (test → build → push → deploy), cache Go modules, parallel jobs, fail-fast strategy.
- Kubernetes: Deployment pattern with resource requests/limits, rolling update strategy, liveness/readiness probes, secrets via `secretKeyRef`.
- client-go best practices: use context with timeouts, list with selectors, wait loops for readiness with bounded retries.
- Redis caching: `GetOrFetch` pattern, TTL policies, explicit invalidation on write operations.
- Monitoring: structured logging with fields, Prometheus metrics for latency and error counts, health (`/health`) and readiness (`/ready`) endpoints.
- Workflow & scheduler ops: deployment and integration tips for Argo/Volcano, scripts for K8s integration tests (Kind), and operational runbooks.

## 5. Security & Compliance

- Authentication: unified middleware supporting JWT (web) and `X-API-Key` (service), validate and attach user to context.
- Authorization: RBAC mapping of roles → permissions; middleware `RequirePermission` and resource-level checks.
- API Key lifecycle: generate secure random key, hash before storage (sha256), store `KeyHash`, return plain key once at creation, support expiry.
- Secure coding: input binding/validation tags, bcrypt password hashing (cost ~12), parameterized queries with GORM, never log secrets.
- Initialization: safe auto-migration and default admin creation guarded by environment variables (e.g., `ADMIN_PASSWORD`).
- Security checklist: ensure bcrypt for passwords, hashed API keys, HTTPS in prod, rate limiting, CORS restrictions, CSRF for stateful ops, security headers, and scheduled audits.

## Usage & Next Steps

- This consolidated `SKILL.md` is intended as a quick reference for backend engineers and integrators to know what to implement and which contracts to follow.
- If you want these split into separate skill files per the agentskills spec (one file per skill), I can create them individually and wire validation scripts.

---

See consolidated implementation flows in `references/INDEX.md` for step-by-step sequences and file mappings.

## Implementation Details (References)

This section maps the consolidated guidance to the actual implementation locations and concrete function names in the repository. Use these references when integrating clients or implementing new backend features.

### API Layer (Handlers & Routes)
- Route registration: `internal/api/routes/` — routes group endpoints and attach handlers (e.g., `internal/api/routes/project.go`).
- Handlers: `internal/api/handlers/` contains HTTP handlers by domain. Examples:
  - Projects: `internal/api/handlers/project_handler.go` — `ProjectHandler.GetProjects`, `GetProjectsByUser`, `GetProjectByID`, `CreateProject`.
  - Users: `internal/api/handlers/user_handler.go` — `GetUsers`, `GetUserByID`.
  - ConfigFiles: `internal/api/handlers/configfile_handler.go` — `GetConfigFileHandler`, `CreateConfigFileHandler`, `CreateInstanceHandler`.
  - PVC binding: `internal/api/handlers/pvc_binding_handler.go` — `CreateBinding`.

### Application / Service Layer
- Services live under `internal/application/` and encapsulate business logic.
  - Projects: `internal/application/project/service.go` — `ProjectService.GetProject`, `GetProjectsByUser`, `CreateProject`.
  - Users: `internal/application/user/service.go` — `UserService.RegisterUser`, etc.
  - ConfigFile lifecycle & deploy: `internal/application/configfile/service.go` and `deploy.go` — `CreateConfigFile`, `CreateInstance` and helper functions that bind volumes and create K8s objects.
  - K8s / Storage managers: `internal/application/k8s/` — `K8sService`, `StorageManager`, `PVCBindingManager`, and `FileBrowserManager` with methods like `CreateGroupPVC`, `CreateProjectPVCBinding`, `GetFileBrowserAccess`.

### Domain Models & DTOs
- Domain structs and validation live in `internal/domain/`:
  - `internal/domain/project/model.go` (type `Project`) and `dto.go` (e.g., `CreateProjectDTO`).
  - `internal/domain/user/model.go`, `internal/domain/form/model.go` and `dto.go` files include validation hooks (`BeforeCreate`).
  - Storage models: `internal/domain/storage/*` (e.g., `GroupPVC`, `StorageHub`, `GroupStoragePermission`).

### Repository / Persistence
- Repositories are under `internal/repository/`. Typical patterns:
  - `internal/repository/project.go` — `GetProjectByID`, `CreateProject`.
  - `internal/repository/user.go`, `internal/repository/storage_permission.go`, etc., each exposing DB CRUD used by services.
  - Database initialization & views: `internal/config/db/database.go` (enum creation, views, migrations helpers).
  - SQL schema and migration files: `infra/db/schema.sql` and `infra/db/migrations/`.

### Cache Layer
- Redis helper and caching utilities: `pkg/cache/`.
  - `pkg/cache/get_or_fetch.go` provides `GetOrFetchJSON` pattern used by services for read-through caching.
  - Tests: `pkg/cache/redis_cache_test.go` demonstrates usage and singleflight patterns.

### Kubernetes Integration
- K8s helper utilities: `pkg/k8s/` (client, helpers, namespace, volume utilities) and `internal/application/k8s/` (managers/services performing higher-level operations).
  - `pkg/k8s/client.go` contains functions such as `CreateJob`, `CreateFileBrowserPod`, `CreateFileBrowserService`.
  - Volume/PVC helpers: `pkg/k8s/volumn.go`, `internal/application/k8s/storage_pvc.go` and `pvc_binding_manager.go` implement binding logic.

### MinIO / Filebrowser
- Filebrowser helpers and session management: `pkg/filebrowser/manager.go`, `pkg/filebrowser/session.go`.
- Utilities to create file browser pods and services: `pkg/utils/filebrowser_session.go`.

### Authentication & Middleware
- Middleware and extractors: `internal/api/middleware/` including `extractors.go` used to load project/context from requests.
- Token/API key handling in middleware lives with auth helpers under `pkg/utils` and the `security-compliance` guidance above.

### Configuration
- Central configuration: `internal/config/config.go` — environment variables and config parsing helpers (see `getEnv`).

### Tests & Integration
- Integration tests: `test/integration/` includes HTTP client helpers and integration cases for projects and K8s fixtures.
- Unit tests for services and application logic under `internal/application/*_test.go`.

### Scripts, CI, and Deployment
- Dockerfiles: `Dockerfile`, `Dockerfile.integration`.
- CI scripts and helper scripts under `scripts/` and `./.github/scripts/` used by build pipelines.
- Kubernetes manifests: `k8s/` (manifests for api, postgres, redis, secrets).

## Implementation Contracts (Endpoints → Service → Repo)
Below are representative contracts you can rely on when integrating clients.

- Create Project
  - Endpoint: POST `/api/v1/projects` (registered in `internal/api/routes/project.go`)
  - Handler: `internal/api/handlers/project_handler.go:CreateProject`
  - Service: `internal/application/project/service.go:CreateProject(c, input project.CreateProjectDTO)` — validates DTO, creates `project.Project` model, calls `Repos.Project.CreateProject`.
  - Repository: `internal/repository/project.go:CreateProject` — persists project, returns `PID`.

- Get Project
  - Endpoint: GET `/api/v1/projects/:id`
  - Handler: `ProjectHandler.GetProjectByID` → Service `ProjectService.GetProject(id)` → Repo `GetProjectByID`.

- Create ConfigFile Instance (deploy)
  - Endpoint: POST `/api/v1/configfiles/{id}/instances` (see configfile routes)
  - Handler: `internal/api/handlers/configfile_handler.go:CreateInstanceHandler`
  - Service: `internal/application/configfile/deploy.go:CreateInstance` — resolves project, binds volumes, creates K8s resources via `pkg/k8s` and `internal/application/k8s` managers.

## DB Schema Pointers
- Schema DDL: `infra/db/schema.sql` — canonical table structures.
- GORM model hooks (timestamps, BeforeCreate hooks) are implemented in domain model files (e.g., `internal/domain/project/model.go`, `internal/domain/user/model.go`).

## Observability & Health
- Health endpoints: `internal/application/...` or handlers expose `/health` and `/ready` patterns (see `pkg` and `internal` monitoring code examples). Use these for readiness/liveness probes in Kubernetes manifests.

## Practical Integration Checklist (for backend clients)
- Auth: support `Authorization: Bearer <JWT>` and `X-API-Key` headers; validate per middleware rules in `internal/api/middleware`.
- Pagination: use query params `page`, `limit` and respect server-side max `limit` (handlers already parse these patterns in list handlers).
- Error handling: expect JSON envelope with `success` boolean and `error` object; map HTTP status codes to client errors.
- Caching: when calling list or expensive endpoints, use TTL and invalidate on write operations (see `pkg/cache` usage patterns and `internal/application/*` invalidation calls).
- K8s interactions: services that create pods/PVCs rely on `pkg/k8s` APIs; provide proper timeouts and retries.

---

If you'd like, I can now:
- (A) Split this consolidated `SKILL.md` into separate skill files under `.github/skills-consolidated/` and add lightweight validation scripts; or
- (B) Generate a machine-readable agentskills YAML/JSON following https://agentskills.io/specification for each skill using the repository references above.

Tell me which of (A) or (B) you prefer, or request specific sections to expand further.
