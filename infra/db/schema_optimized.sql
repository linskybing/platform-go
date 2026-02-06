-- Optimized Database Schema with Foreign Keys and Cascade Deletes
-- Platform-Go Production Database Schema
-- Created: 2026-02-05

CREATE DATABASE IF NOT EXISTS platform;
\c platform;

-- ============================================================================
-- ENUMS
-- ============================================================================
CREATE TYPE resource_type AS ENUM ('pod', 'service', 'deployment', 'configmap', 'ingress', 'job');
CREATE TYPE user_type AS ENUM ('origin', 'oauth2');
CREATE TYPE user_status AS ENUM ('online', 'offline', 'delete');
CREATE TYPE user_role AS ENUM ('admin', 'manager', 'user');

-- ============================================================================
-- CORE TABLES
-- ============================================================================

-- Users: Base user table with soft delete support
CREATE TABLE users (
  u_id VARCHAR(20) PRIMARY KEY,
  username VARCHAR(50) NOT NULL UNIQUE,
  password VARCHAR(255) NOT NULL,
  email VARCHAR(100),
  full_name VARCHAR(50),
  type user_type NOT NULL DEFAULT 'origin',
  status user_status NOT NULL DEFAULT 'offline',
  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  -- Indexes for common queries
  INDEX idx_username (username),
  INDEX idx_email (email),
  INDEX idx_status (status),
  INDEX idx_created_at (create_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Groups: Organization/Team groups
CREATE TABLE group_list (
  g_id VARCHAR(20) PRIMARY KEY,
  group_name VARCHAR(100) NOT NULL,
  description TEXT,
  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  -- Indexes for common queries
  INDEX idx_group_name (group_name),
  INDEX idx_created_at (create_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- UserGroup: Many-to-many relationship between users and groups
CREATE TABLE user_group (
  u_id VARCHAR(20) NOT NULL,
  g_id VARCHAR(20) NOT NULL,
  role user_role NOT NULL DEFAULT 'user',
  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (u_id, g_id),
  FOREIGN KEY (u_id) REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (g_id) REFERENCES group_list(g_id) ON DELETE CASCADE ON UPDATE CASCADE,
  -- Indexes for common queries
  INDEX idx_g_id (g_id),
  INDEX idx_role (role),
  INDEX idx_created_at (create_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Projects: Projects belonging to groups with GPU quota
CREATE TABLE project_list (
  p_id VARCHAR(20) PRIMARY KEY,
  project_name VARCHAR(100) NOT NULL,
  description TEXT,
  g_id VARCHAR(20) NOT NULL,
  gpu_quota INT DEFAULT 0,
  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (g_id) REFERENCES group_list(g_id) ON DELETE CASCADE ON UPDATE CASCADE,
  -- Indexes for common queries
  INDEX idx_project_name (project_name),
  INDEX idx_g_id (g_id),
  INDEX idx_gpu_quota (gpu_quota),
  INDEX idx_created_at (create_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- CONFIGURATION TABLES
-- ============================================================================

-- ConfigFiles: YAML configuration files for projects
CREATE TABLE config_files (
  cf_id VARCHAR(21) PRIMARY KEY,
  filename VARCHAR(200) NOT NULL,
  content VARCHAR(10000),
  project_id VARCHAR(20) NOT NULL,
  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  update_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  FOREIGN KEY (project_id) REFERENCES project_list(p_id) ON DELETE CASCADE ON UPDATE CASCADE,
  -- Indexes for common queries
  INDEX idx_project_id (project_id),
  INDEX idx_filename (filename),
  INDEX idx_created_at (create_at),
  UNIQUE KEY uk_project_filename (project_id, filename)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Resources: Kubernetes resources parsed from config files
CREATE TABLE resources (
  r_id VARCHAR(21) PRIMARY KEY,
  cf_id VARCHAR(21) NOT NULL,
  type resource_type NOT NULL,
  name VARCHAR(50) NOT NULL,
  parsed_yaml JSON NOT NULL,
  description TEXT,
  create_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (cf_id) REFERENCES config_files(cf_id) ON DELETE CASCADE ON UPDATE CASCADE,
  -- Indexes for common queries
  INDEX idx_cf_id (cf_id),
  INDEX idx_type (type),
  INDEX idx_name (name),
  INDEX idx_created_at (create_at),
  UNIQUE KEY uk_cf_name (cf_id, name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- FORMS & MESSAGES TABLES
-- ============================================================================

-- Forms: User request forms with project association
CREATE TABLE forms (
  id VARCHAR(21) PRIMARY KEY,
  user_id VARCHAR(20) NOT NULL,
  project_id VARCHAR(20),
  title VARCHAR(255) NOT NULL,
  description TEXT,
  tag VARCHAR(100),
  status VARCHAR(20) NOT NULL DEFAULT 'Pending',
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  FOREIGN KEY (user_id) REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (project_id) REFERENCES project_list(p_id) ON DELETE CASCADE ON UPDATE CASCADE,
  -- Indexes for common queries
  INDEX idx_user_id (user_id),
  INDEX idx_project_id (project_id),
  INDEX idx_status (status),
  INDEX idx_tag (tag),
  INDEX idx_created_at (created_at),
  INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- FormMessages: Comments/messages on forms
CREATE TABLE form_messages (
  id VARCHAR(21) PRIMARY KEY,
  form_id VARCHAR(21) NOT NULL,
  user_id VARCHAR(20) NOT NULL,
  content TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (form_id) REFERENCES forms(id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (user_id) REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
  -- Indexes for common queries
  INDEX idx_form_id (form_id),
  INDEX idx_user_id (user_id),
  INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- IMAGE REGISTRY TABLES
-- ============================================================================

-- ContainerRepositories: Docker/Container registries
CREATE TABLE container_repositories (
  id VARCHAR(21) PRIMARY KEY,
  registry VARCHAR(255) NOT NULL DEFAULT 'docker.io',
  namespace VARCHAR(255),
  name VARCHAR(255) NOT NULL,
  full_name VARCHAR(512) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  -- Indexes for common queries
  INDEX idx_name (name),
  INDEX idx_full_name (full_name),
  INDEX idx_registry (registry),
  INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ContainerTags: Tags for container images
CREATE TABLE container_tags (
  id VARCHAR(21) PRIMARY KEY,
  repository_id VARCHAR(21) NOT NULL,
  name VARCHAR(128) NOT NULL,
  digest VARCHAR(255),
  size BIGINT,
  pushed_at TIMESTAMP NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  FOREIGN KEY (repository_id) REFERENCES container_repositories(id) ON DELETE CASCADE ON UPDATE CASCADE,
  -- Indexes for common queries
  INDEX idx_repository_id (repository_id),
  INDEX idx_name (name),
  INDEX idx_digest (digest),
  INDEX idx_pushed_at (pushed_at),
  INDEX idx_deleted_at (deleted_at),
  UNIQUE KEY uk_repo_tag (repository_id, name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ImageRequests: Requests to add images to allowlist
CREATE TABLE image_requests (
  id VARCHAR(21) PRIMARY KEY,
  user_id VARCHAR(20) NOT NULL,
  project_id VARCHAR(20),
  input_registry VARCHAR(255),
  input_image_name VARCHAR(255) NOT NULL,
  input_tag VARCHAR(128) NOT NULL,
  status VARCHAR(32) NOT NULL DEFAULT 'pending',
  reviewer_id VARCHAR(20),
  reviewed_at TIMESTAMP NULL,
  reviewer_note TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  FOREIGN KEY (user_id) REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (project_id) REFERENCES project_list(p_id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (reviewer_id) REFERENCES users(u_id) ON DELETE SET NULL ON UPDATE CASCADE,
  -- Indexes for common queries
  INDEX idx_user_id (user_id),
  INDEX idx_project_id (project_id),
  INDEX idx_status (status),
  INDEX idx_reviewer_id (reviewer_id),
  INDEX idx_created_at (created_at),
  INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ImageAllowList: Approved images for projects
CREATE TABLE image_allow_list (
  id VARCHAR(21) PRIMARY KEY,
  project_id VARCHAR(20),
  tag_id VARCHAR(21),
  repository_id VARCHAR(21) NOT NULL,
  request_id VARCHAR(21),
  created_by VARCHAR(20) NOT NULL,
  is_enabled BOOLEAN NOT NULL DEFAULT true,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  FOREIGN KEY (project_id) REFERENCES project_list(p_id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (tag_id) REFERENCES container_tags(id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (repository_id) REFERENCES container_repositories(id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (request_id) REFERENCES image_requests(id) ON DELETE SET NULL ON UPDATE CASCADE,
  FOREIGN KEY (created_by) REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
  -- Indexes for common queries
  INDEX idx_project_id (project_id),
  INDEX idx_tag_id (tag_id),
  INDEX idx_repository_id (repository_id),
  INDEX idx_request_id (request_id),
  INDEX idx_created_by (created_by),
  INDEX idx_is_enabled (is_enabled),
  INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ClusterImageStatus: Track which images are available in cluster
CREATE TABLE cluster_image_status (
  id VARCHAR(21) PRIMARY KEY,
  tag_id VARCHAR(21) NOT NULL UNIQUE,
  is_pulled BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP NULL,
  FOREIGN KEY (tag_id) REFERENCES container_tags(id) ON DELETE CASCADE ON UPDATE CASCADE,
  -- Indexes for common queries
  INDEX idx_is_pulled (is_pulled),
  INDEX idx_deleted_at (deleted_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- AUDIT & LOGGING TABLES
-- ============================================================================

-- AuditLogs: Complete audit trail for all operations
CREATE TABLE audit_logs (
  id INT AUTO_INCREMENT PRIMARY KEY,
  user_id VARCHAR(20) NOT NULL,
  action VARCHAR(20) NOT NULL,
  resource_type VARCHAR(50) NOT NULL,
  resource_id VARCHAR(255) NOT NULL,
  old_data JSON,
  new_data JSON,
  ip_address VARCHAR(45),
  user_agent TEXT,
  description TEXT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
  -- Indexes for common queries
  INDEX idx_user_id (user_id),
  INDEX idx_action (action),
  INDEX idx_resource_type (resource_type),
  INDEX idx_resource_id (resource_id),
  INDEX idx_created_at (created_at),
  INDEX idx_user_action_resource (user_id, action, resource_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- COURSE MANAGEMENT TABLES
-- ============================================================================

-- CourseWorkloads: High-priority course pods
CREATE TABLE course_workloads (
  id VARCHAR(21) PRIMARY KEY,
  user_id VARCHAR(20) NOT NULL,
  project_id VARCHAR(20) NOT NULL,
  name VARCHAR(100) NOT NULL,
  namespace VARCHAR(100) NOT NULL,
  image VARCHAR(255) NOT NULL,
  status VARCHAR(50) NOT NULL DEFAULT 'Pending',
  k8s_pod_name VARCHAR(100) NOT NULL,
  priority INT NOT NULL DEFAULT 1000,
  resource_cpu VARCHAR(50),
  resource_mem VARCHAR(50),
  resource_gpu INT,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  started_at TIMESTAMP NULL,
  FOREIGN KEY (user_id) REFERENCES users(u_id) ON DELETE CASCADE ON UPDATE CASCADE,
  FOREIGN KEY (project_id) REFERENCES project_list(p_id) ON DELETE CASCADE ON UPDATE CASCADE,
  -- Indexes for common queries
  INDEX idx_user_id (user_id),
  INDEX idx_project_id (project_id),
  INDEX idx_status (status),
  INDEX idx_priority (priority),
  INDEX idx_created_at (created_at),
  UNIQUE KEY uk_k8s_pod_name (k8s_pod_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- VIEWS
-- ============================================================================

-- View: ProjectGroupViews - Group with project counts
CREATE OR REPLACE VIEW project_group_views AS
SELECT
  g.g_id,
  g.group_name,
  COUNT(DISTINCT p.p_id) AS project_count,
  COUNT(DISTINCT r.r_id) AS resource_count,
  COUNT(DISTINCT ug.u_id) AS user_count,
  g.create_at AS group_create_at,
  g.update_at AS group_update_at
FROM group_list g
LEFT JOIN project_list p ON p.g_id = g.g_id
LEFT JOIN config_files cf ON cf.project_id = p.p_id
LEFT JOIN resources r ON r.cf_id = cf.cf_id
LEFT JOIN user_group ug ON ug.g_id = g.g_id
GROUP BY g.g_id, g.group_name, g.create_at, g.update_at;

-- View: ProjectResourceViews - Projects with their resources
CREATE OR REPLACE VIEW project_resource_views AS
SELECT
  p.p_id,
  p.project_name,
  r.r_id,
  r.type,
  r.name,
  cf.filename,
  r.create_at AS resource_create_at
FROM project_list p
JOIN config_files cf ON cf.project_id = p.p_id
JOIN resources r ON r.cf_id = cf.cf_id
WHERE r.deleted_at IS NULL;

-- View: GroupResourceViews - Groups with resources
CREATE OR REPLACE VIEW group_resource_views AS
SELECT
  g.g_id,
  g.group_name,
  p.p_id,
  p.project_name,
  r.r_id,
  r.type AS resource_type,
  r.name AS resource_name,
  cf.filename,
  r.create_at AS resource_create_at
FROM group_list g
LEFT JOIN project_list p ON p.g_id = g.g_id
LEFT JOIN config_files cf ON cf.project_id = p.p_id
LEFT JOIN resources r ON r.cf_id = cf.cf_id
WHERE r.r_id IS NOT NULL AND r.deleted_at IS NULL;

-- View: UserGroupViews - User group membership with roles
CREATE OR REPLACE VIEW user_group_views AS
SELECT
  u.u_id,
  u.username,
  g.g_id,
  g.group_name,
  ug.role
FROM users u
JOIN user_group ug ON u.u_id = ug.u_id
JOIN group_list g ON ug.g_id = g.g_id;

-- View: UsersWithSuperAdmin - Users with super admin status
CREATE OR REPLACE VIEW users_with_superadmin AS
SELECT
  u.u_id,
  u.username,
  u.password,
  u.email,
  u.full_name,
  u.type,
  u.status,
  u.create_at,
  u.update_at,
  (SELECT COUNT(*) FROM user_group WHERE u_id = u.u_id AND g_id = (SELECT g_id FROM group_list WHERE group_name = 'super') AND role = 'admin') > 0 AS is_super_admin
FROM users u;

-- View: ProjectUserViews - Users per project
CREATE OR REPLACE VIEW project_user_views AS
SELECT
  p.p_id,
  p.project_name,
  g.g_id,
  g.group_name,
  u.u_id,
  u.username
FROM project_list p
JOIN group_list g ON p.g_id = g.g_id
JOIN user_group ug ON ug.g_id = g.g_id
JOIN users u ON u.u_id = ug.u_id;

-- ============================================================================
-- STORED PROCEDURES FOR MAINTENANCE
-- ============================================================================

-- Procedure: Cleanup soft deleted records
DELIMITER $$
CREATE PROCEDURE cleanup_soft_deleted(
  IN retention_days INT
)
BEGIN
  -- Delete soft-deleted forms older than retention days
  DELETE FROM forms 
  WHERE deleted_at IS NOT NULL 
    AND deleted_at < DATE_SUB(NOW(), INTERVAL retention_days DAY);
  
  -- Delete soft-deleted container tags older than retention days
  DELETE FROM container_tags 
  WHERE deleted_at IS NOT NULL 
    AND deleted_at < DATE_SUB(NOW(), INTERVAL retention_days DAY);
  
  -- Delete soft-deleted images older than retention days
  DELETE FROM image_allow_list 
  WHERE deleted_at IS NOT NULL 
    AND deleted_at < DATE_SUB(NOW(), INTERVAL retention_days DAY);
END$$
DELIMITER ;

-- Procedure: Cleanup old audit logs
DELIMITER $$
CREATE PROCEDURE cleanup_audit_logs(
  IN retention_days INT
)
BEGIN
  DELETE FROM audit_logs 
  WHERE created_at < DATE_SUB(NOW(), INTERVAL retention_days DAY);
END$$
DELIMITER ;

-- ============================================================================
-- INITIALIZATION DATA
-- ============================================================================

-- Insert default super group
INSERT INTO group_list (g_id, group_name, description)
VALUES ('super_group', 'Super Administrator Group', 'System administrators with full access')
ON DUPLICATE KEY UPDATE group_name = 'Super Administrator Group';

-- Insert default admin user
INSERT INTO users (u_id, username, password, type, status, full_name, email)
VALUES (
  'admin_001',
  'admin',
  '$2a$10$nsXJXOUAbVyLbvtPizj0RectJWdInu17C2NpWEVKNvwzKQcg8bchu',
  'origin',
  'offline',
  'Administrator',
  'admin@platform.local'
)
ON DUPLICATE KEY UPDATE username = 'admin';

-- Add admin to super group
INSERT INTO user_group (u_id, g_id, role)
VALUES ('admin_001', 'super_group', 'admin')
ON DUPLICATE KEY UPDATE role = 'admin';

-- ============================================================================
-- CONSTRAINTS VERIFICATION
-- ============================================================================

-- Enable foreign key constraints verification
SET FOREIGN_KEY_CHECKS = 1;
