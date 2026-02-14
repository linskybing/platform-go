# System Architecture

This document describes the architectural decisions and patterns used in the platform's backend.

## 1. Domain-Driven Design (DDD) with Class Table Inheritance

To solve the complexity of polymorphic relationships (e.g., a `Storage` volume can be owned by a `User` or a `Group`, logs belong to various entities), we utilize **Class Table Inheritance (CTI)**.

### The `resource_owners` Supertype

Instead of using multiple foreign keys or loose string-based polymorphic associations (`owner_type`, `owner_id`), we define a base table:

```sql
CREATE TABLE resource_owners (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_type VARCHAR(50) NOT NULL, -- 'USER' or 'GROUP'
    created_at TIMESTAMP
);
```

Both `users` and `groups` tables inherit from this:

- **Users Table**: `id` is a Foreign Key to `resource_owners(id)`.
- **Groups Table**: `id` is a Foreign Key to `resource_owners(id)`.

**Benefits:**
- **Referential Integrity**: A `storage` record references `resource_owners(id)`. The database guarantees that the ID exists, whether it's a user or a group.
- **Unified Auditing**: Audit logs can simply reference `resource_owner_id` to track who performed an action or who owns a resource.

## 2. Hierarchical Topology with `ltree`

The platform supports a deeply nested hierarchy of Groups, Projects, and Sub-Projects. Traditional Adjacency Lists (parent_id) make subtree queries (e.g., "Find all projects under Group A") slow and recursive.

We use the PostgreSQL **`ltree`** extension to store the materialized path of each node.

### Schema
```sql
CREATE TABLE projects (
    p_id UUID PRIMARY KEY,
    path ltree NOT NULL, -- e.g., "root_uuid.group_uuid.project_uuid"
    ...
);
CREATE INDEX idx_project_path ON projects USING GIST (path);
```

### Operations
- **Find Descendants**: `WHERE path <@ 'root.group.project'` (Very fast with GIST index)
- **Find Ancestors**: `WHERE path @> 'root.group.project'`
- **Move Subtree**: When a project is moved, we calculate the new path and update all descendants in a single transaction:
  ```sql
  UPDATE projects SET path = new_path || subpath(path, nlevel(old_path)) WHERE path <@ old_path
  ```

## 3. High-Concurrency Job Queue

The platform handles job submissions that need to be processed by workers (e.g., K8s deployers, image builders). To avoid race conditions and locking contention, we use **PostgreSQL as a Queue**.

### The `SKIP LOCKED` Pattern

Workers fetch jobs using this atomic query:

```sql
SELECT * FROM jobs
WHERE status = 'PENDING'
ORDER BY priority_value DESC, created_at ASC
LIMIT 1
FOR UPDATE SKIP LOCKED;
```

**How it works:**
1.  **FOR UPDATE**: Locks the row so no other transaction can modify it.
2.  **SKIP LOCKED**: If a row is already locked by another worker, the database *skips* it immediately instead of waiting.
3.  **Result**: Multiple workers can run this query simultaneously, and each will receive a *unique* job to process, maximizing throughput without external message brokers like RabbitMQ or Kafka.

## 4. Git-like Configuration Versioning (CAS)

We treat configuration files (YAML/JSON) as immutable data, similar to Git.

### Content-Addressable Storage (CAS)
1.  **Blobs**: The content of a file is hashed (SHA-256). We store the content in a `config_blobs` table keyed by this hash.
    -   *Deduplication*: If two projects use the exact same config, we only store the blob once.
2.  **Commits**: A `config_commits` table links a `project_id` to a `blob_hash` with metadata (author, message, timestamp).

This structure allows for instant rollbacks, audit trails, and efficient storage.

## 5. Resource Planning with Exclusion Constraints

To prevent over-allocation of resources (GPUs, CPU) during specific time windows, we use PostgreSQL **Exclusion Constraints**.

### `int4range` Time Windows
A weekly schedule is mapped to seconds (0 to 604800).

```sql
CREATE TABLE resource_plans (
    project_id UUID,
    week_window int4range, -- e.g., '[3600, 7200)' (1am to 2am on Monday)
    gpu_limit INT,
    EXCLUDE USING GIST (
        project_id WITH =,
        week_window WITH &&
    )
);
```

**Enforcement:**
The database creates a GiST index that strictly prevents any two rows for the same project from having overlapping time ranges. Application logic does not need to loop or verify overlaps; the database guarantees correctness.
