package view

import (
	"github.com/linskybing/platform-go/internal/domain/common"
)

type ProjectUserView struct {
	PID         string  `gorm:"column:p_id"`
	ProjectName string  `gorm:"column:project_name"`
	GID         string  `gorm:"column:g_id"`
	GroupName   string  `gorm:"column:group_name"`
	Role        *string `gorm:"column:role"` // Role in group or project
}

// ProjectDetailView aggregates Node and ResourcePlan data.
type ProjectDetailView struct {
	ID            string           `json:"id"`
	ParentID      *string          `json:"parent_id,omitempty"`
	Path          common.Ltree     `json:"path"`
	Name          string           `json:"name"`
	Description   string           `json:"description"`
	GPULimit      int              `json:"gpu_limit"`
	CPULimit      float64          `json:"cpu_limit"`
	MemoryLimitGB float64          `json:"memory_limit_gb"`
	WeekWindow    common.Int4Range `json:"week_window"`
}
