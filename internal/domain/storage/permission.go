package storage

import (
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
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
	ID         string            `gorm:"primaryKey;column:id;size:21"`
	GroupID    string            `gorm:"not null;index:idx_group_pvc_user;size:21"`        // Group ID
	PVCID      string            `gorm:"size:100;not null;index:idx_group_pvc_user"`       // PVC ID (group-{gid}-{uuid})
	PVCName    string            `gorm:"size:100;not null;index"`                          // K8s PVC name for quick lookup
	UserID     string            `gorm:"not null;index:idx_group_pvc_user,unique;size:21"` // User ID
	Permission StoragePermission `gorm:"type:varchar(20);not null;default:'none'"`         // none, read, write
	GrantedBy  string            `gorm:"not null;size:21"`                                 // Admin who granted permission
	GrantedAt  time.Time         `gorm:"column:granted_at;autoCreateTime"`
	UpdatedAt  time.Time         `gorm:"column:updated_at;autoUpdateTime"`
	RevokedAt  *time.Time        `gorm:"column:revoked_at;index"` // NULL if active
}

func (m *GroupStoragePermission) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID, err = gonanoid.New()
	}
	return
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
	ID                string            `gorm:"primaryKey;column:id;size:21"`
	GroupID           string            `gorm:"not null;index;size:21"`
	PVCID             string            `gorm:"size:100;not null;uniqueIndex"`            // One policy per PVC
	DefaultPermission StoragePermission `gorm:"type:varchar(20);not null;default:'none'"` // Default for new members
	AdminOnly         bool              `gorm:"default:false"`                            // Only admins can access
	CreatedBy         string            `gorm:"not null;size:21"`                         // Admin who created policy
	CreatedAt         time.Time         `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time         `gorm:"column:updated_at;autoUpdateTime"`
}

func (m *GroupStorageAccessPolicy) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID, err = gonanoid.New()
	}
	return
}

// TableName specifies the database table name
func (GroupStorageAccessPolicy) TableName() string {
	return "group_storage_access_policies"
}
