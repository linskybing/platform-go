# Integration Test Plan — platform-go

> Date: 2026-02-13
> Currently 82 API Endpoints, 15 route modules, ~60 sub-tests covering 10 modules.

---

## 1. Overview of Existing Test Coverage

| Module | Test File | Covered Endpoints | Status |
|--------|-----------|-------------------|--------|
| Auth | `auth_handler_test.go` | `/login`, `/register`, `/logout`, `/auth/status` | Complete |
| Audit | `audit_handler_test.go` | `GET /audit/logs` | Complete |
| Group | `group_handler_test.go` | CRUD 5 endpoints | Complete |
| User Group | `user_group_handler_test.go` | CRUD 7 endpoints | Complete |
| ConfigFile | `configfile_handler_test.go` | CRUD + Instance 8 endpoints | Complete |
| Project | `project_handler_test.go` | CRUD 5 endpoints | Missing RBAC |
| User | `user_handler_test.go` | Get/Update/Delete | Missing List/Paging |
| Job (Plugin) | `job_handler_test.go` | Submit/List/Get/Cancel | Missing Templates/GPU |
| Storage Permission | `storage_permission_handler_test.go` | Set/Get/Batch/Policy | Complete |
| Repository | `repository_test.go` | Job repo, UserGroup, Storage preload | Basic |
| Redis Cache | `redis_cache_test.go` | Set/Get/TTL | Basic |
| Plugin System | `plugin_integration_test.go` | Init/Event bus | Basic |
| K8s Basic | `k8s_basic_test.go` | NS/Pod/PVC/Service CRUD | Complete (Requires cluster) |

### Modules Missing Coverage (Need New Tests)

| Module | Endpoints | Priority |
|--------|-----------|----------|
| Form / FormMessage | 7 | High |
| Image / Image Request | 7 | High |
| Cluster Summary | 4 | Medium |
| User Storage (Admin) | 4 | Medium |
| Group Storage | 6 | Medium |
| PVC Binding | 4 | Medium |
| K8s User Storage / FileBrowser | 5 | Medium |
| GPU Usage (Job Plugin) | 2 | Medium |
| Notification | 3 | Low (Stub) |
| WebSocket | 4 | Low (Special) |

---

## 2. Test Infrastructure (Existing)

```
test/integration/
├── setup_test.go        # TestContext (singleton), DB init, seed users/groups/project, JWT tokens
├── helpers.go           # TestDataGenerator, DatabaseCleaner, K8sResourceCleaner
├── http_client.go       # Fluent HTTP client (GET/POST/PUT/DELETE + JSON/Form body)
├── k8s_validator.go     # K8s resource validation helpers
├── k8stest/             # KinD test cluster utilities
└── scripts/             # kind-config.yaml
```

### Test Mode
- **Build tag**: `//go:build integration`
- **Shared setUp**: `sync.Once` initializes DB + Router + 3 JWT Tokens (admin/manager/user)
- **Cleanup**: `DatabaseCleaner` + `t.Cleanup()` deletes in dependency order
- **HTTP**: In-process `httptest.NewRecorder` (No real HTTP server)
- **K8s**: `K8S_MOCK=true` defaults to mock, real cluster requires manual enablement

---

## 3. New Test Plan

### Phase 1 — Fill Missing Coverage for Existing Modules (Week 1)

#### 1.1 `user_handler_test.go` Enhancement

```
TestUserHandler_ListAndPaging:
  ├── GET /users/           → List all users
  ├── GET /users/paging     → Paging query (page=1&size=10)
  ├── GET /users/paging     → Out of range page
  ├── GET /users/:id/settings → Get user settings
  └── PUT /users/:id/settings → Update user settings
```

#### 1.2 `project_handler_test.go` RBAC Enhancement

