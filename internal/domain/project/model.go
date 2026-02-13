package project

import (
	"encoding/json"
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Project represents a user project with resource quotas
type Project struct {
	PID         string `gorm:"primaryKey;column:p_id;size:20"`
	ProjectName string `gorm:"size:100;not null"`
	Description string `gorm:"type:text"`

	GID string `gorm:"column:g_id;size:20;not null;index"`

	GPUQuota int `gorm:"default:0;column:gpu_quota"`

	MaxConcurrentJobsPerUser int            `gorm:"default:0;column:max_concurrent_jobs_per_user"`
	MaxQueuedJobsPerUser     int            `gorm:"default:0;column:max_queued_jobs_per_user"`
	MaxJobRuntimeSeconds     int            `gorm:"default:0;column:max_job_runtime_seconds"`
	MaxProjectUsers          int            `gorm:"default:0;column:max_project_users"`
	ScheduleWindows          datatypes.JSON `gorm:"type:jsonb;column:schedule_windows"`

	CreatedAt time.Time `gorm:"column:create_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:update_at;autoUpdateTime"`

	// Note: intentionally not defining a `Group` association field here to avoid
	// GORM inferring/creating incorrect foreign keys during AutoMigrate.
	// Use repository methods to load the group when needed.
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

func (p *Project) ScheduleWindowList() ([]ScheduleWindow, error) {
	if len(p.ScheduleWindows) == 0 {
		return nil, nil
	}
	var windows []ScheduleWindow
	if err := json.Unmarshal(p.ScheduleWindows, &windows); err != nil {
		return nil, err
	}
	return windows, nil
}
