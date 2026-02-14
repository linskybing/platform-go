package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/linskybing/platform-go/internal/api/middleware"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// --------------------- RegisterUser ---------------------
func TestRegisterUser_Success(t *testing.T) {
	svc, stubUser := setupUserService(t)

	input := user.CreateUserInput{
		Username: "alice",
		Password: "password123",
		Email:    ptrString("alice@test.com"),
		FullName: ptrString("Alice"),
	}

	stubUser.getByUsername = func(ctx context.Context, username string) (*user.User, error) {
		return nil, gorm.ErrRecordNotFound
	}
	stubUser.saveUser = func(ctx context.Context, u *user.User) error {
		return nil
	}

	err := svc.RegisterUser(input)
	assert.NoError(t, err)
}

func TestRegisterUser_UsernameTaken(t *testing.T) {
	svc, stubUser := setupUserService(t)

	stubUser.getByUsername = func(ctx context.Context, username string) (*user.User, error) {
		return &user.User{UID: "1"}, nil
	}

	input := user.CreateUserInput{Username: "admin", Password: "password123"}
	err := svc.RegisterUser(input)
	assert.Equal(t, ErrUsernameTaken, err)
}

// --------------------- LoginUser ---------------------
func TestLoginUser_Success(t *testing.T) {
	svc, stubUser := setupUserService(t)

	password := "password123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	usr := user.User{ID: "1", UID: "1", Username: "bob", PasswordHash: string(hashed)}

	stubUser.getByUsername = func(ctx context.Context, username string) (*user.User, error) {
		return &usr, nil
	}

	oldGen := middleware.GenerateToken
	middleware.GenerateToken = func(userID string, username string, expireDuration time.Duration, repos repository.UserGroupRepo) (string, bool, error) {
		return "token123", true, nil
	}
	defer func() { middleware.GenerateToken = oldGen }()

	u, token, isAdmin, err := svc.LoginUser("bob", "password123")
	assert.NoError(t, err)
	assert.Equal(t, "bob", u.Username)
	assert.Equal(t, "token123", token)
	assert.True(t, isAdmin)
}

func TestLoginUser_InvalidPassword(t *testing.T) {
	svc, stubUser := setupUserService(t)

	password := "123456"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	usr := user.User{ID: "1", UID: "1", Username: "bob", PasswordHash: string(hashed)}

	stubUser.getByUsername = func(ctx context.Context, username string) (*user.User, error) {
		return &usr, nil
	}

	u, token, isAdmin, err := svc.LoginUser("bob", "wrong")
	assert.Error(t, err)
	assert.Equal(t, user.User{}, u)
	assert.Empty(t, token)
	assert.False(t, isAdmin)
}

func TestLoginUser_UserNotFound(t *testing.T) {
	svc, stubUser := setupUserService(t)
	stubUser.getByUsername = func(ctx context.Context, username string) (*user.User, error) {
		return nil, errors.New("not found")
	}

	u, token, isAdmin, err := svc.LoginUser("notexist", "123")
	assert.Error(t, err)
	assert.Equal(t, user.User{}, u)
	assert.Empty(t, token)
	assert.False(t, isAdmin)
}

// --------------------- UpdateUser ---------------------
func TestUpdateUser_SuccessChangePassword(t *testing.T) {
	svc, stubUser := setupUserService(t)

	oldPass := "oldpassword123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(oldPass), bcrypt.DefaultCost)
	existing := user.User{ID: "1", UID: "1", PasswordHash: string(hashed)}
	oldHash := existing.PasswordHash

	stubUser.get = func(ctx context.Context, id string) (*user.User, error) {
		return &existing, nil
	}
	stubUser.saveUser = func(ctx context.Context, u *user.User) error {
		return nil
	}

	newPass := "newpassword456"
	input := user.UpdateUserInput{
		OldPassword: &oldPass,
		Password:    &newPass,
	}

	updated, err := svc.UpdateUser("1", input)
	assert.NoError(t, err)
	assert.NotEqual(t, oldHash, updated.PasswordHash)
	assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(updated.PasswordHash), []byte(newPass)))
}

func TestUpdateUser_WrongOldPassword(t *testing.T) {
	svc, stubUser := setupUserService(t)

	oldPass := "oldpassword123"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(oldPass), bcrypt.DefaultCost)
	existing := user.User{ID: "1", UID: "1", PasswordHash: string(hashed)}

	stubUser.get = func(ctx context.Context, id string) (*user.User, error) {
		return &existing, nil
	}

	wrongPass := "wrongpass123"
	input := user.UpdateUserInput{OldPassword: &wrongPass, Password: &wrongPass}

	updated, err := svc.UpdateUser("1", input)
	assert.ErrorIs(t, err, ErrIncorrectPassword)
	assert.Equal(t, user.User{}, updated)
}

func TestUpdateUser_UserNotFound(t *testing.T) {
	svc, stubUser := setupUserService(t)
	stubUser.get = func(ctx context.Context, id string) (*user.User, error) {
		return nil, errors.New("not found")
	}

	input := user.UpdateUserInput{FullName: ptrString("NewName")}
	updated, err := svc.UpdateUser("1", input)
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Equal(t, user.User{}, updated)
}

// --------------------- RemoveUser ---------------------
func TestRemoveUser_Success(t *testing.T) {
	svc, stubUser := setupUserService(t)
	stubUser.get = func(ctx context.Context, id string) (*user.User, error) {
		return &user.User{Username: "testuser"}, nil
	}
	stubUser.deleteUser = func(ctx context.Context, id string) error {
		return nil
	}

	err := svc.RemoveUser("1")
	assert.NoError(t, err)
}

func TestRemoveUser_Fail(t *testing.T) {
	svc, stubUser := setupUserService(t)
	stubUser.get = func(ctx context.Context, id string) (*user.User, error) {
		return &user.User{Username: "testuser"}, nil
	}
	stubUser.deleteUser = func(ctx context.Context, id string) error {
		return errors.New("delete fail")
	}

	err := svc.RemoveUser("1")
	assert.EqualError(t, err, "delete fail")
}
