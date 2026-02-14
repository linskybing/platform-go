package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"

	"github.com/linskybing/platform-go/internal/domain/configfile"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type ConfigFileRepo interface {
	Store(ctx context.Context, projectID, authorID, message string, content []byte) (*configfile.ConfigCommit, error)
	GetHead(ctx context.Context, projectID string) (*configfile.ConfigCommit, error)
	GetCommit(ctx context.Context, id string) (*configfile.ConfigCommit, error)
	DeleteCommit(ctx context.Context, id string) error
	GetBlob(ctx context.Context, hash string) (*configfile.ConfigBlob, error)
	GetHistory(ctx context.Context, projectID string) ([]configfile.ConfigCommit, error)
	ListAllCommits(ctx context.Context) ([]configfile.ConfigCommit, error)
	WithTx(tx *gorm.DB) ConfigFileRepo
}

type ConfigFileRepoImpl struct {
	db *gorm.DB
}

func NewConfigFileRepo(db *gorm.DB) ConfigFileRepo {
	return &ConfigFileRepoImpl{db: db}
}

func (r *ConfigFileRepoImpl) Store(ctx context.Context, projectID, authorID, message string, content []byte) (*configfile.ConfigCommit, error) {
	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])
	contentJSON, err := json.Marshal(string(content))
	if err != nil {
		return nil, err
	}
	blob := configfile.ConfigBlob{Hash: hashStr, Content: datatypes.JSON(contentJSON)}
	commitID, err := gonanoid.New()
	if err != nil {
		return nil, err
	}
	commit := configfile.ConfigCommit{ID: commitID, ProjectID: projectID, BlobHash: hashStr, AuthorID: authorID, Message: message}

	db := r.db.WithContext(ctx)
	if err := db.FirstOrCreate(&blob).Error; err != nil {
		return nil, err
	}
	if err := db.Create(&commit).Error; err != nil {
		return nil, err
	}
	return &commit, nil
}

func (r *ConfigFileRepoImpl) GetHead(ctx context.Context, projectID string) (*configfile.ConfigCommit, error) {
	var c configfile.ConfigCommit
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").First(&c).Error
	return &c, err
}

func (r *ConfigFileRepoImpl) GetCommit(ctx context.Context, id string) (*configfile.ConfigCommit, error) {
	var c configfile.ConfigCommit
	err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error
	return &c, err
}

func (r *ConfigFileRepoImpl) DeleteCommit(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&configfile.ConfigCommit{}, "id = ?", id).Error
}

func (r *ConfigFileRepoImpl) GetBlob(ctx context.Context, hash string) (*configfile.ConfigBlob, error) {
	var b configfile.ConfigBlob
	err := r.db.WithContext(ctx).First(&b, "hash = ?", hash).Error
	return &b, err
}

func (r *ConfigFileRepoImpl) GetHistory(ctx context.Context, projectID string) ([]configfile.ConfigCommit, error) {
	var commits []configfile.ConfigCommit
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at DESC").Find(&commits).Error
	return commits, err
}

func (r *ConfigFileRepoImpl) ListAllCommits(ctx context.Context) ([]configfile.ConfigCommit, error) {
	var commits []configfile.ConfigCommit
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&commits).Error
	return commits, err
}

func (r *ConfigFileRepoImpl) WithTx(tx *gorm.DB) ConfigFileRepo {
	return &ConfigFileRepoImpl{db: tx}
}
