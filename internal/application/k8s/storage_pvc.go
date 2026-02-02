package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/storage"
	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateGroupPVC creates a new PVC for a group with performance logging.
func (sm *StorageManager) CreateGroupPVC(ctx context.Context, groupID uint, req *storage.CreateGroupStorageRequest, createdByID uint) (*storage.GroupPVC, error) {
	startTime := time.Now()
	
	if k8sclient.Clientset == nil {
		slog.Warn("k8s client not initialized, using mock PVC",
			"group_id", groupID,
			"pvc_name", req.Name)
		return &storage.GroupPVC{
			ID:      sm.generateGroupPVCID(groupID),
			GroupID: groupID,
			Name:    req.Name,
		}, nil
	}

	pvcID := sm.generateGroupPVCID(groupID)
	ns := sm.getGroupNamespace(groupID)
	pvcName := fmt.Sprintf("pvc-%s", pvcID[6:])

	if err := sm.ensureGroupNamespace(ctx, ns, groupID); err != nil {
		slog.Error("failed to ensure namespace",
			"group_id", groupID,
			"namespace", ns,
			"error", err)
		return nil, fmt.Errorf("failed to ensure namespace: %w", err)
	}

	qty, err := resource.ParseQuantity(fmt.Sprintf("%dGi", req.Capacity))
	if err != nil {
		slog.Error("invalid capacity format",
			"group_id", groupID,
			"capacity_gi", req.Capacity,
			"error", err)
		return nil, fmt.Errorf("invalid capacity: %w", err)
	}

	scName := config.DefaultStorageClassName
	if req.StorageClass != nil && *req.StorageClass != "" {
		scName = *req.StorageClass
	}

	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: ns,
			Labels: map[string]string{
				"app.kubernetes.io/name":       "group-storage",
				"app.kubernetes.io/managed-by": "platform",
				"storage-type":                 "group",
				"group-id":                     fmt.Sprintf("%d", groupID),
				"pvc-uuid":                     pvcID[6:],
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteMany},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: qty,
				},
			},
			StorageClassName: &scName,
		},
	}

	result, err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(ns).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		slog.Error("failed to create k8s PVC",
			"group_id", groupID,
			"pvc_name", pvcName,
			"namespace", ns,
			"error", err)
		return nil, fmt.Errorf("failed to create K8s PVC: %w", err)
	}

	groupPVC := &storage.GroupPVC{
		ID:           pvcID,
		Name:         req.Name,
		GroupID:      groupID,
		Namespace:    ns,
		PVCName:      result.Name,
		Size:         fmt.Sprintf("%dGi", req.Capacity),
		Capacity:     req.Capacity,
		StorageClass: scName,
		AccessMode:   string(corev1.ReadWriteMany),
		Status:       string(result.Status.Phase),
		CreatedBy:    createdByID,
	}

	sm.invalidateCache(groupID)
	
	elapsed := time.Since(startTime)
	slog.Info("PVC created successfully",
		"group_id", groupID,
		"pvc_id", pvcID,
		"pvc_name", result.Name,
		"capacity_gi", req.Capacity,
		"duration_ms", elapsed.Milliseconds(),
		"storage_class", scName)
	
	return groupPVC, nil
}

// ListGroupPVCs retrieves all PVCs for a group with caching and performance metrics.
// Cache hit path: O(1) with ~1-2ms
// Cache miss path: O(n) with ~100-500ms depending on K8s API latency
func (sm *StorageManager) ListGroupPVCs(ctx context.Context, groupID uint) ([]storage.GroupPVC, error) {
	startTime := time.Now()
	
	// Try cache first
	if cached, ok := sm.getCachedPVCs(groupID); ok {
		elapsed := time.Since(startTime)
		slog.Debug("PVC list retrieved from cache",
			"group_id", groupID,
			"count", len(cached),
			"duration_ms", elapsed.Milliseconds())
		return cached, nil
	}

	if k8sclient.Clientset == nil {
		slog.Debug("k8s client not initialized, returning empty PVC list",
			"group_id", groupID)
		return []storage.GroupPVC{}, nil
	}

	ns := sm.getGroupNamespace(groupID)
	listOpts := metav1.ListOptions{LabelSelector: fmt.Sprintf("storage-type=group,group-id=%d", groupID)}

	pvcs, err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(ns).List(ctx, listOpts)
	if err != nil {
		slog.Error("failed to list k8s PVCs",
			"group_id", groupID,
			"namespace", ns,
			"error", err)
		return nil, fmt.Errorf("failed to list PVCs: %w", err)
	}

	var result []storage.GroupPVC
	for _, pvc := range pvcs.Items {
		pvcUUID := pvc.Labels["pvc-uuid"]
		qty := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
		capacityGB := int(qty.ScaledValue(resource.Giga))

		accessMode := ""
		if len(pvc.Spec.AccessModes) > 0 {
			accessMode = string(pvc.Spec.AccessModes[0])
		}

		result = append(result, storage.GroupPVC{
			ID:           fmt.Sprintf("group-%d-%s", groupID, pvcUUID),
			Name:         pvc.Labels["pvc-name"],
			GroupID:      groupID,
			Namespace:    ns,
			PVCName:      pvc.Name,
			Size:         fmt.Sprintf("%dGi", capacityGB),
			Capacity:     capacityGB,
			StorageClass: *pvc.Spec.StorageClassName,
			AccessMode:   accessMode,
			Status:       string(pvc.Status.Phase),
			CreatedAt:    pvc.CreationTimestamp.Time,
		})
	}

	sm.setCachedPVCs(groupID, result, 5*time.Minute)
	
	elapsed := time.Since(startTime)
	slog.Info("PVC list retrieved from k8s API",
		"group_id", groupID,
		"count", len(result),
		"duration_ms", elapsed.Milliseconds(),
		"cache_ttl_seconds", 300)
	
	return result, nil
}