package k8s

import (
	"context"
	"fmt"
)

// StartFileBrowser starts a file browser instance for a namespace.
func (s *K8sService) StartFileBrowser(ctx context.Context, ns string, pvcNames []string, readOnly bool, baseURL string) (string, error) {
	if ns == "" {
		return "", fmt.Errorf("namespace is required: %w", ErrMissingField)
	}
	if len(pvcNames) == 0 {
		return "", fmt.Errorf("at least one PVC name is required: %w", ErrInvalidInput)
	}
	if baseURL == "" {
		return "", fmt.Errorf("base URL is required: %w", ErrMissingField)
	}

	url, err := s.fileBrowserManager.Start(ctx, ns, pvcNames, readOnly, baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to start file browser in namespace %s with %d PVCs: %w", ns, len(pvcNames), err)
	}
	return url, nil
}

// StopFileBrowser stops the file browser instance in a namespace.
func (s *K8sService) StopFileBrowser(ctx context.Context, ns string) error {
	if ns == "" {
		return fmt.Errorf("namespace is required: %w", ErrMissingField)
	}

	if err := s.fileBrowserManager.Stop(ctx, ns); err != nil {
		return fmt.Errorf("failed to stop file browser in namespace %s: %w", ns, err)
	}
	return nil
}
