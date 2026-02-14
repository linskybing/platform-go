package repository

import (
	"github.com/linskybing/platform-go/internal/domain/storage"
	"gorm.io/gorm"
)

type Repos struct {
	ConfigFile        ConfigFileRepo
	Group             GroupRepo
	Project           ProjectRepo
	User              UserRepo
	Audit             AuditRepo
	Form              FormRepo
	Image             ImageRepo
	StoragePermission StoragePermissionRepo
	Job               JobRepo
	GPUUsage          GPUUsageRepo
	Storage           storage.StorageRepo
	UserGroup         UserGroupRepo
	Resource          ResourceRepo
	db                *gorm.DB
}

func NewRepositories(db *gorm.DB) *Repos {
	return &Repos{
		ConfigFile:        NewConfigFileRepo(db),
		Group:             NewGroupRepo(db),
		Project:           NewProjectRepo(db),
		User:              NewUserRepo(db),
		Audit:             NewAuditRepo(db),
		Form:              NewFormRepo(db),
		Image:             NewImageRepo(db),
		StoragePermission: NewStoragePermissionRepo(db),
		Storage:           NewStorageRepo(db),
		Job:               NewJobRepo(db),
		GPUUsage:          NewGPUUsageRepo(db),
		UserGroup:         NewUserGroupRepo(db),
		Resource:          NewResourceRepo(db),
		db:                db,
	}
}

func (r *Repos) Begin() *gorm.DB { return r.db.Begin() }

func (r *Repos) WithTx(tx *gorm.DB) *Repos {
	txRepos := NewRepositories(tx)
	return txRepos
}

func (r *Repos) DB() *gorm.DB { return r.db }

func (r *Repos) ExecTx(fn func(*Repos) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(r.WithTx(tx))
	})
}
