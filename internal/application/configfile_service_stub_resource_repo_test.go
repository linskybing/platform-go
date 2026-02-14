package application_test

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/audit"
	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubResourceRepo struct {
	createResource           func(ctx context.Context, res *resource.Resource) error
	listResourcesByCommitID  func(ctx context.Context, commitID string) ([]resource.Resource, error)
	deleteResource           func(ctx context.Context, rid string) error
	getResourceByCommitIDAnd func(ctx context.Context, commitID string, name string) (*resource.Resource, error)
}

func (s *stubResourceRepo) CreateResource(ctx context.Context, res *resource.Resource) error {
	if s.createResource != nil {
		return s.createResource(ctx, res)
	}
	return nil
}

func (s *stubResourceRepo) GetResourceByID(ctx context.Context, rid string) (*resource.Resource, error) {
	return nil, gorm.ErrRecordNotFound
}

func (s *stubResourceRepo) UpdateResource(ctx context.Context, res *resource.Resource) error {
	return nil
}

func (s *stubResourceRepo) DeleteResource(ctx context.Context, rid string) error {
	if s.deleteResource != nil {
		return s.deleteResource(ctx, rid)
	}
	return nil
}

func (s *stubResourceRepo) ListResourcesByProjectID(ctx context.Context, pid string) ([]resource.Resource, error) {
	return nil, nil
}

func (s *stubResourceRepo) ListResourcesByCommitID(ctx context.Context, commitID string) ([]resource.Resource, error) {
	if s.listResourcesByCommitID != nil {
		return s.listResourcesByCommitID(ctx, commitID)
	}
	return nil, nil
}

func (s *stubResourceRepo) GetResourceByCommitIDAndName(ctx context.Context, commitID string, name string) (*resource.Resource, error) {
	if s.getResourceByCommitIDAnd != nil {
		return s.getResourceByCommitIDAnd(ctx, commitID, name)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubResourceRepo) GetGroupIDByResourceID(ctx context.Context, rID string) (string, error) {
	return "", nil
}

func (s *stubResourceRepo) WithTx(tx *gorm.DB) repository.ResourceRepo {
	return s
}

type stubAuditRepo struct {
	createAuditLog func(a *audit.AuditLog) error
}

func (s *stubAuditRepo) GetAuditLogs(params repository.AuditQueryParams) ([]audit.AuditLog, error) {
	return nil, nil
}

func (s *stubAuditRepo) CreateAuditLog(a *audit.AuditLog) error {
	if s.createAuditLog != nil {
		return s.createAuditLog(a)
	}
	return nil
}

func (s *stubAuditRepo) DeleteOldAuditLogs(retentionDays int) error {
	return nil
}

func (s *stubAuditRepo) WithTx(tx *gorm.DB) repository.AuditRepo {
	return s
}
