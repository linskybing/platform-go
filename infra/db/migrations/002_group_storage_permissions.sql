-- ================================================
-- Group Storage Permission Management System
-- Created: 2026-02-02
-- Purpose: Implement Linux-like permission system for group storage
-- ================================================

-- Table: group_storage_permissions
-- Stores user permissions for specific group PVCs
CREATE TABLE IF NOT EXISTS group_storage_permissions (
    id SERIAL PRIMARY KEY,
    group_id INTEGER NOT NULL,
    pvc_id VARCHAR(100) NOT NULL,
    pvc_name VARCHAR(100) NOT NULL,
    user_id INTEGER NOT NULL,
    permission VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (permission IN ('none', 'read', 'write')),
    granted_by INTEGER NOT NULL,
    granted_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at TIMESTAMP NULL,
    
    -- Indexes for performance
    CONSTRAINT idx_group_pvc_user UNIQUE (group_id, pvc_id, user_id),
    CONSTRAINT fk_group FOREIGN KEY (group_id) REFERENCES group_list(g_id) ON DELETE CASCADE,
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES user_list(u_id) ON DELETE CASCADE,
    CONSTRAINT fk_granted_by FOREIGN KEY (granted_by) REFERENCES user_list(u_id)
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
    group_id INTEGER NOT NULL,
    pvc_id VARCHAR(100) NOT NULL UNIQUE,
    default_permission VARCHAR(20) NOT NULL DEFAULT 'none' CHECK (default_permission IN ('none', 'read', 'write')),
    admin_only BOOLEAN NOT NULL DEFAULT FALSE,
    created_by INTEGER NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_policy_group FOREIGN KEY (group_id) REFERENCES group_list(g_id) ON DELETE CASCADE,
    CONSTRAINT fk_policy_creator FOREIGN KEY (created_by) REFERENCES user_list(u_id)
);

CREATE INDEX IF NOT EXISTS idx_gsap_group_id ON group_storage_access_policies(group_id);
CREATE INDEX IF NOT EXISTS idx_gsap_pvc_id ON group_storage_access_policies(pvc_id);

COMMENT ON TABLE group_storage_access_policies IS 'Default access policies for group storage PVCs';
COMMENT ON COLUMN group_storage_access_policies.default_permission IS 'Default permission for new group members';
COMMENT ON COLUMN group_storage_access_policies.admin_only IS 'If true, only group admins can access';

-- ================================================

-- Table: project_pvc_bindings
-- Stores PVC bindings from project namespaces to group storage
CREATE TABLE IF NOT EXISTS project_pvc_bindings (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    group_pvc_id VARCHAR(100) NOT NULL,
    project_pvc_name VARCHAR(100) NOT NULL UNIQUE,
    project_namespace VARCHAR(100) NOT NULL,
    source_pv_name VARCHAR(200) NOT NULL,
    access_mode VARCHAR(50) NOT NULL DEFAULT 'ReadOnlyMany',
    status VARCHAR(50) NOT NULL DEFAULT 'Pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_binding_user FOREIGN KEY (user_id) REFERENCES user_list(u_id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_ppb_project_id ON project_pvc_bindings(project_id);
CREATE INDEX IF NOT EXISTS idx_ppb_user_id ON project_pvc_bindings(user_id);
CREATE INDEX IF NOT EXISTS idx_ppb_group_pvc_id ON project_pvc_bindings(group_pvc_id);
CREATE INDEX IF NOT EXISTS idx_ppb_namespace ON project_pvc_bindings(project_namespace);
CREATE UNIQUE INDEX IF NOT EXISTS idx_ppb_unique_pvc ON project_pvc_bindings(project_namespace, project_pvc_name);

COMMENT ON TABLE project_pvc_bindings IS 'PVC bindings allowing projects to mount group storage';
COMMENT ON COLUMN project_pvc_bindings.source_pv_name IS 'PV name to bind to (from group storage)';
COMMENT ON COLUMN project_pvc_bindings.access_mode IS 'ReadOnlyMany or ReadWriteMany based on permission';

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

CREATE TRIGGER trg_update_ppb_updated_at
    BEFORE UPDATE ON project_pvc_bindings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ================================================

-- Function: Cleanup permissions when user leaves group
CREATE OR REPLACE FUNCTION cleanup_permissions_on_group_leave()
RETURNS TRIGGER AS $$
BEGIN
    -- Revoke all storage permissions for this user in this group
    UPDATE group_storage_permissions
    SET revoked_at = CURRENT_TIMESTAMP
    WHERE user_id = OLD.u_id
      AND group_id = OLD.g_id
      AND revoked_at IS NULL;
    
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

-- Trigger: Auto-cleanup when user is removed from group
CREATE TRIGGER trg_cleanup_permissions_on_leave
    BEFORE DELETE ON user_group
    FOR EACH ROW
    EXECUTE FUNCTION cleanup_permissions_on_group_leave();

-- ================================================

-- Example Data (for testing)
-- COMMENT OUT OR REMOVE IN PRODUCTION

-- Example: Set read-write permission for user 1 on group 1's PVC
-- INSERT INTO group_storage_permissions (group_id, pvc_id, pvc_name, user_id, permission, granted_by)
-- VALUES (1, 'group-1-abc123', 'pvc-abc123', 1, 'write', 1);

-- Example: Set default read permission for group 1's PVC
-- INSERT INTO group_storage_access_policies (group_id, pvc_id, default_permission, admin_only, created_by)
-- VALUES (1, 'group-1-abc123', 'read', FALSE, 1);

-- Example: Create project PVC binding
-- INSERT INTO project_pvc_bindings (project_id, user_id, group_pvc_id, project_pvc_name, project_namespace, source_pv_name, access_mode, status)
-- VALUES (1, 1, 'group-1-abc123', 'shared-storage', 'project-1', 'pvc-abc123-pv', 'ReadWriteMany', 'Bound');

-- ================================================
-- Migration Complete
-- ================================================