```
TestProjectHandler_RBAC:
  ├── POST /projects         → manager cannot create (403)
  ├── POST /projects         → user cannot create (403)
  ├── PUT /projects/:id      → user cannot update (403)
  ├── DELETE /projects/:id   → user cannot delete (403)
  ├── GET /projects/:id/config-files → Get project config files
  └── GET /projects/:id/images       → Get project allowed images
```

#### 1.3 `job_handler_test.go` Enhancement

```
TestJobHandler_Extended:
  ├── GET /jobs/templates           → List Job templates
  ├── GET /jobs/:id/gpu-usage       → Get GPU usage (empty result)
  ├── GET /jobs/:id/gpu-summary     → Get GPU summary (empty result)
  ├── POST /jobs/submit             → Invalid params (400)
  ├── POST /jobs/:id/cancel         → Cancel non-existent Job (404)
  └── GET /jobs?status=Running      → Filter query
```

---

### Phase 2 — New Module Tests (Week 2-3)

#### 2.1 `form_handler_test.go` (New File)

```go
// test/integration/form_handler_test.go
// //go:build integration

TestFormHandler_Integration:
  ├── [Create] POST /forms
  │   ├── Create form success (201)
  │   ├── Missing required fields (400)
  │   └── Verify user_id binding
  ├── [List My] GET /forms/my
  │   ├── Query own forms
  │   └── Different users only see their own forms
  ├── [List All] GET /forms
  │   └── List all forms
  ├── [Update Status] PUT /forms/:id/status
  │   ├── Update to Approved (200)
  │   ├── Update to invalid status (400)
  │   └── Non-existent form (404)
  ├── [Create Message] POST /forms/:id/messages
  │   ├── Add message success (201)
  │   └── Empty content (400)
  └── [List Messages] GET /forms/:id/messages
      ├── List messages (200)
      └── Non-existent form (404)
```

#### 2.2 `image_handler_test.go` (New File)

```go
// test/integration/image_handler_test.go
// //go:build integration

TestImageHandler_Integration:
  ├── [Admin] Image Requests
  │   ├── GET /image-requests          → List all (admin, 200)
  │   ├── GET /image-requests          → Non-admin (403)
  │   ├── GET /image-requests?status=  → Filter by status
  │   ├── PUT /image-requests/:id/approve  → Approve (200)
  │   ├── PUT /image-requests/:id/reject   → Reject (200)
  │   └── PUT /image-requests/:id/approve  → Non-existent request (404)
  │
  ├── [Project] Image Allow List
  │   ├── GET /projects/:id/images     → List project allowed images (200)
  │   ├── POST /projects/:id/images    → Add allowed image (201)
  │   ├── POST /projects/:id/images    → Duplicate add (409)
  │   ├── DELETE /projects/:id/images/:img_id → Remove image (204)
  │   └── GET /projects/:id/image-requests  → List project image requests
  │
  └── [RBAC]
      ├── POST /projects/:id/images    → user cannot add (403)
      └── DELETE /projects/:id/images  → user cannot delete (403)
```

#### 2.3 `cluster_handler_test.go` (New File)

```go
// test/integration/cluster_handler_test.go
// //go:build integration

TestClusterHandler_Integration:
  ├── GET /api/cluster/summary     → Get cluster summary (200)
  ├── GET /api/cluster/nodes       → List nodes (200)
  ├── GET /api/cluster/nodes/:name → Get single node (200)
  ├── GET /api/cluster/nodes/:name → Non-existent node (404)
  ├── GET /api/cluster/gpu-usage   → List GPU usage (200)
  └── Unauthenticated → All endpoints (401)
```

#### 2.4 `admin_storage_handler_test.go` (New File)

```go
// test/integration/admin_storage_handler_test.go
// //go:build integration

TestAdminStorageHandler_Integration:
  ├── [Admin User Storage]
  │   ├── GET /admin/user-storage/:username/status  → Query status (200)
  │   ├── POST /admin/user-storage/:username/init   → Init storage (201)
  │   ├── PUT /admin/user-storage/:username/expand   → Expand capacity (200)
  │   ├── DELETE /admin/user-storage/:username       → Delete storage (200)
  │   └── POST /admin/user-storage/:username/init   → Duplicate init (409)
  │
  └── [RBAC]
      ├── GET (user token) → Non-admin (403)
      ├── POST (manager token) → Non-admin (403)
      └── DELETE (user token) → Non-admin (403)
```

