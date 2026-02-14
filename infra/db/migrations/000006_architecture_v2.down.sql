-- Revert changes
-- Note: Data migration reversal is destructive and approximate.

-- Drop indexes first
DROP INDEX IF EXISTS jobs_running_preemption_idx;
DROP INDEX IF EXISTS jobs_queue_idx;

-- Drop new tables
DROP TABLE IF EXISTS priority_classes CASCADE;
DROP TABLE IF EXISTS resource_plans CASCADE;
DROP TABLE IF EXISTS config_commits CASCADE;
DROP TABLE IF EXISTS config_blobs CASCADE;

-- Recreate old storage tables (empty)
CREATE TABLE IF NOT EXISTS user_storage (
  id VARCHAR(21) PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  user_id VARCHAR(21) NOT NULL REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
  pvc_name VARCHAR(100) NOT NULL,
  capacity INTEGER NOT NULL,
  storage_class VARCHAR(100) DEFAULT 'longhorn',
  created_by VARCHAR(21) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS group_storage (
  id VARCHAR(21) PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  group_id VARCHAR(21) NOT NULL REFERENCES group_list(g_id) ON DELETE CASCADE ON UPDATE CASCADE,
  pvc_name VARCHAR(100) NOT NULL,
  capacity INTEGER NOT NULL,
  storage_class VARCHAR(100) DEFAULT 'longhorn',
  created_by VARCHAR(21) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);

-- Drop unified storage
DROP TABLE IF EXISTS storages CASCADE;

-- Revert project_list changes
ALTER TABLE project_list
    DROP COLUMN IF EXISTS parent_id,
    DROP COLUMN IF EXISTS path,
    DROP COLUMN IF EXISTS owner_id;

-- Drop resource_owners (Cascade will remove FKs from users/group_list but keep data)
-- We need to drop constraints first to avoid deleting users
ALTER TABLE users DROP CONSTRAINT IF EXISTS fk_users_resource_owners;
ALTER TABLE group_list DROP CONSTRAINT IF EXISTS fk_groups_resource_owners;

DROP TABLE IF EXISTS resource_owners CASCADE;

-- Drop extensions
DROP EXTENSION IF EXISTS btree_gist;
DROP EXTENSION IF EXISTS ltree;
