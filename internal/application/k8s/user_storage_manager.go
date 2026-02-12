package k8s

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/linskybing/platform-go/pkg/filebrowser"
	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
	"gorm.io/gorm"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

type UserStorageManager struct {
	fbManager   filebrowser.SessionManager
	storageRepo storage.StorageRepo
	userRepo    UserRepo
}

// UserRepo defines the interface for User database operations
type UserRepo interface {
	GetUserByUsername(username string) (user.User, error)
}

func NewUserStorageManager(storageRepo storage.StorageRepo, userRepo UserRepo) *UserStorageManager {
	return &UserStorageManager{
		fbManager:   filebrowser.NewSessionManager(),
		storageRepo: storageRepo,
		userRepo:    userRepo,
	}
}

func (m *UserStorageManager) CheckExists(ctx context.Context, username string) (bool, error) {
	safeUser := k8sclient.ToSafeK8sName(username)
	nsName := fmt.Sprintf(config.UserStorageNs, safeUser)
	exists, err := k8sclient.CheckNamespaceExists(nsName)
	if err == nil {
		return exists, nil
	}

	if apierrors.IsForbidden(err) || apierrors.IsUnauthorized(err) {
		if m.storageRepo == nil {
			return false, nil
		}
		userID := username
		if m.userRepo != nil {
			if u, err := m.userRepo.GetUserByUsername(username); err == nil {
				userID = u.UID
			}
		}
		stored, dbErr := m.storageRepo.GetUserStorageByUserID(ctx, userID)
		if dbErr == nil && stored != nil {
			return true, nil
		}
		if errors.Is(dbErr, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, dbErr
	}

	return false, err
}

func (m *UserStorageManager) Initialize(username, adminID string) error {
	ctx := context.Background()

	safeUser := k8sclient.ToSafeK8sName(username)

	nsName := fmt.Sprintf(config.UserStorageNs, safeUser)
	pvcName := fmt.Sprintf(config.UserStoragePVC, safeUser)

	slog.Info("initializing storage for user",
		"username", username,
		"namespace", nsName)

	if err := k8sclient.CreateNamespace(nsName); err != nil {
		slog.Warn("namespace creation warning",
			"namespace", nsName,
			"error", err)
	}

	if err := k8sclient.CreateHubPVC(nsName, pvcName, config.DefaultStorageClassName, config.UserPVSize); err != nil {
		return fmt.Errorf("failed to create hub pvc: %w", err)
	}

	// Get user ID from username
	var userID string
	if m.userRepo != nil {
		u, err := m.userRepo.GetUserByUsername(username)
		if err != nil {
			slog.Warn("failed to get user by username, using username as ID",
				"username", username,
				"error", err)
			userID = username
		} else {
			userID = u.UID
		}
	} else {
		userID = username
	}

	// Parse capacity from config (e.g., "10Gi" -> 10)
	capacity := parseGiCapacity(config.UserPVSize, 10)

	// Persist to database
	if m.storageRepo != nil {
		existing, err := m.storageRepo.GetUserStorageByUserID(ctx, userID)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("failed to load user storage from database",
				"username", username,
				"error", err)
		} else if existing != nil {
			existing.Name = pvcName
			existing.PVCName = pvcName
			existing.Capacity = capacity
			existing.StorageClass = config.DefaultStorageClassName
			if existing.CreatedBy == "" {
				existing.CreatedBy = adminID
			}
			if err := m.storageRepo.UpdateUserStorage(ctx, existing); err != nil {
				slog.Error("failed to update user storage in database",
					"username", username,
					"error", err)
			}
		} else {
			dbStorage := &storage.UserStorage{
				Name:         pvcName,
				UserID:       userID,
				PVCName:      pvcName,
				Capacity:     capacity,
				StorageClass: config.DefaultStorageClassName,
				CreatedBy:    adminID,
			}
			if err := m.storageRepo.CreateUserStorage(ctx, dbStorage); err != nil {
				slog.Error("failed to persist user storage to database",
					"username", username,
					"error", err)
			} else {
				slog.Info("user storage persisted to database",
					"username", username,
					"storage_id", dbStorage.ID)
			}
		}
	}

	slog.Info("successfully initialized storage resources", "username", username)
	return nil
}

