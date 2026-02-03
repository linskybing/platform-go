package storage

import (
	"time"
)

// StoragePermission defines access levels for group storage
// Similar to Linux permissions (755), but simplified for storage
type StoragePermission string

const (
	// PermissionNone - User cannot access the storage
	PermissionNone StoragePermission = "none"

	// PermissionReadOnly - User can only read files
	PermissionReadOnly StoragePermission = "read"

	// PermissionReadWrite - User can read and write files
	PermissionReadWrite StoragePermission = "write"
)

// GroupStoragePermission represents user permissions for a specific group PVC
type GroupStoragePermission struct {
	ID         uint              `gorm:"primaryKey;column:id;autoIncrement"`
	GroupID    uint              `gorm:"not null;index:idx_group_pvc_user"`          // Group ID
	PVCID      string            `gorm:"size:100;not null;index:idx_group_pvc_user"` // PVC ID (group-{gid}-{uuid})
	PVCName    string            `gorm:"size:100;not null;index"`                    // K8s PVC name for quick lookup
	UserID     uint              `gorm:"not null;index:idx_group_pvc_user,unique"`   // User ID
	Permission StoragePermission `gorm:"type:varchar(20);not null;default:'none'"`   // none, read, write
	GrantedBy  uint              `gorm:"not null"`                                   // Admin who granted permission
	GrantedAt  time.Time         `gorm:"column:granted_at;autoCreateTime"`
	UpdatedAt  time.Time         `gorm:"column:updated_at;autoUpdateTime"`
	RevokedAt  *time.Time        `gorm:"column:revoked_at;index"` // NULL if active
}

// TableName specifies the database table name
func (GroupStoragePermission) TableName() string {
	return "group_storage_permissions"
}

// IsActive returns true if the permission is currently active (not revoked)
func (p *GroupStoragePermission) IsActive() bool {
	return p.RevokedAt == nil
}

// CanRead returns true if user has at least read permission
func (p *GroupStoragePermission) CanRead() bool {
	return p.IsActive() && (p.Permission == PermissionReadOnly || p.Permission == PermissionReadWrite)
}

// CanWrite returns true if user has write permission
func (p *GroupStoragePermission) CanWrite() bool {
	return p.IsActive() && p.Permission == PermissionReadWrite
}

// GroupStorageAccessPolicy defines default access policy for a group PVC
type GroupStorageAccessPolicy struct {
	ID                uint              `gorm:"primaryKey;column:id;autoIncrement"`
	GroupID           uint              `gorm:"not null;index"`
	PVCID             string            `gorm:"size:100;not null;uniqueIndex"`            // One policy per PVC
	DefaultPermission StoragePermission `gorm:"type:varchar(20);not null;default:'none'"` // Default for new members
	AdminOnly         bool              `gorm:"default:false"`                            // Only admins can access
	CreatedBy         uint              `gorm:"not null"`                                 // Admin who created policy
	CreatedAt         time.Time         `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time         `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName specifies the database table name
func (GroupStorageAccessPolicy) TableName() string {
	return "group_storage_access_policies"
}

// ProjectPVCBinding represents a PVC in user's project namespace that binds to group PV
// This allows users to mount group storage in their own project namespaces
type ProjectPVCBinding struct {
	ID               uint      `gorm:"primaryKey;column:id;autoIncrement"`
	ProjectID        uint      `gorm:"not null;index"`                 // Project ID
	UserID           uint      `gorm:"not null;index"`                 // User ID
	GroupPVCID       string    `gorm:"size:100;not null;index"`        // Source group PVC ID
	ProjectPVCName   string    `gorm:"size:100;not null;uniqueIndex"`  // PVC name in project namespace
	ProjectNamespace string    `gorm:"size:100;not null;index"`        // Project namespace
	SourcePVName     string    `gorm:"size:200;not null"`              // Original PV name to bind
	AccessMode       string    `gorm:"size:50;default:'ReadOnlyMany'"` // ReadOnlyMany or ReadWriteMany
	Status           string    `gorm:"size:50;default:'Pending'"`      // Bound, Pending, Failed
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName specifies the database table name
func (ProjectPVCBinding) TableName() string {
	return "project_pvc_bindings"
}
