package configfile

import (
	"context"
	"fmt"

	"github.com/linskybing/platform-go/internal/application/executor"
	"github.com/linskybing/platform-go/internal/application/image"
	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
)

type ConfigFileService struct {
	Repos    *repository.Repos
	cache    *cache.Service
	executor executor.Executor
	imageSvc *image.ImageService
}

func NewConfigFileService(repos *repository.Repos) *ConfigFileService {
	return NewConfigFileServiceWithExecutor(repos, nil, nil)
}

func NewConfigFileServiceWithExecutor(repos *repository.Repos, cacheSvc *cache.Service, exec executor.Executor) *ConfigFileService {
	var imageSvc *image.ImageService
	if repos != nil && repos.Image != nil {
		imageSvc = image.NewImageService(repos.Image)
	}
	return &ConfigFileService{
		Repos:    repos,
		cache:    cacheSvc,
		executor: exec,
		imageSvc: imageSvc,
	}
}

// GetLatestConfig retrieves the current configuration state for a project.
func (s *ConfigFileService) GetLatestConfig(projectID string) (*configfile.ConfigBlob, error) {
	commit, err := s.Repos.ConfigFile.GetHead(context.Background(), projectID)
	if err != nil {
		return nil, err
	}
	return s.Repos.ConfigFile.GetBlob(context.Background(), commit.BlobHash)
}

// UpdateConfig creates a new version of the configuration.
func (s *ConfigFileService) UpdateConfig(projectID, authorID, message string, content []byte) (*configfile.ConfigCommit, error) {
	resourcesToCreate, err := s.parseAndValidateResources(string(content))
	if err != nil {
		return nil, err
	}

	tx := s.Repos.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	commit, err := s.Repos.ConfigFile.WithTx(tx).Store(context.Background(), projectID, authorID, message, content)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update config: %w", err)
	}

	for _, res := range resourcesToCreate {
		res.ConfigCommitID = commit.ID
		if err := s.Repos.Resource.WithTx(tx).CreateResource(context.Background(), res); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create resource %s/%s: %w", res.Type, res.Name, err)
		}
	}

	if res := tx.Commit(); res.Error != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", res.Error)
	}

	return commit, nil
}

// GetHistory retrieves the commit history for a project.
func (s *ConfigFileService) GetHistory(projectID string) ([]configfile.ConfigCommit, error) {
	return s.Repos.ConfigFile.GetHistory(context.Background(), projectID)
}

// ListConfigFiles returns all config commits.
func (s *ConfigFileService) ListConfigFiles() ([]configfile.ConfigCommit, error) {
	return s.Repos.ConfigFile.ListAllCommits(context.Background())
}

// GetConfigFile returns a single config commit by ID.
func (s *ConfigFileService) GetConfigFile(id string) (*configfile.ConfigCommit, error) {
	return s.Repos.ConfigFile.GetCommit(context.Background(), id)
}

// ListConfigFilesByProjectID returns commit history for a project.
func (s *ConfigFileService) ListConfigFilesByProjectID(projectID string) ([]configfile.ConfigCommit, error) {
	return s.Repos.ConfigFile.GetHistory(context.Background(), projectID)
}

func (s *ConfigFileService) GetExecutor() executor.Executor {
	return s.executor
}
