package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/linskybing/platform-go/internal/domain/storage"
	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
)

// FileBrowserManager handles FileBrowser pod creation with permission-based routing
type FileBrowserManager struct {
	// TODO: Inject PermissionManager via constructor once handler is reorganized
}

// NewFileBrowserManager creates a new FileBrowserManager instance
func NewFileBrowserManager() *FileBrowserManager {
	return &FileBrowserManager{}
}

// GetFileBrowserAccess creates or routes to appropriate FileBrowser pod based on user permission
// - Read-write permission -> read-write pod
// - Read-only permission -> read-only pod
// - No permission -> return unauthorized error
func (fbm *FileBrowserManager) GetFileBrowserAccess(ctx context.Context, req *storage.FileBrowserAccessRequest) (*storage.FileBrowserAccessResponse, error) {
	startTime := time.Now()

	// TODO: Implement permission checking once PermissionManager is available
	// For now, assume read access is allowed

	// Temporary: log access attempt
	logger.Info("filebrowser access request",
		"user_id", req.UserID,
		"group_id", req.GroupID,
		"pvc_id", req.PVCID)

	// Get group PVC info to find namespace and PVC name
	groupNamespace := fmt.Sprintf("group-%d-storage", req.GroupID)

	// Find the actual PVC name from labels
	pvcs, err := fbm.listPVCsByID(ctx, groupNamespace, req.PVCID)
	if err != nil || len(pvcs) == 0 {
		return nil, fmt.Errorf("PVC not found: %s", req.PVCID)
	}
	pvcName := pvcs[0].Name

	// TODO: Determine pod type based on permission from PermissionManager
	// Temporary: default to read-only
	readOnly := true

	// Create or get existing FileBrowser pod
	pod, service, err := fbm.ensureFileBrowserPod(ctx, groupNamespace, pvcName, readOnly, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to create filebrowser pod: %w", err)
	}

	// Wait for pod to be ready
	if err := fbm.waitForPodReady(ctx, groupNamespace, pod.Name, 30*time.Second); err != nil {
		logger.Error("filebrowser pod not ready in time",
			"pod", pod.Name,
			"namespace", groupNamespace,
			"error", err)
	}

	// Get NodePort from service
	nodePort := ""
	if len(service.Spec.Ports) > 0 {
		nodePort = fmt.Sprintf("%d", service.Spec.Ports[0].NodePort)
	}

	logger.Info("filebrowser access granted",
		"user_id", req.UserID,
		"group_id", req.GroupID,
		"pvc_id", req.PVCID,
		"read_only", readOnly,
		"pod", pod.Name,
		"duration_ms", time.Since(startTime).Milliseconds())

	return &storage.FileBrowserAccessResponse{
		Allowed:  true,
		URL:      fmt.Sprintf("http://nodeip:%s", nodePort), // Replace with actual node IP in production
		Port:     nodePort,
		PodName:  pod.Name,
		ReadOnly: readOnly,
		Message:  "Access granted",
	}, nil
}

// ensureFileBrowserPod creates or retrieves existing FileBrowser pod
func (fbm *FileBrowserManager) ensureFileBrowserPod(ctx context.Context, namespace, pvcName string, readOnly bool, userID uint) (*corev1.Pod, *corev1.Service, error) {
	if k8sclient.Clientset == nil {
		logger.Info("k8s client not initialized, using mock filebrowser")
		return &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{Name: "mock-fb-pod"},
				Status:     corev1.PodStatus{Phase: corev1.PodRunning},
			}, &corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{{NodePort: 30000}},
				},
			}, nil
	}

	// Generate unique pod name based on PVC and access type
	accessType := "rw"
	if readOnly {
		accessType = "ro"
	}
	podName := fmt.Sprintf("fb-%s-%s", accessType, pvcName)
	svcName := fmt.Sprintf("fb-svc-%s-%s", accessType, pvcName)

	// Check if pod already exists
	existingPod, err := k8sclient.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err == nil && existingPod.Status.Phase == corev1.PodRunning {
		// Pod exists and running, get service
		svc, svcErr := k8sclient.Clientset.CoreV1().Services(namespace).Get(ctx, svcName, metav1.GetOptions{})
		if svcErr == nil {
			logger.Info("reusing existing filebrowser pod", "pod", podName, "namespace", namespace)
			return existingPod, svc, nil
		}
	}

	// Create new pod
	pod := fbm.buildFileBrowserPod(podName, namespace, pvcName, readOnly)
	createdPod, err := k8sclient.Clientset.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, nil, fmt.Errorf("failed to create pod: %w", err)
	}
	if apierrors.IsAlreadyExists(err) {
		createdPod, _ = k8sclient.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	}

	// Create service
	svc := fbm.buildFileBrowserService(svcName, podName, namespace)
	createdSvc, err := k8sclient.Clientset.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return nil, nil, fmt.Errorf("failed to create service: %w", err)
	}
	if apierrors.IsAlreadyExists(err) {
		createdSvc, _ = k8sclient.Clientset.CoreV1().Services(namespace).Get(ctx, svcName, metav1.GetOptions{})
	}

	logger.Info("created filebrowser pod and service",
		"pod", podName,
		"service", svcName,
		"namespace", namespace,
		"read_only", readOnly)

	return createdPod, createdSvc, nil
}