#### 2.5 `group_storage_handler_test.go` (New File)

```go
// test/integration/group_storage_handler_test.go
// //go:build integration

TestGroupStorageHandler_Integration:
  ├── [CRUD]
  │   ├── GET /storage/group/:id        → List group storage (200)
  │   ├── GET /storage/my-storages      → Query my group storage (200)
  │   ├── POST /storage/:id/storage     → Create group storage (201)
  │   ├── DELETE /storage/:id/storage/:pvcId → Delete (200)
  │   └── POST /storage/:id/storage     → Invalid params (400)
  │
  ├── [FileBrowser]
  │   ├── POST /storage/:id/storage/:pvcId/start  → Start FileBrowser
  │   └── DELETE /storage/:id/storage/:pvcId/stop  → Stop FileBrowser
  │
  └── [RBAC]
      ├── POST (user token) → Non-admin (403)
      └── DELETE (user token) → Non-admin (403)
```

#### 2.6 `pvc_binding_handler_test.go` (New File)

```go
// test/integration/pvc_binding_handler_test.go
// //go:build integration

TestPVCBindingHandler_Integration:
  ├── POST /k8s/pvc-binding                           → Create binding (201)
  ├── GET /k8s/pvc-binding/project/:project_id         → List bindings (200)
  ├── DELETE /k8s/pvc-binding/:binding_id              → Delete by ID (200)
  ├── DELETE /k8s/pvc-binding/project/:pid/:pvc_name   → Delete by name (200)
  ├── POST /k8s/pvc-binding                           → Invalid project (400)
  └── [RBAC] user cannot create (403)
```

---

### Phase 3 — K8s / WebSocket / Advanced Tests (Week 4)

#### 3.1 `k8s_user_storage_test.go` (New File — Requires K8s cluster)

```go
TestK8sUserStorage_Integration:
  ├── GET /k8s/user-storage/status  → Query storage status
  ├── POST /k8s/user-storage/browse → Open FileBrowser
  ├── DELETE /k8s/user-storage/browse → Stop FileBrowser
  └── GET /k8s/namespaces/:ns/pods/:name/logs → Get Pod logs
```

#### 3.2 `websocket_test.go` (New File)

```go
TestWebSocket_Integration:
  ├── WS /ws/exec             → Establish exec WebSocket connection
  ├── WS /ws/watch/:namespace  → Watch namespace events
  ├── WS /ws/pod-logs          → Stream pod logs
  └── WS /ws/job-status/:id    → Watch job status
```

> **Note**: WebSocket tests require `gorilla/websocket` + `httptest.NewServer` (Not `httptest.NewRecorder`)

#### 3.3 `notification_handler_test.go` (New File)

```go
TestNotificationHandler_Integration:
  ├── PUT /api/notifications/read-all     → Mark all read (200)
  ├── DELETE /api/notifications/clear-all → Clear all (200)
  ├── PUT /api/notifications/:id/read     → Mark single read (200)
  └── Unauthenticated → All endpoints (401)
```

> **Remark**: Notification handler is currently a stub, expected to return 200 OK but no actual DB operation

---

### Phase 4 — Cross-Functional Tests (Week 5)

#### 4.1 Cross-Module Integration Test `e2e_workflow_test.go`

