package k8s

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/linskybing/platform-go/internal/domain/storage"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/linskybing/platform-go/pkg/cache"
	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
	"github.com/linskybing/platform-go/pkg/logger"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PVCBindingManager handles project PVC bindings to group storage
type PVCBindingManager struct {
	sm    *StorageManager
	pm    *PermissionManager
	repos *repository.Repos
}

// NewPVCBindingManager creates a new PVCBindingManager
func NewPVCBindingManager(repos *repository.Repos, cacheSvc *cache.Service) *PVCBindingManager {
	return &PVCBindingManager{
		sm:    NewStorageManager(repos, cacheSvc),
		pm:    &PermissionManager{repos: repos},
		repos: repos,
	}
}

// CreateProjectPVCBinding creates a PVC in user's project namespace that binds to group storage
// This allows users to mount group storage in their workloads
func (pbm *PVCBindingManager) CreateProjectPVCBinding(ctx context.Context, req *storage.CreateProjectPVCBindingRequest, userID string) (*storage.ProjectPVCBindingInfo, error) {
	startTime := time.Now()

	// Extract group ID from PVC ID (format: group-{gid}-{suffix})
	groupID, err := extractGroupIDFromPVCID(req.GroupPVCID)
	if err != nil {
		return nil, fmt.Errorf("invalid group PVC ID: %w", err)
	}

	// Check user permission for the group PVC
	perm, err := pbm.pm.GetUserPermission(ctx, userID, groupID, req.GroupPVCID)
	if err != nil {
		return nil, fmt.Errorf("failed to check permission: %w", err)
	}

	if !perm.CanRead() {
		logger.Warn("user attempted to bind PVC without read permission",
			"user_id", userID,
			"group_id", groupID,
			"pvc_id", req.GroupPVCID)
		return nil, ErrPermissionDenied
	}

	// Get source group PVC to find the PV name
	groupPVCs, err := pbm.sm.ListGroupPVCs(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to list group PVCs: %w", err)
	}

	var sourcePVC *storage.GroupPVCSpec
	for _, pvc := range groupPVCs {
		if pvc.ID == req.GroupPVCID {
			sourcePVC = &pvc
			break
		}
	}
	if sourcePVC == nil {
		return nil, fmt.Errorf("group PVC not found: %s", req.GroupPVCID)
	}

	// Get the PV name from the bound PVC
	pvName, err := pbm.getPVNameFromPVC(ctx, sourcePVC.Namespace, sourcePVC.PVCName)
	if err != nil {
		return nil, fmt.Errorf("failed to get PV name: %w", err)
	}

	projectNamespace, err := pbm.resolveProjectNamespace(ctx, req.ProjectID, userID)
	if err != nil {
		return nil, err
	}

	// Ensure project namespace exists
	if err := pbm.ensureProjectNamespace(ctx, projectNamespace, req.ProjectID); err != nil {
		return nil, fmt.Errorf("failed to ensure project namespace: %w", err)
	}

	// Determine access mode based on permission
	accessMode := corev1.ReadOnlyMany
	if perm.CanWrite() && !req.ReadOnly {
		accessMode = corev1.ReadWriteMany
	}

	// Compute size from capacity
	pvcSize := fmt.Sprintf("%dGi", sourcePVC.Capacity)

	labels := map[string]string{
		"app.kubernetes.io/managed-by": "platform",
		"binding-type":                 "project-group-binding",
		"group-id":                     groupID,
		"group-pvc-id":                 req.GroupPVCID,
		"project-id":                   req.ProjectID,
		"user-id":                      userID,
	}

	// Create PVC in project namespace that binds to the same PV
	binding, err := pbm.createBindingPVC(ctx, projectNamespace, req.PVCName, pvName, pvcSize, accessMode, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to create binding PVC: %w", err)
	}

	logger.Info("created project PVC binding",
		"project_id", req.ProjectID,
		"user_id", userID,
		"group_pvc_id", req.GroupPVCID,
		"project_pvc", req.PVCName,
		"access_mode", accessMode,
		"duration_ms", time.Since(startTime).Milliseconds())

	return &storage.ProjectPVCBindingInfo{
		ID:               fmt.Sprintf("%s/%s", projectNamespace, req.PVCName),
		ProjectID:        req.ProjectID,
		GroupPVCID:       req.GroupPVCID,
		ProjectPVCName:   req.PVCName,
		ProjectNamespace: projectNamespace,
		AccessMode:       string(accessMode),
		Status:           bindingStatus(binding),
		CreatedAt:        time.Now(),
	}, nil
}

// createBindingPVC creates a PVC that binds to an existing PV
func (pbm *PVCBindingManager) createBindingPVC(ctx context.Context, namespace, pvcName, pvName, size string, accessMode corev1.PersistentVolumeAccessMode, labels map[string]string) (*corev1.PersistentVolumeClaim, error) {
	if k8sclient.Clientset == nil {
		return &corev1.PersistentVolumeClaim{Status: corev1.PersistentVolumeClaimStatus{Phase: corev1.ClaimBound}}, nil
	}

	qty, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, fmt.Errorf("invalid size %s: %w", size, err)
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: namespace,
			Annotations: map[string]string{
				"pv.kubernetes.io/bind-completed":      "yes",
				"pv.kubernetes.io/bound-by-controller": "yes",
			},
			Labels: labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			VolumeName:  pvName,
			AccessModes: []corev1.PersistentVolumeAccessMode{accessMode},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: qty,
				},
			},
			// No storage class needed for static binding
			StorageClassName: nil,
		},
	}

	return k8sclient.Clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
}

