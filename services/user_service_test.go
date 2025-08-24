package services

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/linskybing/platform-go/dto"
	"github.com/linskybing/platform-go/middleware"
	"github.com/linskybing/platform-go/models"
	"github.com/linskybing/platform-go/repositories"
	"github.com/linskybing/platform-go/repositories/mock_repositories"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// --------------------- Setup ---------------------
func setupUserServiceMocks(t *testing.T) (*UserService, *mock_repositories.MockUserRepo) {
	ctrl := gomock.NewController(t)
	t.Cleanup(func() { ctrl.Finish() })

	mockUser := mock_repositories.NewMockUserRepo(ctrl)
	repos := &repositories.Repos{
		User: mockUser,
	}
	svc := NewUserService(repos)
	return svc, mockUser
}

// --------------------- RegisterUser ---------------------
func TestRegisterUser_Success(t *testing.T) {
	svc, mockUser := setupUserServiceMocks(t)

	input := dto.CreateUserInput{
		Username: "alice",
		Password: "123456",
		Email:    ptrString("alice@test.com"),
		FullName: ptrString("Alice"),
	}

	mockUser.EXPECT().GetUserByUsername("alice").Return(models.User{}, gorm.ErrRecordNotFound)
	mockUser.EXPECT().SaveUser(gomock.Any()).Return(nil)

	err := svc.RegisterUser(input)
	assert.NoError(t, err)
}

func TestRegisterUser_UsernameTaken(t *testing.T) {
	svc, mockUser := setupUserServiceMocks(t)

	mockUser.EXPECT().GetUserByUsername("admin").Return(models.User{UID: 1}, nil)

	input := dto.CreateUserInput{Username: "admin", Password: "123456"}
	err := svc.RegisterUser(input)
	assert.Equal(t, ErrUsernameTaken, err)
}

// --------------------- LoginUser ---------------------
func TestLoginUser_Success(t *testing.T) {
	svc, mockUser := setupUserServiceMocks(t)

	password := "123456"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{UID: 1, Username: "bob", Password: string(hashed)}

	mockUser.EXPECT().GetUserByUsername("bob").Return(user, nil)

	oldGen := middleware.GenerateToken
	middleware.GenerateToken = func(uid uint, username string, exp time.Duration, view repositories.ViewRepo) (string, bool, error) {
		return "token123", true, nil
	}
	defer func() { middleware.GenerateToken = oldGen }()

	u, token, isAdmin, err := svc.LoginUser("bob", "123456")
	assert.NoError(t, err)
	assert.Equal(t, "bob", u.Username)
	assert.Equal(t, "token123", token)
	assert.True(t, isAdmin)
}

func TestLoginUser_InvalidPassword(t *testing.T) {
	svc, mockUser := setupUserServiceMocks(t)

	password := "123456"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{UID: 1, Username: "bob", Password: string(hashed)}

	mockUser.EXPECT().GetUserByUsername("bob").Return(user, nil)

	u, token, isAdmin, err := svc.LoginUser("bob", "wrong")
	assert.Error(t, err)
	assert.Equal(t, models.User{}, u)
	assert.Empty(t, token)
	assert.False(t, isAdmin)
}

func TestLoginUser_UserNotFound(t *testing.T) {
	svc, mockUser := setupUserServiceMocks(t)
	mockUser.EXPECT().GetUserByUsername("notexist").Return(models.User{}, errors.New("not found"))

	u, token, isAdmin, err := svc.LoginUser("notexist", "123")
	assert.Error(t, err)
	assert.Equal(t, models.User{}, u)
	assert.Empty(t, token)
	assert.False(t, isAdmin)
}

// --------------------- UpdateUser ---------------------
func TestUpdateUser_SuccessChangePassword(t *testing.T) {
	svc, mockUser := setupUserServiceMocks(t)

	oldPass := "oldpass"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(oldPass), bcrypt.DefaultCost)
	existing := models.User{UID: 1, Password: string(hashed)}

	mockUser.EXPECT().GetUserRawByID(uint(1)).Return(existing, nil)
	mockUser.EXPECT().SaveUser(gomock.Any()).Return(nil)

	newPass := "newpass"
	input := dto.UpdateUserInput{
		OldPassword: &oldPass,
		Password:    &newPass,
	}

	updated, err := svc.UpdateUser(1, input)
	assert.NoError(t, err)
	assert.NotEqual(t, existing.Password, updated.Password)
}

func TestUpdateUser_WrongOldPassword(t *testing.T) {
	svc, mockUser := setupUserServiceMocks(t)

	oldPass := "oldpass"
	hashed, _ := bcrypt.GenerateFromPassword([]byte(oldPass), bcrypt.DefaultCost)
	existing := models.User{UID: 1, Password: string(hashed)}

	mockUser.EXPECT().GetUserRawByID(uint(1)).Return(existing, nil)

	wrongPass := "wrong"
	input := dto.UpdateUserInput{OldPassword: &wrongPass, Password: &wrongPass}

	updated, err := svc.UpdateUser(1, input)
	assert.ErrorIs(t, err, ErrIncorrectPassword)
	assert.Equal(t, models.User{}, updated)
}

func TestUpdateUser_UserNotFound(t *testing.T) {
	svc, mockUser := setupUserServiceMocks(t)
	mockUser.EXPECT().GetUserRawByID(uint(1)).Return(models.User{}, errors.New("not found"))

	input := dto.UpdateUserInput{FullName: ptrString("NewName")}
	updated, err := svc.UpdateUser(1, input)
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Equal(t, models.User{}, updated)
}

// --------------------- RemoveUser ---------------------
func TestRemoveUser_Success(t *testing.T) {
	svc, mockUser := setupUserServiceMocks(t)
	mockUser.EXPECT().DeleteUser(uint(1)).Return(nil)

	err := svc.RemoveUser(1)
	assert.NoError(t, err)
}

func TestRemoveUser_Fail(t *testing.T) {
	svc, mockUser := setupUserServiceMocks(t)
	mockUser.EXPECT().DeleteUser(uint(1)).Return(errors.New("delete fail"))

	err := svc.RemoveUser(1)
	assert.EqualError(t, err, "delete fail")
}

// --------------------- Helper ---------------------
func ptrString(s string) *string { return &s }
