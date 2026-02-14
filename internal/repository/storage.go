package repository

import (
	"context"
	"github.com/linskybing/platform-go/internal/domain/storage"
	"gorm.io/gorm"
)

// StorageRepoImpl implements storage.StorageRepo
type StorageRepoImpl struct {
	db *gorm.DB
}

func NewStorageRepo(db *gorm.DB) storage.StorageRepo {
	return &StorageRepoImpl{db: db}
}

func (r *StorageRepoImpl) setAliases(s *storage.Storage) {
	if s == nil {
		return
	}
	s.GroupID = s.OwnerID
	s.UserID = s.OwnerID
}

func (r *StorageRepoImpl) CreateStorage(ctx context.Context, s *storage.Storage) error {
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *StorageRepoImpl) GetStorage(ctx context.Context, id string) (*storage.Storage, error) {
	var s storage.Storage
	err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error
	r.setAliases(&s)
	return &s, err
}

func (r *StorageRepoImpl) GetStorageByOwnerID(ctx context.Context, ownerID string) (*storage.Storage, error) {
	var s storage.Storage
	err := r.db.WithContext(ctx).First(&s, "owner_id = ?", ownerID).Error
	r.setAliases(&s)
	return &s, err
}

func (r *StorageRepoImpl) ListStorageByOwnerID(ctx context.Context, ownerID string) ([]storage.Storage, error) {
	var ss []storage.Storage
	err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&ss).Error
	for i := range ss {
		r.setAliases(&ss[i])
	}
	return ss, err
}

func (r *StorageRepoImpl) ListAllStorage(ctx context.Context) ([]storage.Storage, error) {
	var ss []storage.Storage
	err := r.db.WithContext(ctx).Find(&ss).Error
	for i := range ss {
		r.setAliases(&ss[i])
	}
	return ss, err
}

func (r *StorageRepoImpl) UpdateStorage(ctx context.Context, s *storage.Storage) error {
	return r.db.WithContext(ctx).Save(s).Error
}

func (r *StorageRepoImpl) DeleteStorage(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&storage.Storage{}, "id = ?", id).Error
}

func (r *StorageRepoImpl) ListGroupStorageByGID(ctx context.Context, gid string) ([]storage.Storage, error) {
	return r.ListStorageByOwnerID(ctx, gid)
}

func (r *StorageRepoImpl) GetGroupStorage(ctx context.Context, id string) (*storage.Storage, error) {
	return r.GetStorage(ctx, id)
}

func (r *StorageRepoImpl) DeleteGroupStorage(ctx context.Context, id string) error {
	return r.DeleteStorage(ctx, id)
}

func (r *StorageRepoImpl) GetUserStorageByUserID(ctx context.Context, userID string) (*storage.Storage, error) {
	return r.GetStorageByOwnerID(ctx, userID)
}

func (r *StorageRepoImpl) DeleteUserStorageByUserID(ctx context.Context, userID string) error {
	s, err := r.GetStorageByOwnerID(ctx, userID)
	if err != nil {
		return err
	}
	return r.DeleteStorage(ctx, s.ID)
}

func (r *StorageRepoImpl) UpdateUserStorage(ctx context.Context, s *storage.Storage) error {
	return r.UpdateStorage(ctx, s)
}

func (r *StorageRepoImpl) CreateUserStorage(ctx context.Context, s *storage.Storage) error {
	return r.CreateStorage(ctx, s)
}

func (r *StorageRepoImpl) WithTx(tx *gorm.DB) storage.StorageRepo {
	if tx == nil {
		return r
	}
	return &StorageRepoImpl{db: tx}
}
