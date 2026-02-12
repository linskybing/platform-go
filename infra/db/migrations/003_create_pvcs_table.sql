-- 003_create_pvcs_table.sql

-- Create the ENUM type for PVC type
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'pvc_type') THEN
        CREATE TYPE pvc_type AS ENUM ('user', 'group');
    END IF;
END$$;

-- Create the pvcs table
CREATE TABLE IF NOT EXISTS pvcs (
    id VARCHAR(21) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    owner_id VARCHAR(21) NOT NULL,
    type pvc_type NOT NULL,
    namespace VARCHAR(100) NOT NULL,
    pvc_name VARCHAR(100) NOT NULL,
    capacity INT NOT NULL,
    storage_class VARCHAR(100) DEFAULT 'longhorn',
    access_mode VARCHAR(50) DEFAULT 'ReadWriteMany',
    status VARCHAR(50) DEFAULT 'Pending',
    created_by VARCHAR(21) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_created_by_user
        FOREIGN KEY(created_by) 
        REFERENCES users(u_id)
        ON DELETE RESTRICT,

    UNIQUE(namespace, pvc_name)
);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_pvcs_owner_id_type ON pvcs(owner_id, type);
CREATE INDEX IF NOT EXISTS idx_pvcs_name ON pvcs(name);
CREATE INDEX IF NOT EXISTS idx_pvcs_created_at ON pvcs(created_at);

-- Drop old tables if they exist
DROP TABLE IF EXISTS group_pvcs;
DROP TABLE IF EXISTS persistent_volume_claims;
