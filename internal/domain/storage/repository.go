package storage

import (
	"context"

	"gorm.io/gorm"
)

// Repository defines data access interface for storage resources
type StorageRepo interface {
	// Storage
	CreateUserStorage(ctx context.Context, pvc *UserStorage) error
	CreateGroupStorage(ctx context.Context, pvc *GroupStorage) error

	GetUserStorage(ctx context.Context, id string) (*UserStorage, error)
	GetGroupStorage(ctx context.Context, id string) (*GroupStorage, error)
	GetUserStorageByUserID(ctx context.Context, userID string) (*UserStorage, error)

	ListUserStorage(ctx context.Context) ([]UserStorage, error)
	ListGroupStorage(ctx context.Context) ([]GroupStorage, error)
	ListGroupStorageByGID(ctx context.Context, gid string) ([]GroupStorage, error)

	UpdateUserStorage(ctx context.Context, pvc *UserStorage) error
	UpdateGroupStorage(ctx context.Context, pvc *GroupStorage) error

	DeleteUserStorage(ctx context.Context, id string) error
	DeleteUserStorageByUserID(ctx context.Context, userID string) error
	DeleteGroupStorage(ctx context.Context, id string) error

	WithTx(tx *gorm.DB) StorageRepo
}
