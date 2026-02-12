package repository

import (
	"errors"
	"fmt"

	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/internal/domain/view"
	"gorm.io/gorm"
)

type ResourceRepo interface {
	resource.Repository
	GetProjectResourcesByGroupID(groupID string) ([]view.ProjectResourceView, error)
	GetGroupResourcesByGroupID(groupID string) ([]view.GroupResourceView, error)
	WithTx(tx *gorm.DB) ResourceRepo
}

type DBResourceRepo struct {
	db *gorm.DB
}

func NewResourceRepo(db *gorm.DB) *DBResourceRepo {
	return &DBResourceRepo{
		db: db,
	}
}

func (r *DBResourceRepo) CreateResource(resource *resource.Resource) error {
	return r.db.Create(resource).Error
}

func (r *DBResourceRepo) GetResourceByID(rid string) (*resource.Resource, error) {
	var resource resource.Resource
	err := r.db.First(&resource, "r_id = ?", rid).Error
	if err != nil {
		return nil, err
	}
	return &resource, nil
}

func (r *DBResourceRepo) UpdateResource(resource *resource.Resource) error {
	if resource.RID == "" {
		return fmt.Errorf("resource RID is required: validation failed")
	}
	if err := r.db.Save(resource).Error; err != nil {
		return fmt.Errorf("failed to update resource: %w", err)
	}
	return nil
}

func (r *DBResourceRepo) DeleteResource(rid string) error {
	return r.db.Delete(&resource.Resource{}, "r_id = ?", rid).Error
}

func (r *DBResourceRepo) ListResourcesByProjectID(pid string) ([]resource.Resource, error) {
	var resources []resource.Resource
	err := r.db.
		Joins("JOIN config_files cf ON cf.cf_id = resources.cf_id").
		Where("cf.project_id = ?", pid).
		Find(&resources).Error
	return resources, err
}

func (r *DBResourceRepo) ListResourcesByConfigFileID(cfID string) ([]resource.Resource, error) {
	var resources []resource.Resource
	err := r.db.
		Where("cf_id = ?", cfID).
		Find(&resources).Error
	return resources, err
}

func (r *DBResourceRepo) GetResourceByConfigFileIDAndName(cfID string, name string) (*resource.Resource, error) {
	var resource resource.Resource
	err := r.db.
		Where("cf_id = ? AND name = ?", cfID, name).
		First(&resource).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &resource, nil
}

func (r *DBResourceRepo) GetProjectResourcesByGroupID(groupID string) ([]view.ProjectResourceView, error) {
	var results []view.ProjectResourceView

	err := r.db.Table("project_list p").
		Select(`
            p.p_id, p.project_name, 
            r.r_id, r.type, r.name, 
            cf.filename, r.create_at AS resource_create_at
        `).
		Joins("JOIN config_files cf ON cf.project_id = p.p_id").
		Joins("JOIN resource_list r ON r.cf_id = cf.cf_id").
		Where("p.g_id = ?", groupID).
		Scan(&results).Error

	return results, err
}

func (r *DBResourceRepo) GetGroupResourcesByGroupID(groupID string) ([]view.GroupResourceView, error) {
	var results []view.GroupResourceView

	err := r.db.Table("group_list g").
		Select(`
            g.g_id, g.group_name, 
            p.p_id, p.project_name, 
            r.r_id, r.type AS resource_type, r.name AS resource_name, 
            cf.filename, r.create_at AS resource_create_at
        `).
		Joins("LEFT JOIN project_list p ON p.g_id = g.g_id").
		Joins("LEFT JOIN config_files cf ON cf.project_id = p.p_id").
		Joins("LEFT JOIN resource_list r ON r.cf_id = cf.cf_id").
		Where("g.g_id = ? AND r.r_id IS NOT NULL", groupID).
		Scan(&results).Error

	return results, err
}

func (r *DBResourceRepo) GetGroupIDByResourceID(rID string) (string, error) {
	var gID string
	err := r.db.Table("resources r").
		Select("p.g_id").
		Joins("JOIN config_files cf ON cf.cf_id = r.cf_id").
		Joins("JOIN project_list p ON cf.project_id = p.p_id").
		Where("r.r_id = ?", rID).
		Scan(&gID).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return gID, nil
}

func (r *DBResourceRepo) WithTx(tx *gorm.DB) ResourceRepo {
	if tx == nil {
		return r
	}
	return &DBResourceRepo{
		db: tx,
	}
}
