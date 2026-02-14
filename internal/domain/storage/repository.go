package storage

import (
	"context"
	"gorm.io/gorm"
)

// StorageRepo defines data access interface for storage resources
type StorageRepo interface {
	// Unified Storage Methods
	CreateStorage(ctx context.Context, storage *Storage) error
	GetStorage(ctx context.Context, id string) (*Storage, error)
	GetStorageByOwnerID(ctx context.Context, ownerID string) (*Storage, error)
	ListStorageByOwnerID(ctx context.Context, ownerID string) ([]Storage, error)
	ListAllStorage(ctx context.Context) ([]Storage, error)
	UpdateStorage(ctx context.Context, storage *Storage) error
	DeleteStorage(ctx context.Context, id string) error

	// Legacy support
	ListGroupStorageByGID(ctx context.Context, gid string) ([]Storage, error)
	GetGroupStorage(ctx context.Context, id string) (*Storage, error)
	DeleteGroupStorage(ctx context.Context, id string) error
	GetUserStorageByUserID(ctx context.Context, userID string) (*Storage, error)
	DeleteUserStorageByUserID(ctx context.Context, userID string) error
	UpdateUserStorage(ctx context.Context, s *Storage) error
	CreateUserStorage(ctx context.Context, s *Storage) error

	WithTx(tx *gorm.DB) StorageRepo
}
