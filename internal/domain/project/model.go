package project

import (
	"time"

	"github.com/linskybing/platform-go/internal/domain/common"
)

// Project represents a node in the project tree (Project or SubProject).
type Project struct {
	ID          string       `gorm:"primaryKey;type:uuid;column:p_id;default:uuid_generate_v4()"` // Map to existing p_id
	PID         string       `gorm:"-"`                                                           // Legacy Alias
	ParentID    *string      `gorm:"type:uuid"`
	Path        common.Ltree `gorm:"type:ltree;not null;index:idx_path,type:gist"`
	OwnerID     *string      `gorm:"type:uuid"`
	GID         string       `gorm:"-"`                                     // Legacy Alias for OwnerID
	Name        string       `gorm:"size:100;not null;column:project_name"` // Map to existing project_name
	ProjectName string       `gorm:"-"`                                     // Legacy Alias
	Description string       `gorm:"type:text"`
	CreatedAt   time.Time    `gorm:"autoCreateTime;column:create_at"` // Map to existing create_at
	CreateAt    time.Time    `gorm:"-"`                               // Legacy Alias

	// Legacy fields used by application layer
	GPUQuota                 int `gorm:"column:gpu_quota;default:0"`
	MaxConcurrentJobsPerUser int `gorm:"column:max_concurrent_jobs_per_user;default:0"`
	MaxQueuedJobsPerUser     int `gorm:"column:max_queued_jobs_per_user;default:0"`
	MaxJobRuntimeSeconds     int `gorm:"column:max_job_runtime_seconds;default:0"`
	MaxProjectUsers          int `gorm:"column:max_project_users;default:0"`

	ResourcePlan ResourcePlan `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE"`
}

func (Project) TableName() string { return "projects" }

type HierarchyNode = Project

// ResourcePlan defines time-window based resource quotas.
type ResourcePlan struct {
	ID            string           `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ProjectID     string           `gorm:"type:uuid;not null;index"`
	WeekWindow    common.Int4Range `gorm:"type:int4range;not null"`
	GPULimit      int              `gorm:"not null;default:0"`
	CPULimitCores float64          `gorm:"not null;default:0"`
	MemoryLimitGB float64          `gorm:"not null;default:0"`
}

func (ResourcePlan) TableName() string { return "resource_plans" }
