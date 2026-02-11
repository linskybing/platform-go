package configfile

import (
	"context"
	"log/slog"
	"time"

	"github.com/linskybing/platform-go/internal/application/executor"
	"github.com/linskybing/platform-go/internal/application/image"
	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
)

type ConfigFileService struct {
	Repos        *repository.Repos
	imageService *image.ImageService
	cache        *cache.Service
	executor     executor.Executor
}

func NewConfigFileService(repos *repository.Repos) *ConfigFileService {
	return NewConfigFileServiceWithCache(repos, nil)
}

func NewConfigFileServiceWithCache(repos *repository.Repos, cacheSvc *cache.Service) *ConfigFileService {
	return &ConfigFileService{
		Repos:        repos,
		imageService: image.NewImageService(repos.Image),
		cache:        cacheSvc,
		executor:     nil, // Will use direct K8s deployment if nil
	}
}

func NewConfigFileServiceWithExecutor(repos *repository.Repos, cacheSvc *cache.Service, exec executor.Executor) *ConfigFileService {
	return &ConfigFileService{
		Repos:        repos,
		imageService: image.NewImageService(repos.Image),
		cache:        cacheSvc,
		executor:     exec,
	}
}

// GetExecutor returns the executor used by this service
func (s *ConfigFileService) GetExecutor() executor.Executor {
	return s.executor
}

const configFileCacheTTL = 5 * time.Minute

func (s *ConfigFileService) ListConfigFiles() ([]configfile.ConfigFile, error) {
	if s.cache != nil && s.cache.Enabled() {
		var cached []configfile.ConfigFile
		if err := s.cache.GetJSON(context.Background(), configFileListKey(), &cached); err == nil {
			return cached, nil
		}
	}

	files, err := s.Repos.ConfigFile.ListConfigFiles()
	if err != nil {
		return nil, err
	}
	if s.cache != nil && s.cache.Enabled() {
		if err := s.cache.AsyncSetJSON(context.Background(), configFileListKey(), files, configFileCacheTTL); err != nil {
			slog.Warn("failed to cache config file list",
				"error", err)
		}
	}
	return files, nil
}

func (s *ConfigFileService) GetConfigFile(id string) (*configfile.ConfigFile, error) {
	if s.cache != nil && s.cache.Enabled() {
		var cached configfile.ConfigFile
		if err := s.cache.GetJSON(context.Background(), configFileByIDKey(id), &cached); err == nil {
			return &cached, nil
		}
	}

	cf, err := s.Repos.ConfigFile.GetConfigFileByID(id)
	if err != nil {
		return nil, err
	}
	if s.cache != nil && s.cache.Enabled() {
		if err := s.cache.AsyncSetJSON(context.Background(), configFileByIDKey(id), cf, configFileCacheTTL); err != nil {
			slog.Warn("failed to cache config file",
				"config_id", id,
				"error", err)
		}
	}
	return cf, nil
}

func (s *ConfigFileService) ListConfigFilesByProjectID(projectID string) ([]configfile.ConfigFile, error) {
	if s.cache != nil && s.cache.Enabled() {
		var cached []configfile.ConfigFile
		if err := s.cache.GetJSON(context.Background(), configFileByProjectKey(projectID), &cached); err == nil {
			return cached, nil
		}
	}

	files, err := s.Repos.ConfigFile.GetConfigFilesByProjectID(projectID)
	if err != nil {
		return nil, err
	}
	if s.cache != nil && s.cache.Enabled() {
		if err := s.cache.AsyncSetJSON(context.Background(), configFileByProjectKey(projectID), files, configFileCacheTTL); err != nil {
			slog.Warn("failed to cache project config files",
				"project_id", projectID,
				"error", err)
		}
	}
	return files, nil
}
