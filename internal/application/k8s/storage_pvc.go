package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/storage"
	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const pvcCacheTTL = 5 * time.Minute

// CreateGroupPVC creates a new PVC for a group with performance logging.
func (sm *StorageManager) CreateGroupPVC(ctx context.Context, groupID string, req *storage.CreateGroupStorageRequest, createdByID string) (*storage.GroupPVCSpec, error) {
	startTime := time.Now()

	pvcID, err := sm.generateGroupPVCID(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PVC ID: %w", err)
	}

	ns := sm.getGroupNamespace(groupID)

	// pvcID is group-{gid}-{suffix}
	// We want unique PVC name. replace 'group' with 'pvc'
	pvcName := strings.Replace(pvcID, "group", "pvc", 1)

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
				"group-id":                     groupID,
				"pvc-uuid":                     pvcID,
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

	if k8sclient.Clientset == nil {
		slog.Warn("k8s client not initialized, using mock PVC",
			"group_id", groupID,
			"pvc_name", pvcName)
		if sm.storageRepo != nil {
			dbPVC := &storage.GroupStorage{
				ID:           pvcID,
				Name:         req.Name,
				GroupID:      groupID,
				PVCName:      pvcName,
				Capacity:     req.Capacity,
				StorageClass: scName,
				CreatedBy:    createdByID,
			}
			if err := sm.storageRepo.CreateGroupStorage(ctx, dbPVC); err != nil {
				slog.Error("failed to persist group storage to database",
					"group_id", groupID,
					"pvc_id", pvcID,
					"error", err)
			}
		}
		return &storage.GroupPVCSpec{
			ID:           pvcID,
			GroupID:      groupID,
			Name:         req.Name,
			Namespace:    ns,
			PVCName:      pvcName,
			Capacity:     req.Capacity,
			StorageClass: scName,
			AccessMode:   string(corev1.ReadWriteMany),
			Status:       "Unknown",
			CreatedBy:    createdByID,
			CreatedAt:    time.Now(),
		}, nil
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

	if sm.storageRepo != nil {
		dbPVC := &storage.GroupStorage{
			ID:           pvcID,
			Name:         req.Name,
			GroupID:      groupID,
			PVCName:      result.Name,
			Capacity:     req.Capacity,
			StorageClass: scName,
			CreatedBy:    createdByID,
		}
		if err := sm.storageRepo.CreateGroupStorage(ctx, dbPVC); err != nil {
			slog.Error("failed to persist group storage to database",
				"group_id", groupID,
				"pvc_id", pvcID,
				"error", err)
		}
	}

	groupPVC := &storage.GroupPVCSpec{
		ID:           pvcID,
		Name:         req.Name,
		GroupID:      groupID,
		Namespace:    ns,
		PVCName:      result.Name,
		Capacity:     req.Capacity,
		StorageClass: scName,
		AccessMode:   string(corev1.ReadWriteMany),
		Status:       string(result.Status.Phase),
		CreatedBy:    createdByID,
		CreatedAt:    time.Now(),
	}

	sm.invalidatePVCsCache(ctx, groupID)

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
func (sm *StorageManager) ListGroupPVCs(ctx context.Context, groupID string) ([]storage.GroupPVCSpec, error) {
	var records []storage.GroupStorage
	if cached, ok := sm.getCachedPVCs(groupID); ok {
		records = cached
	} else if sm.storageRepo != nil {
		list, err := sm.storageRepo.ListGroupStorageByGID(ctx, groupID)
		if err != nil {
			return nil, err
		}
		records = list
		sm.setCachedPVCs(groupID, list, pvcCacheTTL)
	}

	ns := sm.getGroupNamespace(groupID)
	result := make([]storage.GroupPVCSpec, 0, len(records))

	for _, rec := range records {
		status := "Unknown"
		accessMode := string(corev1.ReadWriteMany)
		storageClass := rec.StorageClass
		capacity := rec.Capacity

		if k8sclient.Clientset != nil {
			pvc, err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(ns).Get(ctx, rec.PVCName, metav1.GetOptions{})
			if err != nil {
				if apierrors.IsNotFound(err) {
					status = "Missing"
				} else {
					return nil, err
				}
			} else {
				status = string(pvc.Status.Phase)
				if pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName != "" {
					storageClass = *pvc.Spec.StorageClassName
				}
				if len(pvc.Spec.AccessModes) > 0 {
					accessMode = string(pvc.Spec.AccessModes[0])
				}
				if q, ok := pvc.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
					capacity = int(q.Value() / (1024 * 1024 * 1024))
				}
			}
		}

		result = append(result, storage.GroupPVCSpec{
			ID:           rec.ID,
			GroupID:      rec.GroupID,
			Name:         rec.Name,
			Namespace:    ns,
			PVCName:      rec.PVCName,
			Capacity:     capacity,
			StorageClass: storageClass,
			AccessMode:   accessMode,
			Status:       status,
			CreatedAt:    rec.CreatedAt,
			CreatedBy:    rec.CreatedBy,
		})
	}

	return result, nil
}

// DeleteGroupPVC deletes a persistent volume claim from a group.
func (sm *StorageManager) DeleteGroupPVC(ctx context.Context, pvcID string) error {
	var groupID string
	var pvcName string

	if sm.storageRepo != nil {
		rec, err := sm.storageRepo.GetGroupStorage(ctx, pvcID)
		if err == nil && rec != nil {
			groupID = rec.GroupID
			pvcName = rec.PVCName
		}
	}

	if groupID == "" {
		parsedGroupID, err := extractGroupIDFromPVCID(pvcID)
		if err != nil {
			return err
		}
		groupID = parsedGroupID
	}

	if pvcName == "" {
		pvcName = strings.Replace(pvcID, "group", "pvc", 1)
	}

	ns := sm.getGroupNamespace(groupID)

	if k8sclient.Clientset != nil {
		err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(ns).Delete(ctx, pvcName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete K8s PVC: %w", err)
		}
	}

	if sm.storageRepo != nil {
		if err := sm.storageRepo.DeleteGroupStorage(ctx, pvcID); err != nil {
			slog.Error("failed to delete group storage from database",
				"pvc_id", pvcID,
				"error", err)
		}
	}

	sm.invalidatePVCsCache(ctx, groupID)
	return nil
}
