package websocket

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/job"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubJobRepo struct {
	get func(ctx context.Context, id string) (*job.Job, error)
}

func (s *stubJobRepo) Create(ctx context.Context, j *job.Job) error {
	return nil
}

func (s *stubJobRepo) Get(ctx context.Context, id string) (*job.Job, error) {
	if s.get != nil {
		return s.get(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubJobRepo) UpdateStatus(ctx context.Context, id string, status string, errorMsg *string) error {
	return nil
}

func (s *stubJobRepo) ListByProject(ctx context.Context, projectID string) ([]job.Job, error) {
	return nil, nil
}

func (s *stubJobRepo) ListByUser(ctx context.Context, userID string) ([]job.Job, error) {
	return nil, nil
}

func (s *stubJobRepo) ListByStatus(ctx context.Context, statuses []string) ([]job.Job, error) {
	return nil, nil
}

func (s *stubJobRepo) ListByProjectAndStatuses(ctx context.Context, projectID string, statuses []string) ([]job.Job, error) {
	return nil, nil
}

func (s *stubJobRepo) CountByUserProjectAndStatuses(ctx context.Context, userID, projectID string, statuses []string) (int64, error) {
	return 0, nil
}

func (s *stubJobRepo) FetchNextPending(ctx context.Context) (*job.Job, error) {
	return nil, gorm.ErrRecordNotFound
}

func (s *stubJobRepo) WithTx(tx *gorm.DB) repository.JobRepo {
	return s
}
