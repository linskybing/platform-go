-- Architecture V2 Schema Proposal
-- Extensions required for Hierarchy and Time-range constraints
CREATE EXTENSION IF NOT EXISTS ltree;
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- --------------------------------------------------------
-- 1. Polymorphic Resource Owners (Class Table Inheritance)
-- --------------------------------------------------------
CREATE TABLE resource_owners (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    kind VARCHAR(50) NOT NULL CHECK (kind IN ('user', 'group'))
);

CREATE TABLE users (
    id UUID PRIMARY KEY REFERENCES resource_owners(id) ON DELETE CASCADE,
    username VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE groups (
    id UUID PRIMARY KEY REFERENCES resource_owners(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT
);

-- --------------------------------------------------------
-- 2. Hierarchical Projects (Mixed Model: Adjacency + Materialized Path)
-- --------------------------------------------------------
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id UUID REFERENCES projects(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    -- ltree path for O(1) ancestor/descendant queries
    -- Format: root_uuid.child_uuid.grandchild_uuid (using replace(uuid, '-', '_'))
    path ltree NOT NULL,
    
    owner_id UUID REFERENCES resource_owners(id),

    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT unique_project_name_in_parent UNIQUE NULLS NOT DISTINCT (parent_id, name)
);

-- Index for efficient hierarchy traversal
CREATE INDEX project_path_gist_idx ON projects USING GIST (path);
CREATE INDEX project_parent_idx ON projects (parent_id);

-- --------------------------------------------------------
-- 3. Storage with Node Affinity (Inheritance)
-- --------------------------------------------------------
CREATE TABLE storages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES resource_owners(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    
    -- K8s PSS & Node Affinity Compliance
    -- Stores nodeSelector and affinity rules
    -- Example: {"nodeSelector": {"disktype": "ssd"}, "affinity": {...}}
    affinity_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    pvc_name VARCHAR(255),
    host_path VARCHAR(255),
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- --------------------------------------------------------
-- 4. Git-like Configuration Versioning
-- --------------------------------------------------------
-- Deduplicated Content Storage (Blobs)
CREATE TABLE config_blobs (
    hash CHAR(64) PRIMARY KEY, -- SHA-256
    content JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Commit History
CREATE TABLE config_commits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    blob_hash CHAR(64) NOT NULL REFERENCES config_blobs(hash),
    
    author_id UUID NOT NULL REFERENCES users(id),
    message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Optional: Tagging/Branching
    tag VARCHAR(50)
);

CREATE INDEX config_commits_project_idx ON config_commits (project_id, created_at DESC);

-- --------------------------------------------------------
-- 5. Resource Planning (Time-Window Constraints)
-- --------------------------------------------------------
-- Mapping: Mon 00:00:00 -> Sun 23:59:59 maps to integer 0 -> 604800
CREATE TABLE resource_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    resource_type VARCHAR(50) NOT NULL, -- e.g., 'nvidia.com/gpu'
    amount INT NOT NULL CHECK (amount > 0),
    
    -- Weekly recurring window (seconds from start of week)
    -- [start, end)
    week_window int4range NOT NULL,
    
    -- Constraint: No two plans for the same project and resource type can overlap in time
    -- This replaces application-level loops
    EXCLUDE USING GIST (
        project_id WITH =,
        resource_type WITH =,
        week_window WITH &&
    )
);

-- --------------------------------------------------------
-- 6. High-Concurrency Preemptive Job Queue
-- --------------------------------------------------------
CREATE TABLE priority_classes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) UNIQUE NOT NULL,
    value INT NOT NULL, -- Higher value = Higher priority
    preemption_policy VARCHAR(50) DEFAULT 'PreemptLowerPriority',
    global_default BOOLEAN DEFAULT FALSE
);

CREATE TYPE job_status AS ENUM ('pending', 'running', 'succeeded', 'failed', 'evicted');

CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id UUID NOT NULL REFERENCES projects(id),
    priority_class_id UUID REFERENCES priority_classes(id),
    
    -- Denormalized priority value for sorting speed
    priority_value INT NOT NULL DEFAULT 0,
    
    status job_status NOT NULL DEFAULT 'pending',
    
    required_gpu INT NOT NULL DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    started_at TIMESTAMPTZ
);

-- Index for Queue Fetching (SKIP LOCKED optimization)
CREATE INDEX jobs_queue_idx ON jobs (priority_value DESC, created_at ASC) 
WHERE status = 'pending';

-- Index for Preemption "Victim" Finding
CREATE INDEX jobs_running_preemption_idx ON jobs (priority_value ASC, created_at DESC) 
WHERE status = 'running';
