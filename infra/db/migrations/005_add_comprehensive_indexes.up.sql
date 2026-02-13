-- This migration is a placeholder to ensure indexes are present if they were missed in previous steps
-- The main schema.sql now includes all necessary indexes
-- Usage: migrate -path infra/db/migrations -database "postgresql://..." up

-- Add missing indexes if they don't exist
CREATE INDEX IF NOT EXISTS idx_config_files_project_id ON config_files(project_id);
CREATE INDEX IF NOT EXISTS idx_resources_cf_id ON resources(cf_id);
CREATE INDEX IF NOT EXISTS idx_jobs_user_id ON jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_jobs_project_id ON jobs(project_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_user_group_g_id ON user_group(g_id);
CREATE INDEX IF NOT EXISTS idx_image_allow_lists_check ON image_allow_lists(project_id, repository_id, is_enabled);
CREATE INDEX IF NOT EXISTS idx_container_repositories_full_name ON container_repositories(full_name);
CREATE INDEX IF NOT EXISTS idx_container_tags_name ON container_tags(name);
