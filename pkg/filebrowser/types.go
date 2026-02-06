package filebrowser

import (
	"context"

	corev1 "k8s.io/api/core/v1"
)

// Config holds FileBrowser pod and service configuration
type Config struct {
	Namespace   string
	PodName     string
	ServiceName string
	PVCName     string
	BaseURL     string
	ReadOnly    bool
	Labels      map[string]string
}

// Manager handles FileBrowser pod and service lifecycle
type Manager interface {
	// CreatePod creates a FileBrowser pod with the specified configuration
	CreatePod(ctx context.Context, cfg *Config) (*corev1.Pod, error)

	// GetPod retrieves an existing FileBrowser pod
	GetPod(ctx context.Context, namespace, podName string) (*corev1.Pod, error)

	// DeletePod deletes a FileBrowser pod
	DeletePod(ctx context.Context, namespace, podName string) error

	// CreateService creates a NodePort service for FileBrowser access
	CreateService(ctx context.Context, cfg *Config) (*corev1.Service, error)

	// GetService retrieves an existing FileBrowser service
	GetService(ctx context.Context, namespace, serviceName string) (*corev1.Service, error)

	// DeleteService deletes a FileBrowser service
	DeleteService(ctx context.Context, namespace, serviceName string) error

	// WaitForReady waits for pod to be in ready state
	WaitForReady(ctx context.Context, namespace, podName string) error

	// DeleteResources deletes both pod and service
	DeleteResources(ctx context.Context, namespace, podName, serviceName string) error
}

// SessionManager manages FileBrowser sessions with automatic cleanup
type SessionManager interface {
	// Start creates FileBrowser pod and service, returns NodePort
	Start(ctx context.Context, cfg *Config) (nodePort string, err error)

	// Stop deletes FileBrowser pod and service
	Stop(ctx context.Context, namespace, podName, serviceName string) error

	// GetOrCreate returns existing session or creates new one
	GetOrCreate(ctx context.Context, cfg *Config) (nodePort string, err error)
}
