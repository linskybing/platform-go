package repository

import (
	"fmt"

	"github.com/linskybing/platform-go/internal/domain/configfile"
	"gorm.io/gorm"
)

type ConfigFileRepo interface {
	CreateConfigFile(cf *configfile.ConfigFile) error
	GetConfigFileByID(id string) (*configfile.ConfigFile, error)
	UpdateConfigFile(cf *configfile.ConfigFile) error
	DeleteConfigFile(id string) error
	ListConfigFiles() ([]configfile.ConfigFile, error)
	GetConfigFilesByProjectID(projectID string) ([]configfile.ConfigFile, error)
	GetGroupIDByConfigFileID(cfID string) (string, error)
	WithTx(tx *gorm.DB) ConfigFileRepo
}

type DBConfigFileRepo struct {
	db *gorm.DB
}

func NewConfigFileRepo(db *gorm.DB) *DBConfigFileRepo {
	return &DBConfigFileRepo{
		db: db,
	}
}

func (r *DBConfigFileRepo) CreateConfigFile(cf *configfile.ConfigFile) error {
	return r.db.Create(cf).Error
}

func (r *DBConfigFileRepo) GetConfigFileByID(id string) (*configfile.ConfigFile, error) {
	var cf configfile.ConfigFile
	if err := r.db.First(&cf, "cf_id = ?", id).Error; err != nil {
		return nil, err
	}
	return &cf, nil
}

func (r *DBConfigFileRepo) UpdateConfigFile(cf *configfile.ConfigFile) error {
	if cf.CFID == "" {
		return fmt.Errorf("missing ConfigFile ID: validation failed")
	}
	if err := r.db.Save(cf).Error; err != nil {
		return fmt.Errorf("failed to update config file: %w", err)
	}
	return nil
}

func (r *DBConfigFileRepo) DeleteConfigFile(id string) error {
	return r.db.Delete(&configfile.ConfigFile{}, "cf_id = ?", id).Error
}

func (r *DBConfigFileRepo) ListConfigFiles() ([]configfile.ConfigFile, error) {
	var list []configfile.ConfigFile
	if err := r.db.Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *DBConfigFileRepo) GetConfigFilesByProjectID(projectID string) ([]configfile.ConfigFile, error) {
	var files []configfile.ConfigFile
	if err := r.db.Where("project_id = ?", projectID).Find(&files).Error; err != nil {
		return nil, err
	}
	return files, nil
}

func (r *DBConfigFileRepo) GetGroupIDByConfigFileID(cfID string) (string, error) {
	var gID string
	err := r.db.Table("config_files cf").
		Select("p.g_id").
		Joins("JOIN project_list p ON cf.project_id = p.p_id").
		Where("cf.cf_id = ?", cfID).
		Scan(&gID).Error

	if err != nil {
		return "", err
	}
	return gID, nil
}

func (r *DBConfigFileRepo) WithTx(tx *gorm.DB) ConfigFileRepo {
	if tx == nil {
		return r
	}
	return &DBConfigFileRepo{
		db: tx,
	}
}
