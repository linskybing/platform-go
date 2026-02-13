package repository

import "gorm.io/gorm"

type Repos struct {
	ConfigFile        ConfigFileRepo
	Group             GroupRepo
	Project           ProjectRepo
	Resource          ResourceRepo
	UserGroup         UserGroupRepo
	User              UserRepo
	Audit             AuditRepo
	Form              FormRepo
	Image             ImageRepo
	StoragePermission StoragePermissionRepo
	Job               JobRepo
	GPUUsage          GPUUsageRepo
	Storage           *DBStorageRepo

	db *gorm.DB
}

func NewRepositories(db *gorm.DB) *Repos {
	return &Repos{
		ConfigFile:        NewConfigFileRepo(db),
		Group:             NewGroupRepo(db),
		Project:           NewProjectRepo(db),
		Resource:          NewResourceRepo(db),
		UserGroup:         NewUserGroupRepo(db),
		User:              NewUserRepo(db),
		Audit:             NewAuditRepo(db),
		Form:              NewFormRepo(db),
		Image:             NewImageRepo(db),
		StoragePermission: NewStoragePermissionRepo(db),
		Storage:           NewStorageRepo(db),
		Job:               NewJobRepo(db),
		GPUUsage:          NewGPUUsageRepo(db),
		db:                db,
	}
}

func (r *Repos) Begin() *gorm.DB {
	return r.db.Begin()
}

func (r *Repos) WithTx(tx *gorm.DB) *Repos {
	txRepos := NewRepositories(tx)
	txRepos.db = tx
	return txRepos
}

// DB returns the underlying GORM database handle.
func (r *Repos) DB() *gorm.DB {
	return r.db
}

func (r *Repos) ExecTx(fn func(*Repos) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		txRepos := r.WithTx(tx)
		return fn(txRepos)
	})
}
