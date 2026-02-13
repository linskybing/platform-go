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
	// Validate username
	if err := validateUsername(input.Username); err != nil {
		return err
	}

	// Validate password
	if err := validatePassword(input.Password); err != nil {
		return err
	}

	_, err := s.Repos.User.GetUserByUsername(input.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
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
		Username: input.Username,
		Password: string(hashed),
		Email:    input.Email,
		FullName: input.FullName,
		Type:     "origin",
		Status:   "offline",
	}

	if input.Type != nil {
		usr.Type = *input.Type
	}
	if input.Status != nil {
		usr.Status = *input.Status
	}
	if err := s.Repos.User.SaveUser(&usr); err != nil {
		return err
	}
	s.invalidateUserListCache()
	return nil
}

func (s *UserService) LoginUser(username, password string) (user.User, string, bool, error) {
	usr, err := s.Repos.User.GetUserByUsername(username)
	if err != nil {
		return user.User{}, "", false, fmt.Errorf("invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(password)); err != nil {
		return user.User{}, "", false, fmt.Errorf("invalid credentials")
	}

	token, isAdmin, err := middleware.GenerateToken(usr.UID, usr.Username, 24*time.Hour, s.Repos.UserGroup)
	if err != nil {
		return user.User{}, "", false, err
	}

	return usr, token, isAdmin, nil
}

func (s *UserService) ListUsers() ([]user.UserWithSuperAdmin, error) {
	if s.cache != nil && s.cache.Enabled() {
		var cached []user.UserWithSuperAdmin
		if err := s.cache.GetJSON(context.Background(), userListKey(), &cached); err == nil {
			return cached, nil
		}
	}

	users, err := s.Repos.User.GetAllUsers()
	if err != nil {
		return nil, err
	}
	if s.cache != nil && s.cache.Enabled() {
		_ = s.cache.AsyncSetJSON(context.Background(), userListKey(), users, userCacheTTL)
	}
	return users, nil
}

func (s *UserService) ListUserByPaging(page, limit int) ([]user.UserWithSuperAdmin, error) {
	if s.cache != nil && s.cache.Enabled() {
		var cached []user.UserWithSuperAdmin
		if err := s.cache.GetJSON(context.Background(), userListPagingKey(page, limit), &cached); err == nil {
			return cached, nil
		}
	}

	users, err := s.Repos.User.ListUsersPaging(page, limit)
	if err != nil {
		return nil, err
	}
	if s.cache != nil && s.cache.Enabled() {
		_ = s.cache.AsyncSetJSON(context.Background(), userListPagingKey(page, limit), users, userCacheTTL)
	}
	return users, nil
}

func (s *UserService) FindUserByID(id string) (user.UserWithSuperAdmin, error) {
	if s.cache != nil && s.cache.Enabled() {
		var cached user.UserWithSuperAdmin
		if err := s.cache.GetJSON(context.Background(), userByIDKey(id), &cached); err == nil {
			return cached, nil
		}
	}

	usr, err := s.Repos.User.GetUserByID(id)
	if err != nil {
		return user.UserWithSuperAdmin{}, err
	}
	if s.cache != nil && s.cache.Enabled() {
		_ = s.cache.AsyncSetJSON(context.Background(), userByIDKey(id), usr, userCacheTTL)
	}
	return usr, nil
}

func (s *UserService) UpdateUser(id string, input user.UpdateUserInput) (user.User, error) {
	usr, err := s.Repos.User.GetUserRawByID(id)
	if err != nil {
		return user.User{}, ErrUserNotFound
	}

	// Prevent downgrading the reserved admin user
	if usr.Username == config.ReservedAdminUsername && input.Type != nil && *input.Type != "admin" {
		return user.User{}, ErrReservedAdminUser
	}

	if input.Password != nil {
		if input.OldPassword == nil {
			return user.User{}, ErrMissingOldPassword
		}
		if err := bcrypt.CompareHashAndPassword([]byte(usr.Password), []byte(*input.OldPassword)); err != nil {
			return user.User{}, ErrIncorrectPassword
		}

		// Validate new password
		if err := validatePassword(*input.Password); err != nil {
			return user.User{}, err
		}

		hashed, err := bcrypt.GenerateFromPassword([]byte(*input.Password), constants.BcryptCost)
		if err != nil {
			return user.User{}, ErrPasswordHashFailure
		}
		usr.Password = string(hashed)
	}

	if input.Type != nil {
		usr.Type = string(*input.Type)
	}
	if input.Status != nil {
		usr.Status = string(*input.Status)
	}
	if input.Email != nil {
		usr.Email = input.Email
	}
	if input.FullName != nil {
		usr.FullName = input.FullName
	}

	if err := s.Repos.User.SaveUser(&usr); err != nil {
		return user.User{}, err
	}
	s.invalidateUserCache(id)
	return usr, nil
}

func (s *UserService) RemoveUser(id string) error {
	usr, err := s.Repos.User.GetUserRawByID(id)
	if err != nil {
		return ErrUserNotFound
	}

	// Prevent deleting the reserved admin user
	if usr.Username == config.ReservedAdminUsername {
		return ErrReservedAdminUser
	}

	if err := s.Repos.User.DeleteUser(id); err != nil {
		return err
	}
	s.invalidateUserCache(id)
	return nil
}

// validateUsername checks if username meets requirements
func validateUsername(username string) error {
	username = strings.TrimSpace(username)

	if len(username) < constants.MinUsernameLength || len(username) > constants.MaxUsernameLength {
		return ErrInvalidUsername
	}

	// Allow only alphanumeric, underscore, and hyphen
	for _, r := range username {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && (r < '0' || r > '9') && r != '_' && r != '-' {
			return ErrInvalidUsername
		}
	}

	return nil
}

// validatePassword checks if password meets minimum requirements
func validatePassword(password string) error {
	if len(password) < constants.MinPasswordLength {
		return ErrInvalidPassword
	}
	return nil
}

func userListKey() string {
	return "cache:user:list"
}

func userListPagingKey(page, limit int) string {
	return fmt.Sprintf("cache:user:list:%d:%d", page, limit)
}

func userByIDKey(id string) string {
	return fmt.Sprintf("cache:user:by-id:%s", id)
}

func (s *UserService) invalidateUserCache(id string) {
	if s.cache == nil || !s.cache.Enabled() {
		return
	}
	ctx := context.Background()
	_ = s.cache.Invalidate(ctx, userByIDKey(id), userListKey())
	_ = s.cache.InvalidatePrefix(ctx, "cache:user:list:")
}

// GetSettings returns user settings, creating defaults if none exist.
func (s *UserService) GetSettings(ctx context.Context, userID string) (*user.UserSettings, error) {
	var settings user.UserSettings
	db := s.Repos.DB()
	err := db.WithContext(ctx).Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaults := user.DefaultSettings(userID)
			if createErr := db.WithContext(ctx).Create(defaults).Error; createErr != nil {
				return nil, fmt.Errorf("failed to create default settings: %w", createErr)
			}
			return defaults, nil
		}
		return nil, fmt.Errorf("failed to get user settings: %w", err)
	}
	return &settings, nil
}

// UpdateSettings updates user settings, creating them if they don't exist.
func (s *UserService) UpdateSettings(ctx context.Context, userID string, updates map[string]interface{}) (*user.UserSettings, error) {
	db := s.Repos.DB()

	// Ensure settings exist
	var settings user.UserSettings
	err := db.WithContext(ctx).Where("user_id = ?", userID).First(&settings).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		settings = *user.DefaultSettings(userID)
		if createErr := db.WithContext(ctx).Create(&settings).Error; createErr != nil {
			return nil, fmt.Errorf("failed to create settings: %w", createErr)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	if err := db.WithContext(ctx).Model(&settings).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("failed to update settings: %w", err)
	}

	// Re-fetch updated settings
	if err := db.WithContext(ctx).Where("user_id = ?", userID).First(&settings).Error; err != nil {
		return nil, fmt.Errorf("failed to get updated settings: %w", err)
	}
	return &settings, nil
}

func (s *UserService) invalidateUserListCache() {
	if s.cache == nil || !s.cache.Enabled() {
		return
	}
	ctx := context.Background()
	_ = s.cache.Invalidate(ctx, userListKey())
	_ = s.cache.InvalidatePrefix(ctx, "cache:user:list:")
}
