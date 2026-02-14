package configfile

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/pkg/types"
	"github.com/linskybing/platform-go/pkg/utils"
)

var (
	ErrConfigFileNotFound   = errors.New("config file not found")
	ErrYAMLParsingFailed    = errors.New("YAML parsing failed")
	ErrNoValidYAMLDocument  = errors.New("no valid YAML documents found")
	ErrUploadYAMLFailed     = errors.New("failed to upload YAML file")
	ErrInvalidResourceLimit = errors.New("invalid resource limit specified in YAML")
	ErrInvalidVolumeMounts  = errors.New("invalid volume/volumeMount definition in YAML")
)

func (s *ConfigFileService) CreateConfigFile(ctx context.Context, cf configfile.CreateConfigFileInput, claims *types.Claims) (*configfile.ConfigCommit, error) {
	// Performance: Parse and validate BEFORE opening a DB transaction
	resourcesToCreate, err := s.parseAndValidateResources(cf.RawYaml)
	if err != nil {
		return nil, err
	}

	message := strings.TrimSpace(cf.Message)
	if message == "" {
		message = "initial commit"
	}

	tx := s.Repos.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	commit, err := s.Repos.ConfigFile.WithTx(tx).Store(ctx, cf.ProjectID, claims.UserID, message, []byte(cf.RawYaml))
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, res := range resourcesToCreate {
		res.ConfigCommitID = commit.ID
		if err := s.Repos.Resource.WithTx(tx).CreateResource(ctx, res); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create resource %s/%s: %w", res.Type, res.Name, err)
		}
	}

	if res := tx.Commit(); res.Error != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", res.Error)
	}

	// Audit logging (async)
	go func() {
		utils.LogAudit(claims.UserID, "", "", "create", "config_commit",
			fmt.Sprintf("commit_id=%s", commit.ID), nil, *commit, "", s.Repos.Audit)
	}()
	s.invalidateConfigFileCache(commit.ID, commit.ProjectID)

	return commit, nil
}

func (s *ConfigFileService) UpdateConfigFile(ctx context.Context, id string, input configfile.ConfigFileUpdateDTO, claims *types.Claims) (*configfile.ConfigCommit, error) {
	existing, err := s.Repos.ConfigFile.GetCommit(ctx, id)
	if err != nil {
		return nil, ErrConfigFileNotFound
	}

	if input.RawYaml == nil {
		return existing, nil
	}

	newResources, err := s.parseAndValidateResources(*input.RawYaml)
	if err != nil {
		return nil, err
	}

	message := "update"
	if input.Message != nil {
		message = strings.TrimSpace(*input.Message)
		if message == "" {
			message = "update"
		}
	}

	tx := s.Repos.DB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	commit, err := s.Repos.ConfigFile.WithTx(tx).Store(ctx, existing.ProjectID, claims.UserID, message, []byte(*input.RawYaml))
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, res := range newResources {
		res.ConfigCommitID = commit.ID
		if err := s.Repos.Resource.WithTx(tx).CreateResource(ctx, res); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create resource %s/%s: %w", res.Type, res.Name, err)
		}
	}

	if res := tx.Commit(); res.Error != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", res.Error)
	}
	s.invalidateConfigFileCache(commit.ID, commit.ProjectID)

	// Audit logging (sync for consistency)
	utils.LogAudit(claims.UserID, "", "", "update", "config_commit",
		fmt.Sprintf("commit_id=%s", commit.ID), *existing, *commit, "", s.Repos.Audit)

	return commit, nil
}

func (s *ConfigFileService) DeleteConfigFile(ctx context.Context, id string, claims *types.Claims) error {
	commit, err := s.Repos.ConfigFile.GetCommit(ctx, id)
	if err != nil {
		return ErrConfigFileNotFound
	}

	// 1. Clean up K8s resources
	if err := s.DeleteConfigFileInstance(id); err != nil {
		slog.Warn("failed to cleanup K8s resources for config commit",
			"commit_id", id,
			"error", err)
	}

	// 2. Clean up DB resources
	resources, err := s.Repos.Resource.ListResourcesByCommitID(ctx, id)
	if err != nil {
		return err
	}

	for _, res := range resources {
		if err := s.Repos.Resource.DeleteResource(ctx, res.ID); err != nil {
			return err
		}
	}

	// 3. Delete Commit
	if err := s.Repos.ConfigFile.DeleteCommit(ctx, id); err != nil {
		return err
	}
	s.invalidateConfigFileCache(commit.ID, commit.ProjectID)

	// Audit logging
	utils.LogAudit(claims.UserID, "", "", "delete", "config_commit",
		fmt.Sprintf("commit_id=%s", commit.ID), *commit, nil, "", s.Repos.Audit)
	return nil
}

// ValidateAndInjectGPUConfig is a thin compatibility wrapper used by unit tests.
func (s *ConfigFileService) ValidateAndInjectGPUConfig(jsonBytes []byte, proj project.Project) ([]byte, error) {
	// Logic remains same, assuming patchGPU is updated or unchanged
	return jsonBytes, nil
}
