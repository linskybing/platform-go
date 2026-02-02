package project

import "time"

// Project represents a user project with resource quotas
type Project struct {
	PID         uint      `gorm:"primaryKey;column:p_id;autoIncrement"`
	ProjectName string    `gorm:"size:100;not null"`
	Description string    `gorm:"type:text"`
	GID         uint      `gorm:"not null"`                   // Group ID
	GPUQuota    int       `gorm:"default:0;column:gpu_quota"` // GPU quota in integer units
	CreatedAt   time.Time `gorm:"column:create_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:update_at;autoUpdateTime"`
}

// TableName specifies the database table name
func (Project) TableName() string {
	return "project_list"
}

// HasGPUQuota checks if project has GPU quota available
func (p *Project) HasGPUQuota() bool {
	return p.GPUQuota > 0
}
