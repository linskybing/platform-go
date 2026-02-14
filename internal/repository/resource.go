package repository

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/resource"
	"gorm.io/gorm"
)

type ResourceRepo interface {
	CreateResource(ctx context.Context, resource *resource.Resource) error
	GetResourceByID(ctx context.Context, rid string) (*resource.Resource, error)
	UpdateResource(ctx context.Context, resource *resource.Resource) error
	DeleteResource(ctx context.Context, rid string) error
	ListResourcesByProjectID(ctx context.Context, pid string) ([]resource.Resource, error)
	ListResourcesByCommitID(ctx context.Context, commitID string) ([]resource.Resource, error)
	GetResourceByCommitIDAndName(ctx context.Context, commitID string, name string) (*resource.Resource, error)
	GetGroupIDByResourceID(ctx context.Context, rID string) (string, error)
	WithTx(tx *gorm.DB) ResourceRepo
}

type ResourceRepoImpl struct {
	db *gorm.DB
}

func NewResourceRepo(db *gorm.DB) ResourceRepo {
	return &ResourceRepoImpl{db: db}
}

func (r *ResourceRepoImpl) CreateResource(ctx context.Context, res *resource.Resource) error {
	return r.db.WithContext(ctx).Create(res).Error
}

func (r *ResourceRepoImpl) GetResourceByID(ctx context.Context, rid string) (*resource.Resource, error) {
	var res resource.Resource
	err := r.db.WithContext(ctx).First(&res, "id = ?", rid).Error
	return &res, err
}

func (r *ResourceRepoImpl) UpdateResource(ctx context.Context, res *resource.Resource) error {
	return r.db.WithContext(ctx).Save(res).Error
}

func (r *ResourceRepoImpl) DeleteResource(ctx context.Context, rid string) error {
	return r.db.WithContext(ctx).Delete(&resource.Resource{}, "id = ?", rid).Error
}

func (r *ResourceRepoImpl) ListResourcesByProjectID(ctx context.Context, pid string) ([]resource.Resource, error) {
	var res []resource.Resource
	err := r.db.WithContext(ctx).Joins("JOIN config_commits cc ON cc.id = resources.config_commit_id").
		Where("cc.project_id = ?", pid).Find(&res).Error
	return res, err
}

func (r *ResourceRepoImpl) ListResourcesByCommitID(ctx context.Context, commitID string) ([]resource.Resource, error) {
	var res []resource.Resource
	err := r.db.WithContext(ctx).Where("config_commit_id = ?", commitID).Find(&res).Error
	return res, err
}

func (r *ResourceRepoImpl) GetResourceByCommitIDAndName(ctx context.Context, commitID string, name string) (*resource.Resource, error) {
	var res resource.Resource
	err := r.db.WithContext(ctx).Where("config_commit_id = ? AND name = ?", commitID, name).First(&res).Error
	return &res, err
}

func (r *ResourceRepoImpl) GetGroupIDByResourceID(ctx context.Context, rID string) (string, error) {
	return "", nil
}

func (r *ResourceRepoImpl) WithTx(tx *gorm.DB) ResourceRepo {
	if tx == nil {
		return r
	}
	return &ResourceRepoImpl{db: tx}
}
