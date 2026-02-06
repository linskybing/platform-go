package storage

import "time"

// ===== Permission Management DTOs =====

// SetStoragePermissionRequest defines the payload for setting user permission on a group PVC
type SetStoragePermissionRequest struct {
	GroupID    string            `json:"group_id" binding:"required"`
	PVCID      string            `json:"pvc_id" binding:"required"` // group-{gid}-{uuid}
	UserID     string            `json:"user_id" binding:"required"`
	Permission StoragePermission `json:"permission" binding:"required,oneof=none read write"`
}

// BatchSetPermissionsRequest allows setting permissions for multiple users at once
type BatchSetPermissionsRequest struct {
	GroupID     string           `json:"group_id" binding:"required"`
	PVCID       string           `json:"pvc_id" binding:"required"`
	Permissions []UserPermission `json:"permissions" binding:"required,dive"`
}

// UserPermission represents a single user's permission
type UserPermission struct {
	UserID     string            `json:"user_id" binding:"required"`
	Permission StoragePermission `json:"permission" binding:"required,oneof=none read write"`
}

// SetStorageAccessPolicyRequest defines default access policy for a group PVC
type SetStorageAccessPolicyRequest struct {
	GroupID           string            `json:"group_id" binding:"required"`
	PVCID             string            `json:"pvc_id" binding:"required"`
	DefaultPermission StoragePermission `json:"default_permission" binding:"required,oneof=none read write"`
	AdminOnly         bool              `json:"admin_only"`
}

// StoragePermissionInfo represents permission information for display
type StoragePermissionInfo struct {
	UserID     string            `json:"user_id"`
	Username   string            `json:"username"`
	Permission StoragePermission `json:"permission"`
	CanRead    bool              `json:"can_read"`
	CanWrite   bool              `json:"can_write"`
	GrantedBy  string            `json:"granted_by"`
	GrantedAt  time.Time         `json:"granted_at"`
}

// GroupPVCWithPermissions extends GroupPVCSpec with user permission info
type GroupPVCWithPermissions struct {
	GroupPVCSpec
	UserPermission StoragePermission `json:"user_permission"` // Current user's permission
	CanAccess      bool              `json:"can_access"`
	CanModify      bool              `json:"can_modify"`
}

// CreateProjectPVCBindingRequest defines the payload for creating a PVC binding in project namespace
type CreateProjectPVCBindingRequest struct {
	ProjectID  string `json:"project_id" binding:"required"`
	GroupPVCID string `json:"group_pvc_id" binding:"required"` // Source group PVC
	PVCName    string `json:"pvc_name" binding:"required"`     // Name in project namespace
	ReadOnly   bool   `json:"read_only"`                       // Mount as read-only
}

// ProjectPVCBindingInfo represents project PVC binding information
type ProjectPVCBindingInfo struct {
	ID               string    `json:"id"`
	ProjectID        string    `json:"project_id"`
	ProjectName      string    `json:"project_name"`
	GroupPVCID       string    `json:"group_pvc_id"`
	ProjectPVCName   string    `json:"project_pvc_name"`
	ProjectNamespace string    `json:"project_namespace"`
	AccessMode       string    `json:"access_mode"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
}

// FileBrowserAccessRequest defines request for accessing FileBrowser
type FileBrowserAccessRequest struct {
	GroupID string `json:"group_id" binding:"required"`
	PVCID   string `json:"pvc_id" binding:"required"`
	UserID  string `json:"user_id" binding:"required"`
}

// FileBrowserAccessResponse contains FileBrowser connection info
type FileBrowserAccessResponse struct {
	Allowed  bool   `json:"allowed"`
	URL      string `json:"url,omitempty"`
	Port     string `json:"port,omitempty"`
	PodName  string `json:"pod_name,omitempty"`
	ReadOnly bool   `json:"read_only"`
	Message  string `json:"message,omitempty"` // Error or info message
}
