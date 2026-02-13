package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/linskybing/platform-go/internal/domain/gpuusage"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GPUUsageRepo interface {
	InsertSnapshots(ctx context.Context, snapshots []gpuusage.JobGPUUsageSnapshot) error
	ListSnapshotsByJob(ctx context.Context, jobID string, limit, offset int) ([]gpuusage.JobGPUUsageSnapshot, int64, error)
	ListAllSnapshotsByJob(ctx context.Context, jobID string) ([]gpuusage.JobGPUUsageSnapshot, error)
	GetSummary(ctx context.Context, jobID string) (*gpuusage.JobGPUUsageSummary, error)
	UpsertSummary(ctx context.Context, summary *gpuusage.JobGPUUsageSummary) error
	DeleteSnapshotsBefore(ctx context.Context, cutoff time.Time) error

	WithTx(tx *gorm.DB) GPUUsageRepo
}

type GPUUsageRepoImpl struct {
	db *gorm.DB
}

func NewGPUUsageRepo(db *gorm.DB) GPUUsageRepo {
	return &GPUUsageRepoImpl{db: db}
}

func (r *GPUUsageRepoImpl) InsertSnapshots(ctx context.Context, snapshots []gpuusage.JobGPUUsageSnapshot) error {
	if len(snapshots) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&snapshots).Error
}

func (r *GPUUsageRepoImpl) ListSnapshotsByJob(ctx context.Context, jobID string, limit, offset int) ([]gpuusage.JobGPUUsageSnapshot, int64, error) {
	var snapshots []gpuusage.JobGPUUsageSnapshot
	var total int64

	query := r.db.WithContext(ctx).Model(&gpuusage.JobGPUUsageSnapshot{}).Where("job_id = ?", jobID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count snapshots: %w", err)
	}

	if limit <= 0 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	err := query.Order("timestamp ASC").Limit(limit).Offset(offset).Find(&snapshots).Error
	if err != nil {
		return nil, 0, fmt.Errorf("list snapshots: %w", err)
	}
	return snapshots, total, nil
}

func (r *GPUUsageRepoImpl) ListAllSnapshotsByJob(ctx context.Context, jobID string) ([]gpuusage.JobGPUUsageSnapshot, error) {
	var snapshots []gpuusage.JobGPUUsageSnapshot
	err := r.db.WithContext(ctx).
		Where("job_id = ?", jobID).
		Order("timestamp ASC").
		Find(&snapshots).Error
	if err != nil {
		return nil, fmt.Errorf("list all snapshots: %w", err)
	}
	return snapshots, nil
}

func (r *GPUUsageRepoImpl) GetSummary(ctx context.Context, jobID string) (*gpuusage.JobGPUUsageSummary, error) {
	var summary gpuusage.JobGPUUsageSummary
	err := r.db.WithContext(ctx).First(&summary, "job_id = ?", jobID).Error
	if err != nil {
		return nil, err
	}
	return &summary, nil
}

func (r *GPUUsageRepoImpl) UpsertSummary(ctx context.Context, summary *gpuusage.JobGPUUsageSummary) error {
	if summary == nil {
		return errors.New("summary is nil")
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "job_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"total_gpu_seconds", "peak_memory_bytes", "avg_utilization", "avg_memory_bytes", "sample_count", "first_sample_at", "last_sample_at", "computed_at"}),
	}).Create(summary).Error
}

func (r *GPUUsageRepoImpl) DeleteSnapshotsBefore(ctx context.Context, cutoff time.Time) error {
	return r.db.WithContext(ctx).
		Where("timestamp < ?", cutoff).
		Delete(&gpuusage.JobGPUUsageSnapshot{}).Error
}

func (r *GPUUsageRepoImpl) WithTx(tx *gorm.DB) GPUUsageRepo {
	if tx == nil {
		return r
	}
	return &GPUUsageRepoImpl{db: tx}
}