// buildFileBrowserPod constructs the pod specification
func (fbm *FileBrowserManager) buildFileBrowserPod(podName, namespace, pvcName string, readOnly bool) *corev1.Pod {
	args := []string{"--noauth", "--root=/srv", "--address=0.0.0.0"}

	// Mount as read-only if permission is read-only
	mountReadOnly := readOnly

	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":                          "filebrowser",
				"pvc":                          pvcName,
				"access-mode":                  map[bool]string{true: "ro", false: "rw"}[readOnly],
				"app.kubernetes.io/name":       "filebrowser",
				"app.kubernetes.io/managed-by": "platform",
			},
		},
		Spec: corev1.PodSpec{
			SecurityContext: &corev1.PodSecurityContext{
				RunAsUser:  int64Ptr(1000),
				RunAsGroup: int64Ptr(1000),
				FSGroup:    int64Ptr(1000),
			},
			Containers: []corev1.Container{
				{
					Name:  "filebrowser",
					Image: "filebrowser/filebrowser:v2",
					Args:  args,
					Ports: []corev1.ContainerPort{{ContainerPort: 80}},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "data",
							MountPath: "/srv",
							ReadOnly:  mountReadOnly,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvcName,
							ReadOnly:  mountReadOnly,
						},
					},
				},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}
}

// buildFileBrowserService constructs the service specification
func (fbm *FileBrowserManager) buildFileBrowserService(svcName, podName, namespace string) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: namespace,
			Labels: map[string]string{
				"app": "filebrowser",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "filebrowser", "pvc": podName},
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       80,
					TargetPort: intstr.FromInt(80),
					NodePort:   0, // Auto-assign
				},
			},
			Type: corev1.ServiceTypeNodePort,
		},
	}
}

// waitForPodReady waits for pod to be ready
func (fbm *FileBrowserManager) waitForPodReady(ctx context.Context, namespace, podName string, timeout time.Duration) error {
	return wait.PollUntilContextTimeout(ctx, 2*time.Second, timeout, true, func(ctx context.Context) (bool, error) {
		pod, err := k8sclient.Clientset.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
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

// listPVCsByID finds PVCs with matching ID label
func (fbm *FileBrowserManager) listPVCsByID(ctx context.Context, namespace, pvcID string) ([]corev1.PersistentVolumeClaim, error) {
	if k8sclient.Clientset == nil {
		return []corev1.PersistentVolumeClaim{{ObjectMeta: metav1.ObjectMeta{Name: "mock-pvc"}}}, nil
	}

	// Extract UUID from PVC ID (format: group-{gid}-{uuid})
	uuid := pvcID[len("group-XX-"):] // Simplified extraction

	listOpts := metav1.ListOptions{
		LabelSelector: fmt.Sprintf("pvc-uuid=%s", uuid),
	}

	pvcList, err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(namespace).List(ctx, listOpts)
	if err != nil {
		return nil, err
	}

	return pvcList.Items, nil
}

// Helper function
func int64Ptr(i int64) *int64 {
	return &i
}
