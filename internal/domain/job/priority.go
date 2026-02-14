package job

import (
	"time"
)

// PriorityClass represents the database mapping of Kubernetes PriorityClass.
type PriorityClass struct {
	ID               string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name             string    `gorm:"size:100;not null;uniqueIndex"`
	Value            int       `gorm:"not null;index"`
	PreemptionPolicy string    `gorm:"size:50;default:'PreemptLowerPriority'"`
	Description      string    `gorm:"type:text"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
}

func (PriorityClass) TableName() string { return "priority_classes" }
