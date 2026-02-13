# Implementation Plan: Integration Tests & Optimization

## 1. Unit Tests (Phase 1)
- [ ] **Events**: Create `internal/events/bus_test.go`
  - Test `MemoryEventBus` (Publish, Subscribe, Unsubscribe)
  - Verify async handling and error safety
- [ ] **Repository**: Create `internal/repository/job_test.go`
  - Use SQLite in-memory for fast feedback
  - Test JobRepoImpl CRUD methods
- [ ] **WebSocket**: Enhance `internal/api/handlers/websocket/ws_job_test.go`
  - Increase coverage > 90%
  - Test edge cases (missing ID, not found, status updates)

## 2. Infrastructure Setup (Phase 2)
- [ ] **Docker**: Start Postgres (5433) and Redis (6380)
- [ ] **Test Setup**: Update `test/integration/setup_test.go`
  - Register `JobPlugin` in test environment
  - Add Redis client support

## 3. Integration Tests (Phase 3)
- [ ] **Job Handler**: Create `test/integration/job_handler_test.go`
  - Test API endpoints: List, Get, Submit, Cancel
- [ ] **Repositories**: Create `test/integration/repository_test.go`
  - Test `JobRepo` with real Postgres
  - Test `UserGroupRepo` and `StorageRepo` Preload (N+1 fixes)
- [ ] **Plugin System**: Create `test/integration/plugin_integration_test.go`
  - Test Manager Init/Shutdown with real DB
  - Test EventBus with real handlers
- [ ] **Redis Cache**: Create `test/integration/redis_cache_test.go`
  - Test Set/Get/Expiry against real Redis

## 4. Execution & Optimization (Phase 4)
- [ ] Run Integration Tests: `tags=integration` with real DB/Redis
- [ ] Run Benchmarks: CPU/Memory profiling
- [ ] Identify Bottlenecks (EventBus, Repos, WebSocket)
- [ ] Implement Optimizations
