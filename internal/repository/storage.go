package repository

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/storage"
	"gorm.io/gorm"
)

type DBStorageRepo struct {
	db *gorm.DB
}

func NewStorageRepo(db *gorm.DB) *DBStorageRepo {
	return &DBStorageRepo{db: db}
}

func (r *DBStorageRepo) WithTx(tx *gorm.DB) storage.StorageRepo {
	if tx == nil {
		return r
	}
	return &DBStorageRepo{
		db: tx,
	}
}

func (r *DBStorageRepo) CreateUserStorage(ctx context.Context, pvc *storage.UserStorage) error {
	return r.db.WithContext(ctx).Create(pvc).Error
}

func (r *DBStorageRepo) CreateGroupStorage(ctx context.Context, pvc *storage.GroupStorage) error {
	return r.db.WithContext(ctx).Create(pvc).Error
}

func (r *DBStorageRepo) GetUserStorage(ctx context.Context, id string) (*storage.UserStorage, error) {
	var pvc storage.UserStorage
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&pvc).Error; err != nil {
		return nil, err
	}
	return &pvc, nil
}

func (r *DBStorageRepo) GetGroupStorage(ctx context.Context, id string) (*storage.GroupStorage, error) {
	var pvc storage.GroupStorage
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&pvc).Error; err != nil {
		return nil, err
	}
	return &pvc, nil
}

func (r *DBStorageRepo) GetUserStorageByUserID(ctx context.Context, userID string) (*storage.UserStorage, error) {
	var pvc storage.UserStorage
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&pvc).Error; err != nil {
		return nil, err
	}
	return &pvc, nil
}

func (r *DBStorageRepo) ListUserStorage(ctx context.Context) ([]storage.UserStorage, error) {
	var pvcs []storage.UserStorage
	if err := r.db.WithContext(ctx).Find(&pvcs).Error; err != nil {
		return nil, err
	}
	return pvcs, nil
}

func (r *DBStorageRepo) ListGroupStorage(ctx context.Context) ([]storage.GroupStorage, error) {
	var pvcs []storage.GroupStorage
	if err := r.db.WithContext(ctx).Find(&pvcs).Error; err != nil {
		return nil, err
	}
	return pvcs, nil
}

func (r *DBStorageRepo) ListGroupStorageByGID(ctx context.Context, gid string) ([]storage.GroupStorage, error) {
	var pvcs []storage.GroupStorage
	if err := r.db.WithContext(ctx).Where("group_id = ?", gid).Find(&pvcs).Error; err != nil {
		return nil, err
	}
	return pvcs, nil
}

func (r *DBStorageRepo) UpdateUserStorage(ctx context.Context, pvc *storage.UserStorage) error {
	return r.db.WithContext(ctx).Model(pvc).Save(pvc).Error
}

func (r *DBStorageRepo) UpdateGroupStorage(ctx context.Context, pvc *storage.GroupStorage) error {
	return r.db.WithContext(ctx).Model(pvc).Save(pvc).Error
}

func (r *DBStorageRepo) DeleteUserStorage(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&storage.UserStorage{}).Error
}

func (r *DBStorageRepo) DeleteUserStorageByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&storage.UserStorage{}).Error
}

func (r *DBStorageRepo) DeleteGroupStorage(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&storage.GroupStorage{}).Error
}
