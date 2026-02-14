-- Create Database
-- CREATE DATABASE platform;
-- \c platform;

-- Extensions
CREATE EXTENSION IF NOT EXISTS ltree;
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- ENUMs
DO $$ BEGIN
    CREATE TYPE resource_type AS ENUM ('Pod','Service','Deployment','ConfigMap','Ingress','Job');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE user_type AS ENUM ('origin','oauth2');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE user_status AS ENUM ('online','offline','delete');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('admin','manager','user');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

-- Polymorphic resource owners
CREATE TABLE IF NOT EXISTS resource_owners (
    id VARCHAR(21) PRIMARY KEY,
    kind VARCHAR(50) NOT NULL CHECK (kind IN ('user', 'group')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- groups
CREATE TABLE IF NOT EXISTS group_list (
  g_id VARCHAR(21) PRIMARY KEY,
  group_name VARCHAR(100) NOT NULL,
  description TEXT,
  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_groups_resource_owners FOREIGN KEY (g_id) REFERENCES resource_owners(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_group_list_group_name ON group_list(group_name);

-- users
CREATE TABLE IF NOT EXISTS users (
  u_id VARCHAR(21) PRIMARY KEY,
  username VARCHAR(50) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  email VARCHAR(100),
  full_name VARCHAR(50),
  type user_type NOT NULL DEFAULT 'origin',
  status user_status NOT NULL DEFAULT 'offline',
  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  CONSTRAINT fk_users_resource_owners FOREIGN KEY (u_id) REFERENCES resource_owners(id) ON DELETE CASCADE
);

-- user_group
CREATE TABLE IF NOT EXISTS user_group (
  u_id VARCHAR(21) NOT NULL,
  g_id VARCHAR(21) NOT NULL,
  role user_role NOT NULL DEFAULT 'user',
  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (u_id, g_id),
  FOREIGN KEY (u_id) REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (g_id) REFERENCES group_list(g_id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_group_g_id ON user_group(g_id);

-- projects (hierarchical)
CREATE TABLE IF NOT EXISTS projects (
  p_id VARCHAR(21) PRIMARY KEY,
  project_name VARCHAR(100) NOT NULL,
  description TEXT,
  g_id VARCHAR(21) REFERENCES group_list(g_id) ON DELETE CASCADE ON UPDATE CASCADE,
  owner_id VARCHAR(21) REFERENCES resource_owners(id) ON DELETE SET NULL,
  parent_id VARCHAR(21) REFERENCES projects(p_id) ON DELETE CASCADE,
  path ltree NOT NULL,
  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_projects_g_id ON projects(g_id);
CREATE INDEX IF NOT EXISTS idx_projects_project_name ON projects(project_name);
CREATE INDEX IF NOT EXISTS project_path_gist_idx ON projects USING GIST (path);
CREATE INDEX IF NOT EXISTS project_parent_idx ON projects(parent_id);

-- git-like config versioning
CREATE TABLE IF NOT EXISTS config_blobs (
    hash CHAR(64) PRIMARY KEY,
    content JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS config_commits (
    id VARCHAR(21) PRIMARY KEY,
    project_id VARCHAR(21) NOT NULL REFERENCES projects(p_id) ON DELETE CASCADE,
    blob_hash CHAR(64) NOT NULL REFERENCES config_blobs(hash),
    author_id VARCHAR(21) NOT NULL REFERENCES resource_owners(id),
    message TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    tag VARCHAR(50)
);

CREATE INDEX IF NOT EXISTS config_commits_project_idx ON config_commits (project_id, created_at DESC);

-- resources
CREATE TABLE IF NOT EXISTS resources (
  r_id VARCHAR(21) PRIMARY KEY,
  config_commit_id VARCHAR(21) NOT NULL REFERENCES config_commits(id) ON DELETE CASCADE ON UPDATE CASCADE,
  type resource_type NOT NULL,
  name VARCHAR(50) NOT NULL,
  parsed_yaml JSONB NOT NULL,
  description TEXT,
  create_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_resources_config_commit_id ON resources(config_commit_id);

-- unified storages
CREATE TABLE IF NOT EXISTS storages (
    id VARCHAR(21) PRIMARY KEY,
    owner_id VARCHAR(21) NOT NULL REFERENCES resource_owners(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    affinity_config JSONB NOT NULL DEFAULT '{}'::jsonb,
    pvc_name VARCHAR(255),
    host_path VARCHAR(255),
    capacity INTEGER NOT NULL,
    storage_class VARCHAR(100),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- resource plans (time windows)
CREATE TABLE IF NOT EXISTS resource_plans (
    id VARCHAR(21) PRIMARY KEY,
    project_id VARCHAR(21) NOT NULL REFERENCES projects(p_id) ON DELETE CASCADE,
    resource_type VARCHAR(50) NOT NULL,
    amount INT NOT NULL CHECK (amount > 0),
    week_window int4range NOT NULL,
    EXCLUDE USING GIST (
        project_id WITH =,
        resource_type WITH =,
        week_window WITH &&
    )
);

-- priority classes
CREATE TABLE IF NOT EXISTS priority_classes (
    id VARCHAR(21) PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    value INT NOT NULL,
    preemption_policy VARCHAR(50) DEFAULT 'PreemptLowerPriority',
    global_default BOOLEAN DEFAULT FALSE
);

-- jobs
CREATE TABLE IF NOT EXISTS jobs (
  id VARCHAR(21) PRIMARY KEY,
  user_id VARCHAR(21) NOT NULL REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
  project_id VARCHAR(21) REFERENCES projects(p_id) ON DELETE CASCADE ON UPDATE CASCADE,
  config_commit_id VARCHAR(21) REFERENCES config_commits(id),
  name VARCHAR(100) NOT NULL,
  namespace VARCHAR(100) NOT NULL,
  image VARCHAR(255) NOT NULL,
  status VARCHAR(50) DEFAULT 'Pending',
  submit_type VARCHAR(20),
  queue_name VARCHAR(50),
  priority INTEGER DEFAULT 0,
  priority_class_id VARCHAR(21) REFERENCES priority_classes(id),
  priority_value INT NOT NULL DEFAULT 0,
  required_gpu INT NOT NULL DEFAULT 0,
  error_message TEXT,
  k8s_job_name VARCHAR(100) NOT NULL,
  submitted_at TIMESTAMP,
  started_at TIMESTAMP,
  completed_at TIMESTAMP,
  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_jobs_user_id ON jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_jobs_project_id ON jobs(project_id);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);
CREATE INDEX IF NOT EXISTS idx_jobs_user_project_status ON jobs(user_id, project_id, status);
CREATE INDEX IF NOT EXISTS jobs_queue_idx ON jobs (priority_value DESC, created_at ASC) WHERE status = 'Pending';
CREATE INDEX IF NOT EXISTS jobs_running_preemption_idx ON jobs (priority_value ASC, created_at DESC) WHERE status = 'Running';

-- audit_logs
CREATE TABLE IF NOT EXISTS audit_logs (
  id SERIAL PRIMARY KEY,
  user_id VARCHAR(21) NOT NULL,
  action VARCHAR(20) NOT NULL,
  resource_type VARCHAR(50) NOT NULL,
  resource_id VARCHAR NOT NULL,
  old_data JSONB,
  new_data JSONB,
  ip_address VARCHAR(45),
  user_agent TEXT,
  description TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_type ON audit_logs(resource_type);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);

-- forms
CREATE TABLE IF NOT EXISTS forms (
    id VARCHAR(21) PRIMARY KEY,
    user_id VARCHAR(21) REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
    project_id VARCHAR(21) REFERENCES projects(p_id) ON DELETE CASCADE ON UPDATE CASCADE,
    title TEXT,
    description TEXT,
    tag TEXT,
    status TEXT DEFAULT 'Pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_forms_user_id ON forms(user_id);
CREATE INDEX IF NOT EXISTS idx_forms_project_id ON forms(project_id);
CREATE INDEX IF NOT EXISTS idx_forms_deleted_at ON forms(deleted_at);

-- form_messages
CREATE TABLE IF NOT EXISTS form_messages (
    id VARCHAR(21) PRIMARY KEY,
    form_id VARCHAR(21) REFERENCES forms(id) ON DELETE CASCADE ON UPDATE CASCADE,
    user_id VARCHAR(21) REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
    content TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_form_messages_form_id ON form_messages(form_id);
CREATE INDEX IF NOT EXISTS idx_form_messages_user_id ON form_messages(user_id);

-- container_repositories
CREATE TABLE IF NOT EXISTS container_repositories (
    id VARCHAR(21) PRIMARY KEY,
    registry VARCHAR(255) DEFAULT 'docker.io',
    namespace VARCHAR(255),
    name VARCHAR(255),
    full_name VARCHAR(512),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_container_repositories_full_name ON container_repositories(full_name);
CREATE INDEX IF NOT EXISTS idx_container_repositories_name ON container_repositories(name);
CREATE INDEX IF NOT EXISTS idx_container_repositories_deleted_at ON container_repositories(deleted_at);

-- container_tags
CREATE TABLE IF NOT EXISTS container_tags (
    id VARCHAR(21) PRIMARY KEY,
    repository_id VARCHAR(21) NOT NULL REFERENCES container_repositories(id) ON DELETE CASCADE ON UPDATE CASCADE,
    name VARCHAR(128),
    digest VARCHAR(255),
    size BIGINT,
    pushed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_container_tags_repository_id ON container_tags(repository_id);
CREATE INDEX IF NOT EXISTS idx_container_tags_name ON container_tags(name);
CREATE INDEX IF NOT EXISTS idx_container_tags_deleted_at ON container_tags(deleted_at);

-- image_requests
CREATE TABLE IF NOT EXISTS image_requests (
    id VARCHAR(21) PRIMARY KEY,
    user_id VARCHAR(21) REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
    project_id VARCHAR(21) REFERENCES projects(p_id) ON DELETE CASCADE ON UPDATE CASCADE,
    input_registry TEXT,
    input_image_name TEXT,
    input_tag TEXT,
    status VARCHAR(32) DEFAULT 'pending',
    reviewer_id VARCHAR(21) REFERENCES users(u_id) ON DELETE SET NULL ON UPDATE CASCADE,
    reviewed_at TIMESTAMP WITH TIME ZONE,
    reviewer_note TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_image_requests_user_id ON image_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_image_requests_project_id ON image_requests(project_id);
CREATE INDEX IF NOT EXISTS idx_image_requests_status ON image_requests(status);
CREATE INDEX IF NOT EXISTS idx_image_requests_deleted_at ON image_requests(deleted_at);

-- image_allow_lists
CREATE TABLE IF NOT EXISTS image_allow_lists (
    id VARCHAR(21) PRIMARY KEY,
    project_id VARCHAR(21) REFERENCES projects(p_id) ON DELETE CASCADE ON UPDATE CASCADE,
    tag_id VARCHAR(21) REFERENCES container_tags(id) ON DELETE CASCADE ON UPDATE CASCADE,
    repository_id VARCHAR(21) NOT NULL REFERENCES container_repositories(id) ON DELETE CASCADE ON UPDATE CASCADE,
    request_id VARCHAR(21) REFERENCES image_requests(id) ON DELETE SET NULL ON UPDATE CASCADE,
    created_by VARCHAR(21) REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
    is_enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_image_allow_lists_project_id ON image_allow_lists(project_id);
CREATE INDEX IF NOT EXISTS idx_image_allow_lists_tag_id ON image_allow_lists(tag_id);
CREATE INDEX IF NOT EXISTS idx_image_allow_lists_repository_id ON image_allow_lists(repository_id);
CREATE INDEX IF NOT EXISTS idx_image_allow_lists_deleted_at ON image_allow_lists(deleted_at);
CREATE INDEX IF NOT EXISTS idx_image_allow_lists_check ON image_allow_lists(project_id, repository_id, is_enabled);

-- cluster_image_statuses
CREATE TABLE IF NOT EXISTS cluster_image_statuses (
    id VARCHAR(21) PRIMARY KEY,
    tag_id VARCHAR(21) REFERENCES container_tags(id) ON DELETE CASCADE ON UPDATE CASCADE,
    is_pulled BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_cluster_image_statuses_tag_id ON cluster_image_statuses(tag_id);

-- group_storage_permissions
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

-- group_storage_access_policies
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

-- Function and triggers to auto-update updated_at for group storage policies
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = CURRENT_TIMESTAMP;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_gsp_updated_at
  BEFORE UPDATE ON group_storage_permissions
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_update_gsap_updated_at
  BEFORE UPDATE ON group_storage_access_policies
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();
