package repository

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/common"
	"github.com/linskybing/platform-go/internal/domain/user"
	"gorm.io/gorm"
)

type UserRepo interface {
	Create(ctx context.Context, u *user.User) error
	Get(ctx context.Context, id string) (*user.User, error)
	GetUserRawByID(ctx context.Context, id string) (*user.User, error)
	GetUserByID(ctx context.Context, id string) (*user.User, error)
	GetByUsername(ctx context.Context, username string) (*user.User, error)
	GetUserByUsername(ctx context.Context, username string) (*user.User, error)
	GetUsernameByID(ctx context.Context, id string) (string, error)
	List(ctx context.Context) ([]user.User, error)
	GetAllUsers(ctx context.Context) ([]user.User, error)
	ListUsersPaging(ctx context.Context, offset, limit int) ([]user.User, int64, error)
	ListUsersByProjectID(ctx context.Context, pid string) ([]user.User, error)
	SaveUser(ctx context.Context, u *user.User) error
	Delete(ctx context.Context, id string) error
	WithTx(tx *gorm.DB) UserRepo
}

type UserRepoImpl struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepo {
	return &UserRepoImpl{db: db}
}

func (r *UserRepoImpl) setAliases(u *user.User) {
	if u == nil {
		return
	}
	u.UID = u.ID
}

func (r *UserRepoImpl) Create(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		owner := common.ResourceOwner{OwnerType: "USER"}
		if err := tx.Create(&owner).Error; err != nil {
			return err
		}
		u.ID = owner.ID
		return tx.Create(u).Error
	})
}

func (r *UserRepoImpl) Get(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).First(&u, "id = ?", id).Error
	r.setAliases(&u)
	return &u, err
}

func (r *UserRepoImpl) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	return r.Get(ctx, id)
}

func (r *UserRepoImpl) GetUserRawByID(ctx context.Context, id string) (*user.User, error) {
	return r.Get(ctx, id)
}

func (r *UserRepoImpl) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var u user.User
	err := r.db.WithContext(ctx).First(&u, "username = ?", username).Error
	r.setAliases(&u)
	return &u, err
}

func (r *UserRepoImpl) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	return r.GetByUsername(ctx, username)
}

func (r *UserRepoImpl) GetUsernameByID(ctx context.Context, id string) (string, error) {
	var u user.User
	err := r.db.WithContext(ctx).Select("username").First(&u, "id = ?", id).Error
	return u.Username, err
}

func (r *UserRepoImpl) List(ctx context.Context) ([]user.User, error) {
	var users []user.User
	err := r.db.WithContext(ctx).Find(&users).Error
	for i := range users {
		r.setAliases(&users[i])
	}
	return users, err
}

func (r *UserRepoImpl) GetAllUsers(ctx context.Context) ([]user.User, error) {
	return r.List(ctx)
}

func (r *UserRepoImpl) ListUsersPaging(ctx context.Context, offset, limit int) ([]user.User, int64, error) {
	var users []user.User
	var count int64
	r.db.WithContext(ctx).Model(&user.User{}).Count(&count)
	err := r.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&users).Error
	for i := range users {
		r.setAliases(&users[i])
	}
	return users, count, err
}

func (r *UserRepoImpl) ListUsersByProjectID(ctx context.Context, pid string) ([]user.User, error) {
	var users []user.User
	err := r.db.WithContext(ctx).Joins("JOIN user_group ug ON ug.user_id = users.id").
		Joins("JOIN projects p ON p.owner_id = ug.group_id").
		Where("p.p_id = ?", pid).Find(&users).Error
	for i := range users {
		r.setAliases(&users[i])
	}
	return users, err
}

func (r *UserRepoImpl) SaveUser(ctx context.Context, u *user.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *UserRepoImpl) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&common.ResourceOwner{}, "id = ?", id).Error
}

func (r *UserRepoImpl) WithTx(tx *gorm.DB) UserRepo {
	if tx == nil {
		return r
	}
	return &UserRepoImpl{db: tx}
}
