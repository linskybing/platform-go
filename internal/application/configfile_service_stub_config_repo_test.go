package application_test

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubConfigFileRepo struct {
	store          func(ctx context.Context, projectID, authorID, message string, content []byte) (*configfile.ConfigCommit, error)
	getCommit      func(ctx context.Context, id string) (*configfile.ConfigCommit, error)
	deleteCommit   func(ctx context.Context, id string) error
	getBlob        func(ctx context.Context, hash string) (*configfile.ConfigBlob, error)
	getHistory     func(ctx context.Context, projectID string) ([]configfile.ConfigCommit, error)
	listAllCommits func(ctx context.Context) ([]configfile.ConfigCommit, error)
}

func (s *stubConfigFileRepo) Store(ctx context.Context, projectID, authorID, message string, content []byte) (*configfile.ConfigCommit, error) {
	if s.store != nil {
		return s.store(ctx, projectID, authorID, message, content)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubConfigFileRepo) GetHead(ctx context.Context, projectID string) (*configfile.ConfigCommit, error) {
	return nil, gorm.ErrRecordNotFound
}

func (s *stubConfigFileRepo) GetCommit(ctx context.Context, id string) (*configfile.ConfigCommit, error) {
	if s.getCommit != nil {
		return s.getCommit(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubConfigFileRepo) DeleteCommit(ctx context.Context, id string) error {
	if s.deleteCommit != nil {
		return s.deleteCommit(ctx, id)
	}
	return nil
}

func (s *stubConfigFileRepo) GetBlob(ctx context.Context, hash string) (*configfile.ConfigBlob, error) {
	if s.getBlob != nil {
		return s.getBlob(ctx, hash)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubConfigFileRepo) GetHistory(ctx context.Context, projectID string) ([]configfile.ConfigCommit, error) {
	if s.getHistory != nil {
		return s.getHistory(ctx, projectID)
	}
	return nil, nil
}

func (s *stubConfigFileRepo) ListAllCommits(ctx context.Context) ([]configfile.ConfigCommit, error) {
	if s.listAllCommits != nil {
		return s.listAllCommits(ctx)
	}
	return nil, nil
}

func (s *stubConfigFileRepo) WithTx(tx *gorm.DB) repository.ConfigFileRepo {
	return s
}