// getPVNameFromPVC finds the PV name bound to a PVC
func (pbm *PVCBindingManager) getPVNameFromPVC(ctx context.Context, namespace, pvcName string) (string, error) {
	if k8sclient.Clientset == nil {
		return "mock-pv", nil
	}

	pvc, err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if pvc.Spec.VolumeName == "" {
		return "", fmt.Errorf("PVC %s/%s is not bound to a PV", namespace, pvcName)
	}

	return pvc.Spec.VolumeName, nil
}

// ensureProjectNamespace ensures the project namespace exists
func (pbm *PVCBindingManager) ensureProjectNamespace(ctx context.Context, namespace string, projectID string) error {
	if k8sclient.Clientset == nil {
		return nil
	}

	_, err := k8sclient.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"managed-by": "platform",
				"type":       "project",
				"project-id": projectID,
			},
		},
	}

	_, err = k8sclient.Clientset.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return err
	}

	logger.Info("ensured project namespace", "namespace", namespace, "project_id", projectID)
	return nil
}

// ListProjectPVCBindings lists all PVC bindings for a project
func (pbm *PVCBindingManager) ListProjectPVCBindings(ctx context.Context, projectID, userID string) ([]storage.ProjectPVCBindingInfo, error) {
	projectNamespace, err := pbm.resolveProjectNamespace(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if k8sclient.Clientset == nil {
		return []storage.ProjectPVCBindingInfo{}, nil
	}

	list, err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(projectNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: "binding-type=project-group-binding",
	})
	if err != nil {
		return nil, err
	}

	result := make([]storage.ProjectPVCBindingInfo, 0, len(list.Items))
	for _, pvc := range list.Items {
		groupPVCID := pvc.Labels["group-pvc-id"]
		accessMode := ""
		if len(pvc.Spec.AccessModes) > 0 {
			accessMode = string(pvc.Spec.AccessModes[0])
		}
		result = append(result, storage.ProjectPVCBindingInfo{
			ID:               fmt.Sprintf("%s/%s", projectNamespace, pvc.Name),
			ProjectID:        projectID,
			GroupPVCID:       groupPVCID,
			ProjectPVCName:   pvc.Name,
			ProjectNamespace: projectNamespace,
			AccessMode:       accessMode,
			Status:           string(pvc.Status.Phase),
			CreatedAt:        pvc.CreationTimestamp.Time,
		})
	}
	return result, nil
}

// DeleteProjectPVCBindingByID deletes a project PVC binding by binding ID
func (pbm *PVCBindingManager) DeleteProjectPVCBindingByID(ctx context.Context, bindingID string) error {
	if k8sclient.Clientset == nil {
		return nil
	}
	namespace, pvcName, err := parseBindingID(bindingID)
	if err != nil {
		return err
	}
	if err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, pvcName, metav1.DeleteOptions{}); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete K8s PVC: %w", err)
	}
	logger.Info("deleted project PVC binding by ID",
		"binding_id", bindingID,
		"namespace", namespace,
		"pvc_name", pvcName)
	return nil
}

// DeleteProjectPVCBinding deletes a project PVC binding
func (pbm *PVCBindingManager) DeleteProjectPVCBinding(ctx context.Context, projectID, userID string, pvcName string) error {
	projectNamespace, err := pbm.resolveProjectNamespace(ctx, projectID, userID)
	if err != nil {
		return err
	}
	if k8sclient.Clientset != nil {
		err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(projectNamespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			logger.Error("failed to delete K8s PVC", "namespace", projectNamespace, "pvc", pvcName, "error", err)
			return err
		}
	}
	logger.Info("deleted project PVC binding",
		"project_id", projectID,
		"pvc_name", pvcName,
		"namespace", projectNamespace)
	return nil
}

func (pbm *PVCBindingManager) resolveProjectNamespace(ctx context.Context, projectID, userID string) (string, error) {
	if pbm.repos == nil || pbm.repos.User == nil {
		safeUser := k8sclient.ToSafeK8sName(userID)
		return k8sclient.FormatNamespaceName(projectID, safeUser), nil
	}
	user, err := pbm.repos.User.GetUserRawByID(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to resolve user for namespace: %w", err)
	}
	safeUsername := k8sclient.ToSafeK8sName(user.Username)
	return k8sclient.FormatNamespaceName(projectID, safeUsername), nil
}

func parseBindingID(bindingID string) (string, string, error) {
	parts := strings.SplitN(bindingID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid binding id format: %s", bindingID)
	}
	return parts[0], parts[1], nil
}

func bindingStatus(binding *corev1.PersistentVolumeClaim) string {
	if binding == nil {
		return "Unknown"
	}
	return string(binding.Status.Phase)
}

// extractGroupIDFromPVCID extracts group ID from PVC ID (format: group-{gid}-{suffix})
func extractGroupIDFromPVCID(pvcID string) (string, error) {
	// Length check: 'group-' (6) + '-' + suffix (10)
	// ID structure: group-{gid}-{suffix}
	if len(pvcID) < 6+10+2 {
		return "", fmt.Errorf("invalid PVC ID format (too short): %s", pvcID)
	}

	// Check prefix
	if pvcID[:6] != "group-" {
		return "", fmt.Errorf("invalid PVC ID format (prefix): %s", pvcID)
	}

	// Check hyphen before suffix
	if pvcID[len(pvcID)-11] != '-' {
		return "", fmt.Errorf("invalid PVC ID format (separator): %s", pvcID)
	}

	// groupID is between 'group-' and '-suffix'
	groupID := pvcID[6 : len(pvcID)-11]

	return groupID, nil
}
