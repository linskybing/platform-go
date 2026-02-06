package project

import (
	"time"

	"github.com/linskybing/platform-go/internal/domain/group"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
)

// Project represents a user project with resource quotas
type Project struct {
	PID         string       `gorm:"primaryKey;column:p_id;size:20"`
	ProjectName string       `gorm:"size:100;not null"`
	Description string       `gorm:"type:text"`
	GID         string       `gorm:"not null;index;foreignKey:GID;references:GID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"` // Group ID
	GPUQuota    int          `gorm:"default:0;column:gpu_quota"`                                                                // GPU quota in integer units
	CreatedAt   time.Time    `gorm:"column:create_at;autoCreateTime"`
	UpdatedAt   time.Time    `gorm:"column:update_at;autoUpdateTime"`
	Group       *group.Group `json:"-" gorm:"foreignKey:GID;references:GID"`
}

// TableName specifies the database table name
func (Project) TableName() string {
	return "project_list"
}

// BeforeCreate hooks into GORM to generate ID
func (p *Project) BeforeCreate(tx *gorm.DB) (err error) {
	if p.PID == "" {
		p.PID, err = gonanoid.New(12)
	}
	return
}

// HasGPUQuota checks if project has GPU quota available
func (p *Project) HasGPUQuota() bool {
	return p.GPUQuota > 0
}
