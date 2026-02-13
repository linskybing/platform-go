CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs (status);
CREATE INDEX IF NOT EXISTS idx_jobs_queue_name ON jobs (queue_name);
CREATE INDEX IF NOT EXISTS idx_jobs_project_status ON jobs (project_id, status);
CREATE INDEX IF NOT EXISTS idx_jobs_user_project_status ON jobs (user_id, project_id, status);

CREATE INDEX IF NOT EXISTS idx_forms_status ON forms (status);
CREATE INDEX IF NOT EXISTS idx_forms_tag ON forms (tag);

CREATE INDEX IF NOT EXISTS idx_project_list_name ON project_list (project_name);

CREATE INDEX IF NOT EXISTS idx_resources_name ON resources (name);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_created_at ON audit_logs (user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs (resource_type, resource_id);

CREATE INDEX IF NOT EXISTS idx_gsp_active_group_pvc ON group_storage_permissions (group_id, pvc_id) WHERE revoked_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_user_group_role ON user_group (role);

CREATE INDEX IF NOT EXISTS idx_users_email ON users (email);
