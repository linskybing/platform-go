package storage

import (
	"time"

	"gorm.io/datatypes"
)

// Storage represents a PVC or HostPath mapping with node affinity.
type Storage struct {
	ID           string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	OwnerID      string         `gorm:"type:uuid;not null;index"` // Polymorphic: User or Group ID
	GroupID      string         `gorm:"-"`                        // Legacy Alias
	UserID       string         `gorm:"-"`                        // Legacy Alias
	Name         string         `gorm:"size:100;not null"`
	K8sNamespace string         `gorm:"size:100;not null"`
	PVCName      string         `gorm:"size:100;not null"`
	HostPath     string         `gorm:"size:255"`
	Capacity     int            `gorm:"not null"`
	StorageClass string         `gorm:"size:100"`
	NodeAffinity datatypes.JSON `gorm:"type:jsonb;not null;default:'{}'"` // PSS Restricted strategy
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
}

func (Storage) TableName() string { return "storages" }
