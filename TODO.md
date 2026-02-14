# Refactoring Plan: Multi-Tenant Cloud-Native Architecture

This document outlines the roadmap for refactoring the platform into a high-concurrency, multi-tenant system with advanced scheduling capabilities using PostgreSQL and Kubernetes best practices.

## Phase 1: Database Architecture & Schema Design (Foundational)

The database is the source of truth and the enforcement engine for concurrency control and data integrity.

- [ ] **1.1. Extensions & Primitives**
    - Enable `ltree` for hierarchical data structures.
    - Enable `btree_gist` for exclusion constraints.
    - Define custom time range types if necessary (or utilize `int4range` mapped to seconds-of-week).

- [ ] **1.2. Global Identity & Polymorphic Storage (Class Table Inheritance)**
    - Create `resource_owners` (Supertype) table.
    - Refactor `users` to inherit/FK from `resource_owners`.
    - Refactor `groups` to inherit/FK from `resource_owners`.
    - Create `storages` table referencing `resource_owners(id)` (Single FK, no `owner_type`).
    - Add JSONB `affinity_config` to `storages` for K8s PSS compliance.

- [ ] **1.3. Hierarchical Project Topology**
    - Create `projects` table with:
        - `id` (UUID).
        - `parent_id` (FK to self) for referential integrity.
        - `path` (ltree) for efficient subtree queries.
        - GiST index on `path`.

- [ ] **1.4. Git-like Configuration Versioning**
    - Create `config_blobs` table:
        - `hash` (PK, computed SHA-256 of content).
        - `content` (JSONB), de-duplicated.
    - Create `config_commits` table:
        - `project_id` (FK).
        - `blob_hash` (FK).
        - `created_at`, `author_id`, `message`.
        - `version_tag` (optional).

- [ ] **1.5. Resource Planning with Time Constraints**
    - Create `resource_plans` table.
    - Columns: `resource_owner_id`, `resource_type` (GPU/CPU), `quantity`.
    - **Constraint:** Use `int4range` (0-604800 seconds) representing the recurring weekly window.
    - **Enforcement:** Add `EXCLUDE USING GIST` constraint on `(resource_owner_id, time_window WITH &&)` to prevent overlapping allocations at the DB level.

- [ ] **1.6. Job Queue & Preemption Schema**
    - Create `priority_classes` table (mapping K8s `PriorityClass`).
        - `value` (int), `preemption_policy` (enum), `name`.
    - Create `jobs` table.
        - `priority` (int), `status` (pending/running/evicted), `required_gpu`.
    - **Optimization:** Indexes optimized for `ORDER BY priority DESC, created_at ASC`.

## Phase 2: Domain Layer Implementation (Go)

- [ ] **2.1. Entity Refactoring**
    - Update Go structs in `internal/domain` to match the new schema.
    - Implement `Scanner` and `Valuer` interfaces for custom DB types (`ltree`, `range`).

- [ ] **2.2. Topology Service**
    - Implement tree operations: `MoveSubtree`, `GetAncestors`, `GetDescendants` using `ltree` operators (`@>`, `<@`).
    - Ensure RBAC permission checks traverse the `ltree` path.

- [ ] **2.3. Config Versioning Service**
    - Implement "Content-Addressable Storage" logic in Go.
    - On Save: Compute Hash -> Check existence in `config_blobs` -> Insert if missing -> Create Commit.

## Phase 3: Scheduling Engine & Concurrency (Core Logic)

- [ ] **3.1. Time Window Validation**
    - Remove any application-level loop checks for time overlaps.
    - Rely on Postgres `PQ` error handling (Code `23P01` exclusion_violation) to detect conflicts.

- [ ] **3.2. High-Concurrency Queue Worker**
    - Implement `FetchNextJob` using `SELECT ... FOR UPDATE SKIP LOCKED`.
    - Ensure atomic state transitions.

- [ ] **3.3. Preemption Logic (The "Victim" Finder)**
    - Implement the Preemption Service.
    - **SQL Logic:** Construct a query using CTEs and Window Functions (`SUM(gpu) OVER (ORDER BY priority ASC)`) to identify lower-priority running jobs that must be evicted to satisfy a high-priority pending job.
    - Trigger K8s Eviction API based on query results.

## Phase 4: Testing & Verification

- [ ] **4.1. Unit Tests**
    - Test `ltree` generation and parsing.
    - Test Content-Addressing hashing logic.
    - Test Preemption Algorithm (mocking the DB or using an in-memory SQL driver).

- [ ] **4.2. Integration Tests**
    - Spin up PostgreSQL with `docker-compose.integration.yml`.
    - Verify `EXCLUDE` constraints actually block overlapping time ranges.
    - Verify `SKIP LOCKED` allows parallel workers to pick different jobs.
    - Verify `ltree` queries return correct subtrees.

## Phase 5: Cleanup & Documentation

- [ ] **5.1. API Docs**
    - Regenerate API documentation based on new architecture.
- [ ] **5.2. Final Code Audit**
    - Ensure no "N+1" queries in the hierarchy traversal.
