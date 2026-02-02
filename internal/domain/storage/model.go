package storage

import "time"

// GroupPVC represents a group's persistent volume claim
type GroupPVC struct {
	ID           string    `gorm:"primaryKey;column:id;size:100"`          // Format: group-{gid}-{uuid}
	Name         string    `gorm:"size:100;not null;index"`                // Display name
	GroupID      uint      `gorm:"not null;index"`                         // Foreign key to group
	Namespace    string    `gorm:"size:100;not null"`                      // K8s namespace
	PVCName      string    `gorm:"size:100;not null;uniqueIndex:uidx_pvc"` // K8s PVC name
	Size         string    `gorm:"size:50;not null"`                       // Size in Gi format (e.g., "100Gi")
	Capacity     int       `gorm:"not null"`                               // Capacity in Gi (numeric)
	StorageClass string    `gorm:"size:100;default:'longhorn'"`            // Storage class name
	AccessMode   string    `gorm:"size:50;default:'ReadWriteMany'"`        // RWX, RWO, etc
	Status       string    `gorm:"size:50;default:'Pending'"`              // K8s PVC status
	CreatedBy    uint      `gorm:"not null"`                               // User ID who created
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;index"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName specifies the database table name
func (GroupPVC) TableName() string {
	return "group_pvcs"
}

// StorageHub represents a storage access point (pod that mounts volumes)
type StorageHub struct {
	ID        uint      `gorm:"primaryKey;column:id"`
	Name      string    `gorm:"size:100;not null"`
	Namespace string    `gorm:"size:100;not null"`
	PVCName   string    `gorm:"size:100;not null"`
	MountPath string    `gorm:"size:255;default:'/data'"`
	Type      string    `gorm:"size:50;default:'group'"` // 'group' or 'user'
	Owner     string    `gorm:"size:100"`                // group ID or username
	Status    string    `gorm:"size:50;default:'Pending'"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName specifies the database table name
func (StorageHub) TableName() string {
	return "storage_hubs"
}

// PersistentVolumeClaim represents a Kubernetes PVC resource (legacy)
type PersistentVolumeClaim struct {
	ID               uint      `gorm:"primaryKey;column:id"`
	Name             string    `gorm:"size:100;not null"`
	Namespace        string    `gorm:"size:100;not null"`
	Size             string    `gorm:"size:50"`
	Status           string    `gorm:"size:50;default:'Pending'"`
	StorageClassName string    `gorm:"size:100"`
	AccessMode       string    `gorm:"size:50;default:'ReadWriteMany'"`
	ProjectID        *uint     `gorm:"column:project_id"`
	ProjectName      *string   `gorm:"size:255"`
	IsGlobal         bool      `gorm:"default:false"`
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName specifies the database table name
func (PersistentVolumeClaim) TableName() string {
	return "persistent_volume_claims"
}
