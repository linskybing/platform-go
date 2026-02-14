# Database Schema Design

The platform uses **PostgreSQL 13+** with advanced extensions (`ltree`, `btree_gist`, `uuid-ossp`) to support its cloud-native architecture.

## Extensions

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp"; -- For UUID generation
CREATE EXTENSION IF NOT EXISTS ltree;        -- For hierarchical structures
CREATE EXTENSION IF NOT EXISTS btree_gist;   -- For exclusion constraints
```

## Identity & Access (CTI Pattern)

### `resource_owners` (Supertype)
The base table for all identity entities.
| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | UUID | PK | Global Identifier |
| `owner_type` | VARCHAR | NOT NULL | Enum: 'USER', 'GROUP' |

### `users`
Inherits from `resource_owners`.
| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | UUID | PK, FK | References `resource_owners(id)` |
| `username` | VARCHAR | UNIQUE | Login handle |
| `email` | VARCHAR | UNIQUE | Contact info |
| `password_hash` | VARCHAR | | Bcrypt hash |

### `groups`
Inherits from `resource_owners`.
| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | UUID | PK, FK | References `resource_owners(id)` |
| `name` | VARCHAR | | Group display name |
| `parent_group_id` | UUID | FK | References `groups(id)` (Adjacency list) |

## Project Topology

### `projects`
Uses `ltree` for fast subtree queries.
| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `p_id` | UUID | PK | Project ID |
| `path` | LTREE | INDEX(GIST) | Materialized path (e.g. `root.group.project`) |
| `parent_id` | UUID | FK | References `projects(p_id)` |
| `owner_id` | UUID | FK | References `groups(id)` |

## Configuration Versioning

### `config_blobs`
Immutable content storage.
| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `hash` | CHAR(64) | PK | SHA-256 of content |
| `content` | JSONB | | The actual config data |

### `config_commits`
History tracking.
| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | UUID | PK | Commit ID |
| `project_id` | UUID | FK | References `projects(p_id)` |
| `blob_hash` | CHAR(64) | FK | References `config_blobs(hash)` |
| `author_id` | UUID | FK | References `users(id)` |

## Job Queue

### `jobs`
Optimized for `SKIP LOCKED` concurrency.
| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `id` | UUID | PK | Job ID |
| `status` | VARCHAR | INDEX | 'PENDING', 'RUNNING', etc. |
| `priority_value` | INT | INDEX | Higher runs first |
| `created_at` | TIMESTAMP | INDEX | FIFO tie-breaker |

## Resource Planning

### `resource_plans`
Enforces schedule validity.
| Column | Type | Constraints | Description |
| :--- | :--- | :--- | :--- |
| `project_id` | UUID | FK | References `projects(p_id)` |
| `week_window` | INT4RANGE | | Seconds [0, 604800) |
| `gpu_limit` | INT | | Max GPUs allowed in this window |

**Constraint**: `EXCLUDE USING GIST (project_id WITH =, week_window WITH &&)` prevents overlapping schedules for the same project.
