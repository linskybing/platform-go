package k8s

import (
	"context"
	"fmt"
)

// CheckUserStorageExists checks if a user's storage hub has been initialized.
func (s *K8sService) CheckUserStorageExists(ctx context.Context, username string) (bool, error) {
	if username == "" {
		return false, fmt.Errorf("username is required: %w", ErrMissingField)
	}

	exists, err := s.userStorageManager.CheckExists(ctx, username)
	if err != nil {
		return false, fmt.Errorf("failed to check storage existence for user %s: %w", username, err)
	}
	return exists, nil
}

// InitializeUserStorageHub initializes a new storage hub for a user.
func (s *K8sService) InitializeUserStorageHub(username, adminID string) error {
	if username == "" {
		return fmt.Errorf("username is required: %w", ErrMissingField)
	}

	if err := s.userStorageManager.Initialize(username, adminID); err != nil {
		return fmt.Errorf("failed to initialize storage hub for user %s: %w", username, err)
	}
	return nil
}

// ExpandUserStorageHub increases the capacity of a user's storage hub.
func (s *K8sService) ExpandUserStorageHub(username, newSize string) error {
	if username == "" {
		return fmt.Errorf("username is required: %w", ErrMissingField)
	}
	if newSize == "" {
		return fmt.Errorf("new size is required: %w", ErrMissingField)
	}

	if err := s.userStorageManager.Expand(username, newSize); err != nil {
		return fmt.Errorf("failed to expand storage hub for user %s to %s: %w", username, newSize, err)
	}
	return nil
}

// DeleteUserStorageHub deletes all storage resources for a user.
func (s *K8sService) DeleteUserStorageHub(ctx context.Context, username string) error {
	if username == "" {
		return fmt.Errorf("username is required: %w", ErrMissingField)
	}

	if err := s.userStorageManager.Delete(ctx, username); err != nil {
		return fmt.Errorf("failed to delete storage hub for user %s: %w", username, err)
	}
	return nil
}

// OpenUserGlobalFileBrowser starts a file browser instance for the user's global storage.
func (s *K8sService) OpenUserGlobalFileBrowser(ctx context.Context, username string) (string, error) {
	if username == "" {
		return "", fmt.Errorf("username is required: %w", ErrMissingField)
	}

	url, err := s.userStorageManager.OpenFileBrowser(ctx, username)
	if err != nil {
		return "", fmt.Errorf("failed to open file browser for user %s: %w", username, err)
	}
	return url, nil
}

// StopUserGlobalFileBrowser stops the file browser instance for the user's global storage.
func (s *K8sService) StopUserGlobalFileBrowser(ctx context.Context, username string) error {
	if username == "" {
		return fmt.Errorf("username is required: %w", ErrMissingField)
	}

	if err := s.userStorageManager.CloseFileBrowser(ctx, username); err != nil {
		return fmt.Errorf("failed to stop file browser for user %s: %w", username, err)
	}
	return nil
}
