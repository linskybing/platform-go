package project

type CreateProjectDTO struct {
	ProjectName              string            `json:"project_name" form:"project_name" binding:"required"`
	Description              *string           `json:"description,omitempty" form:"description,omitempty"`
	GID                      string            `json:"g_id" form:"g_id" binding:"required"`
	GPUQuota                 *int              `json:"gpu_quota,omitempty" form:"gpu_quota,omitempty"` // GPU quota in integer units
	MaxConcurrentJobsPerUser *int              `json:"max_concurrent_jobs_per_user,omitempty" form:"max_concurrent_jobs_per_user,omitempty"`
	MaxQueuedJobsPerUser     *int              `json:"max_queued_jobs_per_user,omitempty" form:"max_queued_jobs_per_user,omitempty"`
	MaxJobRuntimeSeconds     *int              `json:"max_job_runtime_seconds,omitempty" form:"max_job_runtime_seconds,omitempty"`
	MaxProjectUsers          *int              `json:"max_project_users,omitempty" form:"max_project_users,omitempty"`
	ScheduleWindows          *[]ScheduleWindow `json:"schedule_windows,omitempty" form:"schedule_windows,omitempty"`
}

type UpdateProjectDTO struct {
	ProjectName              *string           `json:"project_name,omitempty" form:"project_name,omitempty"`
	Description              *string           `json:"description,omitempty" form:"description,omitempty"`
	GID                      *string           `json:"g_id,omitempty" form:"g_id,omitempty"`
	GPUQuota                 *int              `json:"gpu_quota,omitempty" form:"gpu_quota,omitempty"` // GPU quota in integer units
	MaxConcurrentJobsPerUser *int              `json:"max_concurrent_jobs_per_user,omitempty" form:"max_concurrent_jobs_per_user,omitempty"`
	MaxQueuedJobsPerUser     *int              `json:"max_queued_jobs_per_user,omitempty" form:"max_queued_jobs_per_user,omitempty"`
	MaxJobRuntimeSeconds     *int              `json:"max_job_runtime_seconds,omitempty" form:"max_job_runtime_seconds,omitempty"`
	MaxProjectUsers          *int              `json:"max_project_users,omitempty" form:"max_project_users,omitempty"`
	ScheduleWindows          *[]ScheduleWindow `json:"schedule_windows,omitempty" form:"schedule_windows,omitempty"`
}

type CreateGroupPVCDTO struct {
	Name string `json:"name" binding:"required"`
	Size string `json:"size" binding:"required"`
}

type GIDGetter interface {
	GetGID() string
}

func (d CreateProjectDTO) GetGID() string {
	return d.GID
}
