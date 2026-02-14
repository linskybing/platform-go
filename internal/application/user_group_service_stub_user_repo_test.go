package application_test

import (
	"context"

	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubUserRepoLite struct {
	getUsernameByID func(ctx context.Context, id string) (string, error)
}

func (s *stubUserRepoLite) Create(ctx context.Context, u *user.User) error { return nil }
func (s *stubUserRepoLite) Get(ctx context.Context, id string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepoLite) GetUserRawByID(ctx context.Context, id string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepoLite) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepoLite) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepoLite) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	return nil, gorm.ErrRecordNotFound
}
func (s *stubUserRepoLite) GetUsernameByID(ctx context.Context, id string) (string, error) {
	if s.getUsernameByID != nil {
		return s.getUsernameByID(ctx, id)
	}
	return "", gorm.ErrRecordNotFound
}
func (s *stubUserRepoLite) List(ctx context.Context) ([]user.User, error)        { return nil, nil }
func (s *stubUserRepoLite) GetAllUsers(ctx context.Context) ([]user.User, error) { return nil, nil }
func (s *stubUserRepoLite) ListUsersPaging(ctx context.Context, offset, limit int) ([]user.User, int64, error) {
	return nil, 0, nil
}
func (s *stubUserRepoLite) ListUsersByProjectID(ctx context.Context, pid string) ([]user.User, error) {
	return nil, nil
}
func (s *stubUserRepoLite) SaveUser(ctx context.Context, u *user.User) error { return nil }
func (s *stubUserRepoLite) Delete(ctx context.Context, id string) error      { return nil }
func (s *stubUserRepoLite) WithTx(tx *gorm.DB) repository.UserRepo           { return s }
