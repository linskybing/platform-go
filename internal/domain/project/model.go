package project

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/linskybing/platform-go/internal/domain/common"
	"gorm.io/gorm"
)

// Project represents a node in the project tree (Project or SubProject).
type Project struct {
	ID          string       `gorm:"primaryKey;type:uuid;column:p_id;default:uuid_generate_v4()" json:"id"`
	ParentID    *string      `gorm:"type:uuid" json:"parent_id"`
	Path        common.Ltree `gorm:"type:ltree;not null;index:idx_path,type:gist" json:"path"`
	OwnerID     *string      `gorm:"type:uuid" json:"owner_id"`
	Name        string       `gorm:"size:100;not null;column:project_name" json:"name"`
	Description string       `gorm:"type:text" json:"description"`
	CreatedAt   time.Time    `gorm:"autoCreateTime;column:create_at" json:"created_at"`

	// Quotas
	GPUQuota                 int `gorm:"column:gpu_quota;default:0" json:"gpu_quota"`
	MaxConcurrentJobsPerUser int `gorm:"column:max_concurrent_jobs_per_user;default:0" json:"max_concurrent_jobs_per_user"`
	MaxQueuedJobsPerUser     int `gorm:"column:max_queued_jobs_per_user;default:0" json:"max_queued_jobs_per_user"`
	MaxJobRuntimeSeconds     int `gorm:"column:max_job_runtime_seconds;default:0" json:"max_job_runtime_seconds"`
	MaxProjectUsers          int `gorm:"column:max_project_users;default:0" json:"max_project_users"`

	ResourcePlan ResourcePlan `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"resource_plan"`
}

func (Project) TableName() string { return "projects" }

func (p *Project) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}

	// Initialize path if not set
	if p.Path == "" {
		if p.ParentID != nil {
			var parent Project
			if err := tx.First(&parent, "p_id = ?", *p.ParentID).Error; err != nil {
				return err
			}
			p.Path = common.Ltree(fmt.Sprintf("%s.%s", parent.Path, p.ID))
		} else {
			p.Path = common.Ltree(p.ID)
		}
	}
	// Sanitize path (replace - with _) because ltree doesn't like hyphens
	p.Path = common.Ltree(sanitizePath(string(p.Path)))

	return nil
}

func sanitizePath(path string) string {
	// Simple sanitization for ltree: replace - with _
	return strings.ReplaceAll(path, "-", "_")
}

type HierarchyNode = Project

// ResourcePlan defines time-window based resource quotas.
type ResourcePlan struct {
	ID            string           `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	ProjectID     string           `gorm:"type:uuid;not null;index" json:"project_id"`
	WeekWindow    common.Int4Range `gorm:"type:int4range;not null" json:"week_window"`
	GPULimit      int              `gorm:"not null;default:0" json:"gpu_limit"`
	CPULimitCores float64          `gorm:"not null;default:0" json:"cpu_limit_cores"`
	MemoryLimitGB float64          `gorm:"not null;default:0" json:"memory_limit_gb"`
}

func (ResourcePlan) TableName() string { return "resource_plans" }
