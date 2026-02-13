package repository

import (
	"context"
	"fmt"

	"github.com/linskybing/platform-go/internal/domain/job"
	"gorm.io/gorm"
)

// JobRepo defines job data access interface
type JobRepo interface {
	Create(ctx context.Context, j *job.Job) error
	Get(ctx context.Context, id string) (*job.Job, error)
	UpdateStatus(ctx context.Context, id string, status string, errorMsg *string) error
	ListByProject(ctx context.Context, projectID string) ([]job.Job, error)
	ListByUser(ctx context.Context, userID string) ([]job.Job, error)
	ListByStatus(ctx context.Context, statuses []string) ([]job.Job, error)
	ListByProjectAndStatuses(ctx context.Context, projectID string, statuses []string) ([]job.Job, error)
	CountByUserProjectAndStatuses(ctx context.Context, userID, projectID string, statuses []string) (int64, error)

	WithTx(tx *gorm.DB) JobRepo
}

// JobRepoImpl implements JobRepo
type JobRepoImpl struct {
	db *gorm.DB
}

// NewJobRepo creates a new JobRepo
func NewJobRepo(db *gorm.DB) JobRepo {
	return &JobRepoImpl{db: db}
}

// Create creates a new job record
func (r *JobRepoImpl) Create(ctx context.Context, j *job.Job) error {
	if err := r.db.WithContext(ctx).Create(j).Error; err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}
	return nil
}

// Get retrieves a job by ID
func (r *JobRepoImpl) Get(ctx context.Context, id string) (*job.Job, error) {
	var j job.Job
	err := r.db.WithContext(ctx).First(&j, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &j, nil
}

// UpdateStatus updates job status and optionally error message
func (r *JobRepoImpl) UpdateStatus(ctx context.Context, id string, status string, errorMsg *string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMsg != nil {
		updates["error_message"] = *errorMsg
	}

	return r.db.WithContext(ctx).
		Model(&job.Job{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// ListByProject retrieves all jobs for a project
func (r *JobRepoImpl) ListByProject(ctx context.Context, projectID string) ([]job.Job, error) {
	var jobs []job.Job
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Order("submitted_at DESC").
		Find(&jobs).Error
	return jobs, err
}

// ListByUser retrieves all jobs for a user
func (r *JobRepoImpl) ListByUser(ctx context.Context, userID string) ([]job.Job, error) {
	var jobs []job.Job
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("submitted_at DESC").
		Find(&jobs).Error
	return jobs, err
}

// ListByStatus retrieves jobs with any of the given statuses
func (r *JobRepoImpl) ListByStatus(ctx context.Context, statuses []string) ([]job.Job, error) {
	var jobs []job.Job
	if len(statuses) == 0 {
		return jobs, nil
	}
	err := r.db.WithContext(ctx).
		Where("status IN ?", statuses).
		Order("submitted_at DESC").
		Find(&jobs).Error
	return jobs, err
}

// ListByProjectAndStatuses retrieves jobs in a project with any of the given statuses.
func (r *JobRepoImpl) ListByProjectAndStatuses(ctx context.Context, projectID string, statuses []string) ([]job.Job, error) {
	var jobs []job.Job
	if len(statuses) == 0 {
		return jobs, nil
	}
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Where("status IN ?", statuses).
		Order("submitted_at DESC").
		Find(&jobs).Error
	return jobs, err
}

// CountByUserProjectAndStatuses counts jobs for a user in a project with any of the given statuses.
func (r *JobRepoImpl) CountByUserProjectAndStatuses(ctx context.Context, userID, projectID string, statuses []string) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&job.Job{}).
		Where("user_id = ?", userID).
		Where("project_id = ?", projectID)
	if len(statuses) > 0 {
		query = query.Where("status IN ?", statuses)
	}
	if err := query.Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (r *JobRepoImpl) WithTx(tx *gorm.DB) JobRepo {
	if tx == nil {
		return r
	}
	return &JobRepoImpl{db: tx}
}
