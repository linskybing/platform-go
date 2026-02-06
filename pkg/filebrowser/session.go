package filebrowser

import (
	"context"
	"fmt"
)

// sessionManager implements SessionManager interface
type sessionManager struct {
	mgr Manager
}

// NewSessionManager creates a new FileBrowser session manager
func NewSessionManager() SessionManager {
	return &sessionManager{
		mgr: NewManager(),
	}
}

// Start creates FileBrowser pod and service, returns NodePort
func (sm *sessionManager) Start(ctx context.Context, cfg *Config) (string, error) {
	// Create pod
	pod, err := sm.mgr.CreatePod(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("failed to create pod: %w", err)
	}

	// Create service
	svc, err := sm.mgr.CreateService(ctx, cfg)
	if err != nil {
		// Cleanup pod on service creation failure
		_ = sm.mgr.DeletePod(ctx, cfg.Namespace, cfg.PodName)
		return "", fmt.Errorf("failed to create service: %w", err)
	}

	// Wait for pod to be ready (best effort, don't fail if timeout)
	_ = sm.mgr.WaitForReady(ctx, cfg.Namespace, pod.Name)

	// Extract NodePort from service
	nodePort := ""
	if len(svc.Spec.Ports) > 0 {
		nodePort = fmt.Sprintf("%d", svc.Spec.Ports[0].NodePort)
	}

	return nodePort, nil
}

// Stop deletes FileBrowser pod and service
func (sm *sessionManager) Stop(ctx context.Context, namespace, podName, serviceName string) error {
	return sm.mgr.DeleteResources(ctx, namespace, podName, serviceName)
}

// GetOrCreate returns existing session or creates new one
func (sm *sessionManager) GetOrCreate(ctx context.Context, cfg *Config) (string, error) {
	// Check if pod already exists and is running
	existingPod, err := sm.mgr.GetPod(ctx, cfg.Namespace, cfg.PodName)
	if err == nil && existingPod.Status.Phase == "Running" {
		// Pod exists, get service
		svc, svcErr := sm.mgr.GetService(ctx, cfg.Namespace, cfg.ServiceName)
		if svcErr == nil && len(svc.Spec.Ports) > 0 {
			nodePort := fmt.Sprintf("%d", svc.Spec.Ports[0].NodePort)
			return nodePort, nil
		}
	}

	// Pod doesn't exist or not running, create new session
	return sm.Start(ctx, cfg)
}
