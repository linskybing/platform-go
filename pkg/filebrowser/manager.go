package filebrowser

import (
	"context"
	"fmt"
	"time"

	"github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	DefaultImage      = "filebrowser/filebrowser:v2"
	DefaultPort       = 80
	DefaultRunAsUser  = 1000
	DefaultRunAsGroup = 1000
	DefaultFSGroup    = 1000
)

// manager implements Manager interface for FileBrowser pod and service management
type manager struct{}

// NewManager creates a new FileBrowser manager
func NewManager() Manager {
	return &manager{}
}

// CreatePod creates a FileBrowser pod with the specified configuration
func (m *manager) CreatePod(ctx context.Context, cfg *Config) (*corev1.Pod, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	if k8s.Clientset == nil {
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: cfg.PodName, Namespace: cfg.Namespace},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		}, nil
	}

	pod := buildPodSpec(cfg)
	createdPod, err := k8s.Clientset.CoreV1().Pods(cfg.Namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return k8s.Clientset.CoreV1().Pods(cfg.Namespace).Get(ctx, cfg.PodName, metav1.GetOptions{})
		}
		return nil, fmt.Errorf("failed to create pod: %w", err)
	}

	logger.Info("created filebrowser pod",
		"pod", cfg.PodName,
		"namespace", cfg.Namespace,
		"read_only", cfg.ReadOnly)

	return createdPod, nil
}

// GetPod retrieves an existing FileBrowser pod
func (m *manager) GetPod(ctx context.Context, namespace, podName string) (*corev1.Pod, error) {
	if k8s.Clientset == nil {
		return &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: podName, Namespace: namespace},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		}, nil
	}

	return k8s.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
}

// DeletePod deletes a FileBrowser pod
func (m *manager) DeletePod(ctx context.Context, namespace, podName string) error {
	if k8s.Clientset == nil {
		logger.Debug("mock: deleted filebrowser pod", "pod", podName, "namespace", namespace)
		return nil
	}

	err := k8s.Clientset.CoreV1().Pods(namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete pod: %w", err)
	}

	logger.Info("deleted filebrowser pod", "pod", podName, "namespace", namespace)
	return nil
}

// CreateService creates a NodePort service for FileBrowser access
func (m *manager) CreateService(ctx context.Context, cfg *Config) (*corev1.Service, error) {
	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	if k8s.Clientset == nil {
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: cfg.ServiceName, Namespace: cfg.Namespace},
			Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{NodePort: 30000}}},
		}, nil
	}

	svc := buildServiceSpec(cfg)
	createdSvc, err := k8s.Clientset.CoreV1().Services(cfg.Namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return k8s.Clientset.CoreV1().Services(cfg.Namespace).Get(ctx, cfg.ServiceName, metav1.GetOptions{})
		}
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	logger.Info("created filebrowser service",
		"service", cfg.ServiceName,
		"namespace", cfg.Namespace)

	return createdSvc, nil
}

// GetService retrieves an existing FileBrowser service
func (m *manager) GetService(ctx context.Context, namespace, serviceName string) (*corev1.Service, error) {
	if k8s.Clientset == nil {
		return &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: serviceName, Namespace: namespace},
			Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{NodePort: 30000}}},
		}, nil
	}

	return k8s.Clientset.CoreV1().Services(namespace).Get(ctx, serviceName, metav1.GetOptions{})
}

// DeleteService deletes a FileBrowser service
func (m *manager) DeleteService(ctx context.Context, namespace, serviceName string) error {
	if k8s.Clientset == nil {
		logger.Debug("mock: deleted filebrowser service", "service", serviceName, "namespace", namespace)
		return nil
	}

	err := k8s.Clientset.CoreV1().Services(namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	logger.Info("deleted filebrowser service", "service", serviceName, "namespace", namespace)
	return nil
}

// WaitForReady waits for pod to be in ready state
func (m *manager) WaitForReady(ctx context.Context, namespace, podName string) error {
	if k8s.Clientset == nil {
		return nil
	}

	return wait.PollUntilContextTimeout(ctx, 2*time.Second, 30*time.Second, true, func(ctx context.Context) (bool, error) {
		pod, err := k8s.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if pod.Status.Phase == corev1.PodRunning {
			for _, cond := range pod.Status.Conditions {
				if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
					return true, nil
				}
			}
		}
		return false, nil
	})
}

// DeleteResources deletes both pod and service
func (m *manager) DeleteResources(ctx context.Context, namespace, podName, serviceName string) error {
	svcErr := m.DeleteService(ctx, namespace, serviceName)
	podErr := m.DeletePod(ctx, namespace, podName)

	if svcErr != nil {
		return fmt.Errorf("service deletion error: %w", svcErr)
	}
	if podErr != nil {
		return fmt.Errorf("pod deletion error: %w", podErr)
	}

	return nil
}

// buildPodSpec constructs the pod specification
func buildPodSpec(cfg *Config) *corev1.Pod {
	labels := cfg.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	labels["app"] = "filebrowser"
	labels["app.kubernetes.io/name"] = "filebrowser"
	labels["app.kubernetes.io/managed-by"] = "platform"

	args := []string{
		"--noauth",
		"--database", "/tmp/filebrowser.db",
		"--root", "/srv",
		"--port", fmt.Sprintf("%d", DefaultPort),
		"--address", "0.0.0.0",
	}

	if cfg.BaseURL != "" {
		args = append(args, "--baseURL", cfg.BaseURL)
	}

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.PodName,
			Namespace: cfg.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser:  int64Ptr(DefaultRunAsUser),
				RunAsGroup: int64Ptr(DefaultRunAsGroup),
				FSGroup:    int64Ptr(DefaultFSGroup),
			},
			Containers: []corev1.Container{
				{
					Name:  "filebrowser",
					Image: DefaultImage,
					Args:  args,
					Ports: []corev1.ContainerPort{{ContainerPort: DefaultPort}},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "data",
							MountPath: "/srv",
							ReadOnly:  cfg.ReadOnly,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: cfg.PVCName,
							ReadOnly:  cfg.ReadOnly,
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
}

// buildServiceSpec constructs the service specification
func buildServiceSpec(cfg *Config) *corev1.Service {
	labels := cfg.Labels
	if labels == nil {
		labels = make(map[string]string)
	}
	
	// Service selector must match pod labels
	selector := map[string]string{
		"app": "filebrowser",
	}
	for k, v := range labels {
		selector[k] = v
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cfg.ServiceName,
			Namespace: cfg.Namespace,
			Labels: map[string]string{
				"app": "filebrowser",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       DefaultPort,
					TargetPort: intstr.FromInt(DefaultPort),
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
}

// validateConfig validates the configuration
func validateConfig(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}
	if cfg.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if cfg.PodName == "" {
		return fmt.Errorf("pod name is required")
	}
	if cfg.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}
	if cfg.PVCName == "" {
		return fmt.Errorf("PVC name is required")
	}
	return nil
}

// int64Ptr returns a pointer to an int64 value
func int64Ptr(i int64) *int64 {
	return &i
}
