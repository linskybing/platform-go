package configfile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

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

func (s *ConfigFileService) CreateConfigFile(ctx context.Context, cf configfile.CreateConfigFileInput, claims *types.Claims) (*configfile.ConfigFile, error) {
	// Performance: Parse and validate BEFORE opening a DB transaction
	resourcesToCreate, err := s.parseAndValidateResources(cf.RawYaml)
	if err != nil {
		return nil, err
	}

	tx := s.Repos.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	createdCF := &configfile.ConfigFile{
		Filename:  cf.Filename,
		Content:   cf.RawYaml,
		ProjectID: cf.ProjectID,
	}

	if err := s.Repos.ConfigFile.WithTx(tx).CreateConfigFile(createdCF); err != nil {
		tx.Rollback()
		return nil, err
	}

	for _, res := range resourcesToCreate {
		res.CFID = createdCF.CFID
		if err := s.Repos.Resource.WithTx(tx).CreateResource(res); err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create resource %s/%s: %w", res.Type, res.Name, err)
		}
	}

	if res := tx.Commit(); res.Error != nil {
		return nil, fmt.Errorf("transaction commit failed: %w", res.Error)
	}

	// Audit logging (async)
	go func() {
		utils.LogAudit(claims.UserID, "", "", "create", "config_file",
			fmt.Sprintf("cf_id=%s", createdCF.CFID), nil, *createdCF, "", s.Repos.Audit)
	}()
	s.invalidateConfigFileCache(createdCF.CFID, createdCF.ProjectID)

	return createdCF, nil
}

func (s *ConfigFileService) UpdateConfigFile(ctx context.Context, id string, input configfile.ConfigFileUpdateDTO, claims *types.Claims) (*configfile.ConfigFile, error) {
	existing, err := s.Repos.ConfigFile.GetConfigFileByID(id)
	if err != nil {
		return nil, ErrConfigFileNotFound
	}

	oldCF := *existing

	if input.Filename != nil {
		existing.Filename = *input.Filename
	}

	if input.RawYaml != nil {
		// Prepare new resources first
		newResources, err := s.parseAndValidateResources(*input.RawYaml)
		if err != nil {
			return nil, err
		}

		// Use helper to handle the diff logic (delete old, create/update new)
		// We pass the parsed resources to avoid re-parsing inside the helper
		if err = s.syncConfigFileResources(ctx, existing, *input.RawYaml, newResources, claims); err != nil {
			return nil, err
		}
		existing.Content = *input.RawYaml
	}

	err = s.Repos.ConfigFile.UpdateConfigFile(existing)
	if err != nil {
		return nil, err
	}
	s.invalidateConfigFileCache(existing.CFID, existing.ProjectID)

	// Audit logging (sync for consistency)
	utils.LogAudit(claims.UserID, "", "", "update", "config_file",
		fmt.Sprintf("cf_id=%s", existing.CFID), oldCF, *existing, "", s.Repos.Audit)

	return existing, nil
}

func (s *ConfigFileService) DeleteConfigFile(ctx context.Context, id string, claims *types.Claims) error {
	cf, err := s.Repos.ConfigFile.GetConfigFileByID(id)
	if err != nil {
		return ErrConfigFileNotFound
	}

	// 1. Clean up K8s resources
	if err := s.DeleteConfigFileInstance(id); err != nil {
		// Log warning but proceed to delete DB records if possible, or return error depending on policy
		slog.Warn("failed to cleanup K8s resources for config file",
			"config_id", id,
			"error", err)
	}

	// 2. Clean up DB resources
	resources, err := s.Repos.Resource.ListResourcesByConfigFileID(id)
	if err != nil {
		return err
	}

	for _, res := range resources {
		if err := s.Repos.Resource.DeleteResource(res.RID); err != nil {
			return err
		}
	}

	// 3. Delete ConfigFile
	if err := s.Repos.ConfigFile.DeleteConfigFile(id); err != nil {
		return err
	}
	s.invalidateConfigFileCache(cf.CFID, cf.ProjectID)

	// Audit logging
	utils.LogAudit(claims.UserID, "", "", "delete", "config_file",
		fmt.Sprintf("cf_id=%s", cf.CFID), *cf, nil, "", s.Repos.Audit)
	return nil
}

// ValidateAndInjectGPUConfig is a thin compatibility wrapper used by unit tests.
// It unmarshals a JSON object, runs GPU validation/injection on any PodSpecs,
// and returns the resulting JSON bytes.
func (s *ConfigFileService) ValidateAndInjectGPUConfig(jsonBytes []byte, proj project.Project) ([]byte, error) {
	var obj map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &obj); err != nil {
		return nil, err
	}

	podSpecs := findPodSpecs(obj)
	if len(podSpecs) == 0 {
		return jsonBytes, nil
	}

	for _, spec := range podSpecs {
		if err := s.patchGPU(spec, proj); err != nil {
			return nil, err
		}
	}

	out, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return out, nil
}
