package user

import (
	"context"
	"errors"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/stretchr/testify/assert"
)

func TestListUsers_Success(t *testing.T) {
	svc, stubUser := setupUserService(t)

	users := []user.User{
		{UID: "1", Username: "alice"},
		{UID: "2", Username: "bob"},
	}
	stubUser.list = func(ctx context.Context) ([]user.User, error) {
		return users, nil
	}

	result, err := svc.ListUsers()
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestListUserByPaging_Success(t *testing.T) {
	svc, stubUser := setupUserService(t)

	users := []user.User{
		{UID: "1", Username: "alice"},
	}
	stubUser.listUsersPaging = func(ctx context.Context, offset, limit int) ([]user.User, int64, error) {
		return users, int64(len(users)), nil
	}

	result, err := svc.ListUserByPaging(1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
}

func TestFindUserByID_Success(t *testing.T) {
	svc, stubUser := setupUserService(t)

	usr := user.User{UID: "1", Username: "alice"}
	stubUser.get = func(ctx context.Context, id string) (*user.User, error) {
		return &usr, nil
	}

	result, err := svc.FindUserByID("1")
	assert.NoError(t, err)
	assert.Equal(t, "alice", result.Username)
}

func TestFindUserByID_NotFound(t *testing.T) {
	svc, stubUser := setupUserService(t)

	stubUser.get = func(ctx context.Context, id string) (*user.User, error) {
		return nil, errors.New("not found")
	}

	_, err := svc.FindUserByID("999")
	assert.Error(t, err)
}

func TestUpdateUser_SuccessNoPasswordChange(t *testing.T) {
	svc, stubUser := setupUserService(t)

	existing := user.User{UID: "1", Username: "alice", Email: "old@test.com"}
	stubUser.get = func(ctx context.Context, id string) (*user.User, error) {
		return &existing, nil
	}

	stubUser.saveUser = func(ctx context.Context, u *user.User) error {
		assert.Equal(t, "new@test.com", u.Email)
		return nil
	}

	input := user.UpdateUserInput{Email: ptrString("new@test.com")}
	updated, err := svc.UpdateUser("1", input)
	assert.NoError(t, err)
	assert.Equal(t, "new@test.com", updated.Email)
}

func TestUpdateUser_FailSave(t *testing.T) {
	svc, stubUser := setupUserService(t)

	existing := user.User{UID: "1"}
	stubUser.get = func(ctx context.Context, id string) (*user.User, error) {
		return &existing, nil
	}
	stubUser.saveUser = func(ctx context.Context, u *user.User) error {
		return errors.New("db error")
	}

	input := user.UpdateUserInput{Email: ptrString("new@test.com")}
	_, err := svc.UpdateUser("1", input)
	assert.Error(t, err)
}
