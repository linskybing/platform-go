-- ================================================
-- Group Storage Permission Management System
-- Created: 2026-02-02
-- Purpose: Implement Linux-like permission system for group storage
-- Modified: 2026-02-13
-- Changes: Use string IDs, fix FK references, remove unused project_pvc_bindings
-- ================================================

-- Table: group_storage_permissions
-- Stores user permissions for specific group PVCs
CREATE TABLE IF NOT EXISTS group_storage_permissions (
    id SERIAL PRIMARY KEY,
    group_id VARCHAR(21) NOT NULL,
    pvc_id VARCHAR(21) NOT NULL,
    pvc_name VARCHAR(100) NOT NULL,
    user_id VARCHAR(21) NOT NULL,
    permission VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (permission IN ('none', 'read', 'write')),
    granted_by VARCHAR(21) NOT NULL,
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP NULL,
    
    -- Indexes for performance
    CONSTRAINT idx_group_pvc_user UNIQUE (group_id, pvc_id, user_id),
    CONSTRAINT fk_group FOREIGN KEY (group_id) REFERENCES group_list(g_id) ON DELETE CASCADE,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(u_id) ON DELETE CASCADE,
    CONSTRAINT fk_granted_by FOREIGN KEY (granted_by) REFERENCES users(u_id)
);

CREATE INDEX IF NOT EXISTS idx_gsp_group_id ON group_storage_permissions(group_id);
CREATE INDEX IF NOT EXISTS idx_gsp_pvc_id ON group_storage_permissions(pvc_id);
CREATE INDEX IF NOT EXISTS idx_gsp_user_id ON group_storage_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_gsp_pvc_name ON group_storage_permissions(pvc_name);
CREATE INDEX IF NOT EXISTS idx_gsp_revoked_at ON group_storage_permissions(revoked_at);

COMMENT ON TABLE group_storage_permissions IS 'User-level permissions for group storage PVCs';
COMMENT ON COLUMN group_storage_permissions.permission IS 'Permission level: none (no access), read (read-only), write (read-write)';
COMMENT ON COLUMN group_storage_permissions.revoked_at IS 'NULL if active, timestamp if revoked';

-- ================================================

-- Table: group_storage_access_policies
-- Stores default access policies for group PVCs
CREATE TABLE IF NOT EXISTS group_storage_access_policies (
    id SERIAL PRIMARY KEY,
    group_id VARCHAR(21) NOT NULL,
    pvc_id VARCHAR(21) NOT NULL UNIQUE,
    default_permission VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (default_permission IN ('none', 'read', 'write')),
    admin_only BOOLEAN NOT NULL DEFAULT FALSE,
    created_by VARCHAR(21) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_policy_group FOREIGN KEY (group_id) REFERENCES group_list(g_id) ON DELETE CASCADE,
    CONSTRAINT fk_policy_creator FOREIGN KEY (created_by) REFERENCES users(u_id)
);

CREATE INDEX IF NOT EXISTS idx_gsap_group_id ON group_storage_access_policies(group_id);
CREATE INDEX IF NOT EXISTS idx_gsap_pvc_id ON group_storage_access_policies(pvc_id);

COMMENT ON TABLE group_storage_access_policies IS 'Default access policies for group storage PVCs';
COMMENT ON COLUMN group_storage_access_policies.default_permission IS 'Default permission for new group members';
COMMENT ON COLUMN group_storage_access_policies.admin_only IS 'If true, only group admins can access';

-- ================================================

-- Function: Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for auto-updating updated_at
CREATE TRIGGER trg_update_gsp_updated_at
    BEFORE UPDATE ON group_storage_permissions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_update_gsap_updated_at
    BEFORE UPDATE ON group_storage_access_policies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
