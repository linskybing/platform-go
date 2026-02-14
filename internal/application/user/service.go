package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/linskybing/platform-go/internal/api/middleware"
	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/constants"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrIncorrectPassword   = errors.New("old password is incorrect")
	ErrMissingOldPassword  = errors.New("old password is required to change password")
	ErrPasswordHashFailure = errors.New("failed to hash new password")
	ErrUsernameTaken       = errors.New("username already taken")
	ErrReservedAdminUser   = errors.New("cannot delete or downgrade reserved admin user '" + config.ReservedAdminUsername + "'")
	ErrInvalidUsername     = errors.New("username must be 3-50 characters and contain only alphanumeric, underscore, or hyphen")
	ErrInvalidPassword     = errors.New("password must be at least 8 characters")
)

type UserService struct {
	Repos *repository.Repos
	cache *cache.Service
}

func NewUserService(repos *repository.Repos) *UserService {
	return NewUserServiceWithCache(repos, nil)
}

func NewUserServiceWithCache(repos *repository.Repos, cacheSvc *cache.Service) *UserService {
	return &UserService{
		Repos: repos,
		cache: cacheSvc,
	}
}

const userCacheTTL = 5 * time.Minute

func (s *UserService) RegisterUser(input user.CreateUserInput) error {
	ctx := context.Background()
	if err := validateUsername(input.Username); err != nil {
		return err
	}
	if err := validatePassword(input.Password); err != nil {
		return err
	}

	_, err := s.Repos.User.GetByUsername(ctx, input.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		return ErrUsernameTaken
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), constants.BcryptCost)
	if err != nil {
		return ErrPasswordHashFailure
	}

	usr := user.User{
		Username:     input.Username,
		PasswordHash: string(hashed),
		Email:        getValue(input.Email),
		FullName:     getValue(input.FullName),
		Type:         "origin",
		Status:       "offline",
	}

	if input.Type != nil {
		usr.Type = *input.Type
	}
	if input.Status != nil {
		usr.Status = *input.Status
	}

	if err := s.Repos.User.SaveUser(ctx, &usr); err != nil {
		return err
	}
	s.invalidateUserListCache()
	return nil
}

func (s *UserService) LoginUser(username, password string) (user.User, string, bool, error) {
	ctx := context.Background()
	usr, err := s.Repos.User.GetByUsername(ctx, username)
	if err != nil {
		return user.User{}, "", false, fmt.Errorf("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(usr.PasswordHash), []byte(password)); err != nil {
		return user.User{}, "", false, fmt.Errorf("invalid credentials")
	}

	token, isAdmin, err := middleware.GenerateToken(usr.ID, usr.Username, 24*time.Hour, s.Repos.UserGroup)
	if err != nil {
		return user.User{}, "", false, err
	}

	return *usr, token, isAdmin, nil
}

func (s *UserService) ListUsers() ([]user.UserWithSuperAdmin, error) {
	ctx := context.Background()
	if s.cache != nil && s.cache.Enabled() {
		var cached []user.UserWithSuperAdmin
		if err := s.cache.GetJSON(ctx, userListKey(), &cached); err == nil {
			return cached, nil
		}
	}

	users, err := s.Repos.User.List(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]user.UserWithSuperAdmin, len(users))
	for i, u := range users {
		result[i] = user.UserWithSuperAdmin{User: u}
	}

	if s.cache != nil && s.cache.Enabled() {
		_ = s.cache.AsyncSetJSON(ctx, userListKey(), result, userCacheTTL)
	}
	return result, nil
}

func (s *UserService) ListUserByPaging(page, limit int) ([]user.UserWithSuperAdmin, error) {
	ctx := context.Background()
	offset := (page - 1) * limit
	users, _, err := s.Repos.User.ListUsersPaging(ctx, offset, limit)
	if err != nil {
		return nil, err
	}
	result := make([]user.UserWithSuperAdmin, len(users))
	for i, u := range users {
		result[i] = user.UserWithSuperAdmin{User: u}
	}
	return result, nil
}