```go
TestE2E_FullWorkflow:
  ├── [Step 1] Register new user → POST /register
  ├── [Step 2] Login to get Token → POST /login
  ├── [Step 3] Admin creates Group → POST /groups
  ├── [Step 4] Admin adds User to Group → POST /user-groups
  ├── [Step 5] Admin creates Project → POST /projects
  ├── [Step 6] Create ConfigFile → POST /configfiles
  ├── [Step 7] Deploy Instance → POST /instance/:id
  ├── [Step 8] Submit Job → POST /jobs/submit
  ├── [Step 9] Query Job Status → GET /jobs/:id
  ├── [Step 10] Cancel Job → POST /jobs/:id/cancel
  ├── [Step 11] Destroy Instance → DELETE /instance/:id
  ├── [Step 12] Delete ConfigFile → DELETE /configfiles/:id
  ├── [Step 13] Verify Audit Log → GET /audit/logs
  └── [Step 14] Logout → POST /logout
```

#### 4.2 Concurrent Access Test `concurrent_test.go`

```go
TestConcurrent_Operations:
  ├── Multiple users creating ConfigFiles simultaneously
  ├── Multiple users submitting Jobs simultaneously
  ├── Simultaneous Read + Write of group members
  └── User Storage initialization under race conditions
```

#### 4.3 Pagination and Filter Test `pagination_test.go`

```go
TestPagination_AllEndpoints:
  ├── GET /audit/logs?page=1&size=5
  ├── GET /users/paging?page=2&size=10
  ├── GET /jobs?page=1&size=20&status=Running
  └── Verify total count / has_more return
```

---

## 4. Execution Method

### Local Execution (Docker Compose)

```bash
# Start test dependencies
docker-compose -f docker-compose.integration.yml up -d

# Wait for PostgreSQL and Redis readiness
./scripts/run-integration-tests.sh db local

# Run all integration tests
go test -v -timeout 30m -tags=integration ./test/integration/...

# Run specific module
go test -v -timeout 10m -tags=integration ./test/integration/... -run TestFormHandler
go test -v -timeout 10m -tags=integration ./test/integration/... -run TestImageHandler
```

### Docker Execution

```bash
./scripts/run-integration-tests.sh db docker
```

### K8s Test (Requires KinD cluster)

```bash
./scripts/run-integration-tests.sh k8s docker
```

---

## 5. Naming Convention

```
TestXxxHandler_Integration        # Basic CRUD Integration Test
TestXxxHandler_RBAC               # Role/Permission Test
TestXxxHandler_Validation         # Input Validation Test
TestXxxHandler_EdgeCases          # Edge Case Test
TestE2E_xxx                       # End-to-End Flow Test
TestConcurrent_xxx                # Concurrency Test
```

---

## 6. New File List

| File | Module | Phase | Est. Sub-tests |
|------|--------|-------|----------------|
| `form_handler_test.go` | Form | 2 | 10 |
| `image_handler_test.go` | Image | 2 | 12 |
| `cluster_handler_test.go` | Cluster | 2 | 6 |
| `admin_storage_handler_test.go` | Admin Storage | 2 | 8 |
| `group_storage_handler_test.go` | Group Storage | 2 | 9 |
| `pvc_binding_handler_test.go` | PVC Binding | 2 | 6 |
| `k8s_user_storage_test.go` | K8s Storage | 3 | 4 |
| `websocket_test.go` | WebSocket | 3 | 4 |
| `notification_handler_test.go` | Notification | 3 | 4 |
| `e2e_workflow_test.go` | E2E | 4 | 14 |
| `concurrent_test.go` | Concurrent | 4 | 4 |
| `pagination_test.go` | Pagination | 4 | 4 |

**Total**: Adding 12 test files, ~85 new sub-tests, covering all 82 endpoints.

---

## 7. Expected Final Coverage

| Category | Current | After Completion |
|----------|---------|------------------|
| API Endpoint Coverage | ~50/82 (61%) | 82/82 (100%) |
| RBAC Tests | 4 Modules | All Modules |
| Error Paths | Partial | Comprehensive |
| E2E Flow | None | Complete User Flow |
| Concurrency Safety | None | Basic Coverage |
| WebSocket | None | Connection Verification |
