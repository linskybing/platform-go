-- Enable extensions
CREATE EXTENSION IF NOT EXISTS ltree;
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- -----------------------------------------------------------------------------
-- 1. Polymorphic Resource Owners
-- -----------------------------------------------------------------------------

-- Create the supertype table
CREATE TABLE resource_owners (
    id VARCHAR(21) PRIMARY KEY, -- Matches existing NanoID format
    kind VARCHAR(50) NOT NULL CHECK (kind IN ('user', 'group')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Migrate existing users to resource_owners
INSERT INTO resource_owners (id, kind, created_at)
SELECT u_id, 'user', create_at FROM users;

-- Migrate existing groups to resource_owners
INSERT INTO resource_owners (id, kind, created_at)
SELECT g_id, 'group', create_at FROM group_list;

-- Add FK constraints to existing tables to enforce inheritance
ALTER TABLE users
    ADD CONSTRAINT fk_users_resource_owners
    FOREIGN KEY (u_id) REFERENCES resource_owners(id) ON DELETE CASCADE;

ALTER TABLE group_list
    ADD CONSTRAINT fk_groups_resource_owners
    FOREIGN KEY (g_id) REFERENCES resource_owners(id) ON DELETE CASCADE;

-- -----------------------------------------------------------------------------
-- 2. Hierarchical Projects
-- -----------------------------------------------------------------------------

-- We will enhance the existing project_list table to support hierarchy
-- First, rename it to 'projects' for cleaner nomenclature
ALTER TABLE project_list RENAME TO projects;

ALTER TABLE projects
    ADD COLUMN parent_id VARCHAR(21) REFERENCES projects(p_id) ON DELETE CASCADE,
    ADD COLUMN path ltree,
    ADD COLUMN owner_id VARCHAR(21) REFERENCES resource_owners(id);

-- Initialize path for existing projects (flat hierarchy)
-- We treat them as root nodes for now. Path = p_id (sanitized)
UPDATE projects
SET path = text2ltree(replace(p_id, '-', '_'));

ALTER TABLE projects
    ALTER COLUMN path SET NOT NULL;

CREATE INDEX project_path_gist_idx ON projects USING GIST (path);
CREATE INDEX project_parent_idx ON projects (parent_id);

-- -----------------------------------------------------------------------------
-- 3. Unified Storage & Node Affinity
-- -----------------------------------------------------------------------------

CREATE TABLE storages (
    id VARCHAR(21) PRIMARY KEY,
    owner_id VARCHAR(21) NOT NULL REFERENCES resource_owners(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    
    -- PSS Compliance
    affinity_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    
    pvc_name VARCHAR(255),
    host_path VARCHAR(255),
    
    capacity INTEGER NOT NULL,
    storage_class VARCHAR(100),
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Migrate user_storage
INSERT INTO storages (id, owner_id, name, pvc_name, capacity, storage_class, created_at, updated_at)
SELECT id, user_id, name, pvc_name, capacity, storage_class, created_at, updated_at
FROM user_storage;

-- Migrate group_storage
INSERT INTO storages (id, owner_id, name, pvc_name, capacity, storage_class, created_at, updated_at)
SELECT id, group_id, name, pvc_name, capacity, storage_class, created_at, updated_at
FROM group_storage;

-- Drop old storage tables
DROP TABLE user_storage;
DROP TABLE group_storage;

-- -----------------------------------------------------------------------------
-- 4. Git-like Config Versioning
-- -----------------------------------------------------------------------------

CREATE TABLE config_blobs (
    hash CHAR(64) PRIMARY KEY,
    content JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE config_commits (
    id VARCHAR(21) PRIMARY KEY,
    project_id VARCHAR(21) NOT NULL REFERENCES project_list(p_id) ON DELETE CASCADE,
    blob_hash CHAR(64) NOT NULL REFERENCES config_blobs(hash),
    
    author_id VARCHAR(21) NOT NULL REFERENCES resource_owners(id), -- Could be user
    message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    tag VARCHAR(50)
);

CREATE INDEX config_commits_project_idx ON config_commits (project_id, created_at DESC);

-- -----------------------------------------------------------------------------
-- 5. Resource Plans (Time Windows)
-- -----------------------------------------------------------------------------

CREATE TABLE resource_plans (
    id VARCHAR(21) PRIMARY KEY,
    project_id VARCHAR(21) NOT NULL REFERENCES project_list(p_id) ON DELETE CASCADE,
    
    resource_type VARCHAR(50) NOT NULL,
    amount INT NOT NULL CHECK (amount > 0),
    
    -- Weekly recurring window (seconds 0-604800)
    week_window int4range NOT NULL,
    
    EXCLUDE USING GIST (
        project_id WITH =,
        resource_type WITH =,
        week_window WITH &&
    )
);

-- -----------------------------------------------------------------------------
-- 6. Priority Classes & Preemptive Jobs
-- -----------------------------------------------------------------------------

CREATE TABLE priority_classes (
    id VARCHAR(21) PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    value INT NOT NULL,
    preemption_policy VARCHAR(50) DEFAULT 'PreemptLowerPriority',
    global_default BOOLEAN DEFAULT FALSE
);

-- Enhance jobs table
ALTER TABLE jobs
    ADD COLUMN priority_class_id VARCHAR(21) REFERENCES priority_classes(id),
    ADD COLUMN priority_value INT NOT NULL DEFAULT 0,
    ADD COLUMN required_gpu INT NOT NULL DEFAULT 0,
    ADD COLUMN config_commit_id VARCHAR(21) REFERENCES config_commits(id);

-- Indexes for Queue
CREATE INDEX jobs_queue_idx ON jobs (priority_value DESC, created_at ASC) 
WHERE status = 'Pending';

CREATE INDEX jobs_running_preemption_idx ON jobs (priority_value ASC, created_at DESC) 
WHERE status = 'Running';