func (m *UserStorageManager) Expand(username, newSize string) error {
	safeUser := k8sclient.ToSafeK8sName(username)
	nsName := fmt.Sprintf(config.UserStorageNs, safeUser)
	pvcName := fmt.Sprintf(config.UserStoragePVC, safeUser)

	if err := k8sclient.ExpandPVC(nsName, pvcName, newSize); err != nil {
		return err
	}

	if m.storageRepo != nil {
		capacity := parseGiCapacity(newSize, 0)
		if capacity > 0 {
			ctx := context.Background()
			if m.userRepo != nil {
				if u, err := m.userRepo.GetUserByUsername(username); err == nil {
					if existing, err := m.storageRepo.GetUserStorageByUserID(ctx, u.UID); err == nil && existing != nil {
						existing.Capacity = capacity
						if err := m.storageRepo.UpdateUserStorage(ctx, existing); err != nil {
							slog.Warn("failed to update user storage capacity",
								"username", username,
								"error", err)
						}
					}
				}
			}
		}
	}

	return nil
}

func (m *UserStorageManager) Delete(ctx context.Context, username string) error {
	safeUser := k8sclient.ToSafeK8sName(username)
	nsName := fmt.Sprintf(config.UserStorageNs, safeUser)

	// Get user ID to delete from database
	var userID string
	if m.userRepo != nil {
		u, err := m.userRepo.GetUserByUsername(username)
		if err != nil {
			slog.Warn("failed to get user by username for deletion",
				"username", username,
				"error", err)
			userID = username
		} else {
			userID = u.UID
		}
	} else {
		userID = username
	}

	// Delete from database first
	if m.storageRepo != nil {
		if err := m.storageRepo.DeleteUserStorageByUserID(ctx, userID); err != nil {
			slog.Error("failed to delete user storage from database",
				"username", username,
				"user_id", userID,
				"error", err)
		} else {
			slog.Info("user storage deleted from database",
				"username", username,
				"user_id", userID)
		}
	}

	// Delete from Kubernetes
	if err := k8sclient.DeleteNamespace(nsName); err != nil {
		return fmt.Errorf("failed to delete user storage namespace '%s': %w", nsName, err)
	}

	return nil
}

func (m *UserStorageManager) OpenFileBrowser(ctx context.Context, username string) (string, error) {
	safeUser := k8sclient.ToSafeK8sName(username)
	nsName := fmt.Sprintf(config.UserStorageNs, safeUser)
	pvcName := fmt.Sprintf(config.UserStoragePVC, safeUser)
	podName := fmt.Sprintf("fb-hub-%s", safeUser)
	svcName := fmt.Sprintf("fb-hub-svc-%s", safeUser)

	cfg := &filebrowser.Config{
		Namespace:   nsName,
		PodName:     podName,
		ServiceName: svcName,
		PVCName:     pvcName,
		ReadOnly:    false,
	}

	nodePort, err := m.fbManager.GetOrCreate(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to start filebrowser: %w", err)
	}

	return nodePort, nil
}

func (m *UserStorageManager) CloseFileBrowser(ctx context.Context, username string) error {
	safeUser := k8sclient.ToSafeK8sName(username)
	nsName := fmt.Sprintf(config.UserStorageNs, safeUser)
	podName := fmt.Sprintf("fb-hub-%s", safeUser)
	svcName := fmt.Sprintf("fb-hub-svc-%s", safeUser)

	return m.fbManager.Stop(ctx, nsName, podName, svcName)
}

func parseGiCapacity(value string, fallback int) int {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimSuffix(trimmed, "Gi")
	if trimmed == "" {
		return fallback
	}
	capacity, err := strconv.Atoi(trimmed)
	if err != nil {
		return fallback
	}
	return capacity
}
