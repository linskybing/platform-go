package k8s

import (
	"context"
	"fmt"
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
		pm:    &PermissionManager{},
		repos: repos,
	}
}

// CreateProjectPVCBinding creates a PVC in user's project namespace that binds to group storage
// This allows users to mount group storage in their workloads
func (pbm *PVCBindingManager) CreateProjectPVCBinding(ctx context.Context, req *storage.CreateProjectPVCBindingRequest, userID uint) (*storage.ProjectPVCBindingInfo, error) {
	startTime := time.Now()

	// Extract group ID from PVC ID (format: group-{gid}-{uuid})
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
		return nil, fmt.Errorf("permission denied: you don't have access to this storage")
	}

	// Get source group PVC to find the PV name
	groupPVCs, err := pbm.sm.ListGroupPVCs(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to list group PVCs: %w", err)
	}

	var sourcePVC *storage.GroupPVC
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

	// Determine project namespace
	projectNamespace := fmt.Sprintf("project-%d", req.ProjectID)

	// Ensure project namespace exists
	if err := pbm.ensureProjectNamespace(ctx, projectNamespace, req.ProjectID); err != nil {
		return nil, fmt.Errorf("failed to ensure project namespace: %w", err)
	}

	// Determine access mode based on permission
	accessMode := corev1.ReadOnlyMany
	if perm.CanWrite() && !req.ReadOnly {
		accessMode = corev1.ReadWriteMany
	}

	// Create PVC in project namespace that binds to the same PV
	binding, err := pbm.createBindingPVC(ctx, projectNamespace, req.PVCName, pvName, sourcePVC.Size, accessMode)
	if err != nil {
		return nil, fmt.Errorf("failed to create binding PVC: %w", err)
	}

	// Store binding in database
	projectBinding := &storage.ProjectPVCBinding{
		ProjectID:        req.ProjectID,
		UserID:           userID,
		GroupPVCID:       req.GroupPVCID,
		ProjectPVCName:   req.PVCName,
		ProjectNamespace: projectNamespace,
		SourcePVName:     pvName,
		AccessMode:       string(accessMode),
		Status:           "Bound",
		CreatedAt:        time.Now(),
	}

	// Persist PVC binding to database
	if err := pbm.repos.ProjectPVCBinding.CreateBinding(ctx, projectBinding); err != nil {
		// Rollback K8s PVC creation
		_ = k8sclient.Clientset.CoreV1().PersistentVolumeClaims(projectNamespace).Delete(ctx, req.PVCName, metav1.DeleteOptions{})
		return nil, fmt.Errorf("failed to store binding: %w", err)
	}

	logger.Info("created project PVC binding",
		"project_id", req.ProjectID,
		"user_id", userID,
		"group_pvc_id", req.GroupPVCID,
		"project_pvc", req.PVCName,
		"access_mode", accessMode,
		"duration_ms", time.Since(startTime).Milliseconds())

	return &storage.ProjectPVCBindingInfo{
		ID:               projectBinding.ID,
		ProjectID:        req.ProjectID,
		GroupPVCID:       req.GroupPVCID,
		ProjectPVCName:   req.PVCName,
		ProjectNamespace: projectNamespace,
		AccessMode:       string(accessMode),
		Status:           string(binding.Status.Phase),
		CreatedAt:        time.Now(),
	}, nil
}

// createBindingPVC creates a PVC that binds to an existing PV
func (pbm *PVCBindingManager) createBindingPVC(ctx context.Context, namespace, pvcName, pvName, size string, accessMode corev1.PersistentVolumeAccessMode) (*corev1.PersistentVolumeClaim, error) {
	if k8sclient.Clientset == nil {
		logger.Info("k8s client not initialized, using mock binding PVC")
		return &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{Name: pvcName, Namespace: namespace},
			Status:     corev1.PersistentVolumeClaimStatus{Phase: corev1.ClaimBound},
		}, nil
	}

	qty, err := resource.ParseQuantity(size)
	if err != nil {
		return nil, fmt.Errorf("invalid size: %w", err)
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "project-storage-binding",
				"app.kubernetes.io/managed-by": "platform",
				"storage-type":                 "group-binding",
				"source-pv":                    pvName,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{accessMode},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: qty,
				},
			},
			VolumeName: pvName, // Bind to specific PV
		},
	}

	result, err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			return k8sclient.Clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
		}
		return nil, err
	}

	logger.Info("created binding PVC in project namespace",
		"namespace", namespace,
		"pvc_name", pvcName,
		"pv_name", pvName,
		"access_mode", accessMode)

	return result, nil
}

// getPVNameFromPVC retrieves the PV name from a bound PVC
func (pbm *PVCBindingManager) getPVNameFromPVC(ctx context.Context, namespace, pvcName string) (string, error) {
	if k8sclient.Clientset == nil {
		return "mock-pv-name", nil
	}

	pvc, err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, pvcName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get PVC: %w", err)
	}

	if pvc.Spec.VolumeName == "" {
		return "", fmt.Errorf("PVC not bound to any PV yet")
	}

	return pvc.Spec.VolumeName, nil
}

// ensureProjectNamespace ensures the project namespace exists
func (pbm *PVCBindingManager) ensureProjectNamespace(ctx context.Context, namespace string, projectID uint) error {
	if k8sclient.Clientset == nil {
		return nil
	}

	_, err := k8sclient.Clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err == nil {
		return nil // Already exists
	}

	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "platform",
				"project-id":                   fmt.Sprintf("%d", projectID),
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

// DeleteProjectPVCBinding deletes a project PVC binding
func (pbm *PVCBindingManager) DeleteProjectPVCBinding(ctx context.Context, projectID uint, pvcName string) error {
	projectNamespace := fmt.Sprintf("project-%d", projectID)

	// Get binding info from database
	binding, err := pbm.repos.ProjectPVCBinding.GetBindingByProjectPVC(ctx, projectNamespace, pvcName)
	if err != nil {
		logger.Warn("binding not found in database", "namespace", projectNamespace, "pvc_name", pvcName)
	}

	// Delete K8s PVC
	if k8sclient.Clientset != nil {
		err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(projectNamespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			logger.Error("failed to delete K8s PVC", "namespace", projectNamespace, "pvc", pvcName, "error", err)
		}
	}

	// Delete binding record from database
	if binding != nil {
		if err := pbm.repos.ProjectPVCBinding.DeleteBinding(ctx, binding.ID); err != nil {
			logger.Error("failed to delete binding record", "binding_id", binding.ID, "error", err)
		}
	}

	logger.Info("deleted project PVC binding",
		"project_id", projectID,
		"pvc_name", pvcName)

	return nil
}

// extractGroupIDFromPVCID extracts group ID from PVC ID (format: group-{gid}-{uuid})
func extractGroupIDFromPVCID(pvcID string) (uint, error) {
	var groupID uint
	_, err := fmt.Sscanf(pvcID, "group-%d-", &groupID)
	if err != nil {
		return 0, fmt.Errorf("invalid PVC ID format: %s", pvcID)
	}
	return groupID, nil
}
