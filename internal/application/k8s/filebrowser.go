package k8s

import (
	"context"
	"fmt"

	"github.com/linskybing/platform-go/pkg/filebrowser"
	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	pvcName := pvcNames[0]
	accessType := "rw"
	if readOnly {
		accessType = "ro"
	}

	cfg := &filebrowser.Config{
		Namespace:   ns,
		PodName:     fmt.Sprintf("fb-%s-%s", accessType, pvcName),
		ServiceName: fmt.Sprintf("fb-svc-%s-%s", accessType, pvcName),
		PVCName:     pvcName,
		BaseURL:     baseURL,
		ReadOnly:    readOnly,
		Labels: map[string]string{
			"pvc":         pvcName,
			"access-mode": accessType,
		},
	}

	nodePort, err := s.FileBrowserManager.sessionMgr.GetOrCreate(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to start file browser: %w", err)
	}

	return nodePort, nil
}

// StopFileBrowser stops the file browser instance in a namespace.
func (s *K8sService) StopFileBrowser(ctx context.Context, ns string) error {
	if ns == "" {
		return fmt.Errorf("namespace is required: %w", ErrMissingField)
	}

	// Stop all filebrowser pods/services in the namespace by listing and cleaning up
	// both read-write and read-only variants
	for _, accessType := range []string{"rw", "ro"} {
		pods, err := listFileBrowserPods(ctx, ns, accessType)
		if err != nil {
			continue
		}
		for _, podName := range pods {
			svcName := fmt.Sprintf("fb-svc-%s", podName[3:]) // strip "fb-" prefix, add "fb-svc-" prefix
			_ = s.FileBrowserManager.sessionMgr.Stop(ctx, ns, podName, svcName)
		}
	}

	return nil
}

// listFileBrowserPods lists filebrowser pods by access mode label in a namespace.
func listFileBrowserPods(ctx context.Context, ns, accessMode string) ([]string, error) {
	if k8sclient.Clientset == nil {
		return nil, nil
	}

	pods, err := k8sclient.Clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=filebrowser,access-mode=%s", accessMode),
	})
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(pods.Items))
	for _, pod := range pods.Items {
		names = append(names, pod.Name)
	}
	return names, nil
}
