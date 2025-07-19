package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/dto"
	"github.com/linskybing/platform-go/models"
	"github.com/linskybing/platform-go/repositories"
	"github.com/linskybing/platform-go/utils"
)

var (
	ErrConfigFileNotFound    = errors.New("config file not found")
	ErrYAMLParsingFailed     = errors.New("YAML parsing failed")
	ErrNoInvalidYAMLDocument = errors.New("no valid YAML documents found")
	ErrUploadYAMLFailed      = errors.New("failed to upload YAML file")
)

func ListConfigFiles() ([]models.ConfigFile, error) {
	return repositories.ListConfigFiles()
}

func GetConfigFile(id uint) (*models.ConfigFile, error) {
	return repositories.GetConfigFileByID(id)
}

func CreateConfigFile(c *gin.Context, cf dto.CreateConfigFileInput) (*models.ConfigFile, error) {

	filename := fmt.Sprintf("config_%d_%d.yaml", cf.ProjectID, time.Now().Unix())
	if err := utils.UploadObject(c, filename, "application/x-yaml", strings.NewReader(cf.RawYaml), int64(len(cf.RawYaml))); err != nil {
		return nil, ErrUploadYAMLFailed
	}

	createdCF := &models.ConfigFile{
		Filename:  filename,
		MinIOPath: fmt.Sprintf("config/%s", filename),
		ProjectID: cf.ProjectID,
	}

	if err := repositories.CreateConfigFile(createdCF); err != nil {
		return nil, err
	}

	// 	userID, _ := utils.GetUserIDFromContext(c)
	// _ = utils.LogAudit(c, userID, "update", "config_file", existing.CFID, oldCF, *existing, "")

	// yamlArray := utils.SplitYAMLDocuments(cf.RawYaml)
	// if len(yamlArray) == 0 {
	// 	return nil, ErrNoInvalidYAMLDocument
	// }

	// for _, doc := range yamlArray {
	// 	yamlContent, _ := utils.YAMLToJSON(doc)

	// }
	return nil, nil
}

func UpdateConfigFile(c *gin.Context, id uint, input dto.ConfigFileUpdateDTO) (*models.ConfigFile, error) {
	existing, err := repositories.GetConfigFileByID(id)
	if err != nil {
		return nil, ErrConfigFileNotFound
	}

	oldCF := *existing

	if input.Filename != nil {
		existing.Filename = *input.Filename
	}
	if input.MinIOPath != nil {
		existing.MinIOPath = *input.MinIOPath
	}
	if input.ProjectID != nil {
		existing.ProjectID = *input.ProjectID
	}

	err = repositories.UpdateConfigFile(existing)
	if err != nil {
		return nil, err
	}

	userID, _ := utils.GetUserIDFromContext(c)
	_ = utils.LogAudit(c, userID, "update", "config_file", existing.CFID, oldCF, *existing, "")

	return existing, nil
}

func DeleteConfigFile(c *gin.Context, id uint) error {
	cf, err := repositories.GetConfigFileByID(id)
	if err != nil {
		return ErrConfigFileNotFound
	}

	err = repositories.DeleteConfigFile(id)
	if err != nil {
		return err
	}

	userID, _ := utils.GetUserIDFromContext(c)
	_ = utils.LogAudit(c, userID, "delete", "config_file", cf.CFID, *cf, nil, "")

	return nil
}

func ListConfigFilesByProjectID(projectID uint) ([]models.ConfigFile, error) {
	return repositories.GetConfigFilesByProjectID(projectID)
}