func (s *UserService) FindUserByID(id string) (user.UserWithSuperAdmin, error) {
	ctx := context.Background()
	usr, err := s.Repos.User.Get(ctx, id)
	if err != nil {
		return user.UserWithSuperAdmin{}, err
	}
	return user.UserWithSuperAdmin{User: *usr}, nil
}

func (s *UserService) UpdateUser(id string, input user.UpdateUserInput) (user.User, error) {
	ctx := context.Background()
	usr, err := s.Repos.User.Get(ctx, id)
	if err != nil {
		return user.User{}, ErrUserNotFound
	}

	if usr.Username == config.ReservedAdminUsername && input.Type != nil && *input.Type != "admin" {
		return user.User{}, ErrReservedAdminUser
	}

	if input.Password != nil {
		if input.OldPassword == nil {
			return user.User{}, ErrMissingOldPassword
		}
		if err := bcrypt.CompareHashAndPassword([]byte(usr.PasswordHash), []byte(*input.OldPassword)); err != nil {
			return user.User{}, ErrIncorrectPassword
		}
		hashed, err := bcrypt.GenerateFromPassword([]byte(*input.Password), constants.BcryptCost)
		if err != nil {
			return user.User{}, ErrPasswordHashFailure
		}
		usr.PasswordHash = string(hashed)
	}

	if input.Type != nil {
		usr.Type = *input.Type
	}
	if input.Status != nil {
		usr.Status = *input.Status
	}
	if input.Email != nil {
		usr.Email = *input.Email
	}
	if input.FullName != nil {
		usr.FullName = *input.FullName
	}

	if err := s.Repos.User.SaveUser(ctx, usr); err != nil {
		return user.User{}, err
	}
	s.invalidateUserCache(id)
	return *usr, nil
}

func (s *UserService) RemoveUser(id string) error {
	ctx := context.Background()
	usr, err := s.Repos.User.Get(ctx, id)
	if err != nil {
		return ErrUserNotFound
	}

	if usr.Username == config.ReservedAdminUsername {
		return ErrReservedAdminUser
	}

	return s.Repos.User.Delete(ctx, id)
}

func (s *UserService) GetSettings(ctx context.Context, userID string) (*user.UserSettings, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var settings user.UserSettings
	err := s.Repos.DB().WithContext(ctx).First(&settings, "user_id = ?", userID).Error
	if err == nil {
		return &settings, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	defaults := user.DefaultSettings(userID)
	if err := s.Repos.DB().WithContext(ctx).Create(defaults).Error; err != nil {
		return nil, err
	}
	return defaults, nil
}

func (s *UserService) UpdateSettings(ctx context.Context, userID string, updates map[string]interface{}) (*user.UserSettings, error) {
	settings, err := s.GetSettings(ctx, userID)
	if err != nil {
		return nil, err
	}
	if updates == nil {
		return settings, nil
	}
	if v, ok := updates["theme"].(string); ok {
		settings.Theme = v
	}
	if v, ok := updates["language"].(string); ok {
		settings.Language = v
	}
	if v, ok := updates["receive_notifications"].(bool); ok {
		settings.ReceiveNotifications = v
	}
	if err := s.Repos.DB().WithContext(ctx).Save(settings).Error; err != nil {
		return nil, err
	}
	return settings, nil
}

func validateUsername(username string) error {
	username = strings.TrimSpace(username)
	if len(username) < constants.MinUsernameLength || len(username) > constants.MaxUsernameLength {
		return ErrInvalidUsername
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < constants.MinPasswordLength {
		return ErrInvalidPassword
	}
	return nil
}

func getValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func userListKey() string          { return "cache:user:list" }
func userByIDKey(id string) string { return fmt.Sprintf("cache:user:by-id:%s", id) }

func (s *UserService) invalidateUserCache(id string) {
	if s.cache == nil || !s.cache.Enabled() {
		return
	}
	ctx := context.Background()
	_ = s.cache.Invalidate(ctx, userByIDKey(id), userListKey())
}

func (s *UserService) invalidateUserListCache() {
	if s.cache == nil || !s.cache.Enabled() {
		return
	}
	ctx := context.Background()
	_ = s.cache.Invalidate(ctx, userListKey())
}
