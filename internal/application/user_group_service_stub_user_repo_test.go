package application_test

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubUserRepoLiteForUserGroup struct {
	getUsernameByID func(ctx context.Context, id string) (string, error)
}

func (s *stubUserRepoLiteForUserGroup) Create(ctx context.Context, u *user.User) error { return nil }
func (s *stubUserRepoLiteForUserGroup) Get(ctx context.Context, id string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepoLiteForUserGroup) GetUserRawByID(ctx context.Context, id string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepoLiteForUserGroup) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepoLiteForUserGroup) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepoLiteForUserGroup) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepoLiteForUserGroup) GetUsernameByID(ctx context.Context, id string) (string, error) {
	if s.getUsernameByID != nil {
		return s.getUsernameByID(ctx, id)
	}
	return "", gorm.ErrRecordNotFound
}
func (s *stubUserRepoLiteForUserGroup) List(ctx context.Context) ([]user.User, error) {
	return nil, nil
}
func (s *stubUserRepoLiteForUserGroup) GetAllUsers(ctx context.Context) ([]user.User, error) {
	return nil, nil
}
func (s *stubUserRepoLiteForUserGroup) ListUsersPaging(ctx context.Context, offset, limit int) ([]user.User, int64, error) {
	return nil, 0, nil
}
func (s *stubUserRepoLiteForUserGroup) ListUsersByProjectID(ctx context.Context, pid string) ([]user.User, error) {
	return nil, nil
}
func (s *stubUserRepoLiteForUserGroup) SaveUser(ctx context.Context, u *user.User) error { return nil }
func (s *stubUserRepoLiteForUserGroup) Delete(ctx context.Context, id string) error      { return nil }
func (s *stubUserRepoLiteForUserGroup) WithTx(tx *gorm.DB) repository.UserRepo           { return s }
