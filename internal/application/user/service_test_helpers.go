package user

import (
	"context"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/internal/repository"
	"gorm.io/gorm"
)

type stubUserRepo struct {
	getByUsername   func(ctx context.Context, username string) (*user.User, error)
	get             func(ctx context.Context, id string) (*user.User, error)
	list            func(ctx context.Context) ([]user.User, error)
	listUsersPaging func(ctx context.Context, offset, limit int) ([]user.User, int64, error)
	saveUser        func(ctx context.Context, u *user.User) error
	deleteUser      func(ctx context.Context, id string) error
}

func (s *stubUserRepo) Create(ctx context.Context, u *user.User) error {
	return s.SaveUser(ctx, u)
}

func (s *stubUserRepo) Get(ctx context.Context, id string) (*user.User, error) {
	if s.get != nil {
		return s.get(ctx, id)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubUserRepo) GetUserRawByID(ctx context.Context, id string) (*user.User, error) {
	return s.Get(ctx, id)
}

func (s *stubUserRepo) GetUserByID(ctx context.Context, id string) (*user.User, error) {
	return s.Get(ctx, id)
}

func (s *stubUserRepo) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	if s.getByUsername != nil {
		return s.getByUsername(ctx, username)
	}
	return nil, gorm.ErrRecordNotFound
}

func (s *stubUserRepo) GetUserByUsername(ctx context.Context, username string) (*user.User, error) {
	return s.GetByUsername(ctx, username)
}

func (s *stubUserRepo) GetUsernameByID(ctx context.Context, id string) (string, error) {
	usr, err := s.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return usr.Username, nil
}

func (s *stubUserRepo) List(ctx context.Context) ([]user.User, error) {
	if s.list != nil {
		return s.list(ctx)
	}
	return nil, nil
}

func (s *stubUserRepo) GetAllUsers(ctx context.Context) ([]user.User, error) {
	return s.List(ctx)
}

func (s *stubUserRepo) ListUsersPaging(ctx context.Context, offset, limit int) ([]user.User, int64, error) {
	if s.listUsersPaging != nil {
		return s.listUsersPaging(ctx, offset, limit)
	}
	return nil, 0, nil
}

func (s *stubUserRepo) ListUsersByProjectID(ctx context.Context, pid string) ([]user.User, error) {
	return nil, nil
}

func (s *stubUserRepo) SaveUser(ctx context.Context, u *user.User) error {
	if s.saveUser != nil {
		return s.saveUser(ctx, u)
	}
	return nil
}

func (s *stubUserRepo) Delete(ctx context.Context, id string) error {
	if s.deleteUser != nil {
		return s.deleteUser(ctx, id)
	}
	return nil
}

func (s *stubUserRepo) WithTx(tx *gorm.DB) repository.UserRepo {
	return s
}

func setupUserService(t *testing.T) (*UserService, *stubUserRepo) {
	t.Helper()
	stub := &stubUserRepo{}
	repos := &repository.Repos{User: stub}
	return NewUserService(repos), stub
}

func ptrString(s string) *string { return &s }
