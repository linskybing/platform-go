package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"
	"strings"

	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/pkg/filebrowser"
	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
)

type UserStorageManager struct {
	fbManager filebrowser.SessionManager
}

func NewUserStorageManager() *UserStorageManager {
	return &UserStorageManager{
		fbManager: filebrowser.NewSessionManager(),
	}
}

func (m *UserStorageManager) CheckExists(ctx context.Context, username string) (bool, error) {
	safeUser := strings.ToLower(username)
	nsName := fmt.Sprintf("user-%s-storage", safeUser)
	return k8sclient.CheckNamespaceExists(nsName)
}

func (m *UserStorageManager) Initialize(username string) error {
	safeUser := strings.ToLower(username)
	if reg, err := regexp.Compile("[^a-z0-9-]+"); err == nil {
		safeUser = reg.ReplaceAllString(safeUser, "-")
	}

	nsName := fmt.Sprintf("user-%s-storage", safeUser)
	pvcName := fmt.Sprintf("user-%s-disk", safeUser)

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

	slog.Info("successfully initialized storage resources", "username", username)
	return nil
}

func (m *UserStorageManager) Expand(username, newSize string) error {
	safeUser := strings.ToLower(username)
	nsName := fmt.Sprintf("user-%s-storage", safeUser)
	pvcName := fmt.Sprintf("user-%s-disk", safeUser)

	return k8sclient.ExpandPVC(nsName, pvcName, newSize)
}

func (m *UserStorageManager) Delete(ctx context.Context, username string) error {
	safeUser := strings.ToLower(username)
	nsName := fmt.Sprintf("user-%s-storage", safeUser)

	if err := k8sclient.DeleteNamespace(nsName); err != nil {
		return fmt.Errorf("failed to delete user storage namespace '%s': %w", nsName, err)
	}

	return nil
}

func (m *UserStorageManager) OpenFileBrowser(ctx context.Context, username string) (string, error) {
	safeUser := strings.ToLower(username)
	nsName := fmt.Sprintf("user-%s-storage", safeUser)
	pvcName := fmt.Sprintf("user-%s-disk", safeUser)
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
	safeUser := strings.ToLower(username)
	nsName := fmt.Sprintf("user-%s-storage", safeUser)
	podName := fmt.Sprintf("fb-hub-%s", safeUser)
	svcName := fmt.Sprintf("fb-hub-svc-%s", safeUser)

	return m.fbManager.Stop(ctx, nsName, podName, svcName)
}
