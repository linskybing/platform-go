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
func (sm *StorageManager) CreateGroupPVC(ctx context.Context, groupID string, req *storage.CreateGroupStorageRequest, createdByID string) (*storage.GroupPVC, error) {
	startTime := time.Now()

	pvcID, err := sm.generateGroupPVCID(groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PVC ID: %w", err)
	}

	if k8sclient.Clientset == nil {
		slog.Warn("k8s client not initialized, using mock PVC",
			"group_id", groupID,
			"pvc_name", req.Name)
		return &storage.GroupPVC{
			ID:      pvcID,
			GroupID: groupID,
			Name:    req.Name,
		}, nil
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
func (sm *StorageManager) ListGroupPVCs(ctx context.Context, groupID string) ([]storage.GroupPVC, error) {
	// Try cache first
	if cached, ok := sm.getCachedPVCs(groupID); ok {
		return cached, nil
	}

	cacheKey := sm.pvcCacheKey(groupID)
	if sm.cache != nil && sm.cache.Enabled() {
		var cached []storage.GroupPVC
		if err := sm.cache.GetJSON(ctx, cacheKey, &cached); err == nil {
			sm.setCachedPVCs(groupID, cached, pvcCacheTTL)
			return cached, nil
		}
	}

	if k8sclient.Clientset == nil {
		return []storage.GroupPVC{}, nil
	}

	ns := sm.getGroupNamespace(groupID)
	pvcList, err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("group-id=%s", groupID),
	})

	if err != nil {
		// Namespace might not exist, return empty
		return []storage.GroupPVC{}, nil
	}

	var pvcs []storage.GroupPVC
	for _, p := range pvcList.Items {
		pvcID := p.Labels["pvc-uuid"]
		if pvcID == "" {
			pvcID = p.Name
		}

		capacity := 0
		if q, ok := p.Spec.Resources.Requests[corev1.ResourceStorage]; ok {
			capacity = int(q.Value() / (1024 * 1024 * 1024))
		}

		sc := ""
		if p.Spec.StorageClassName != nil {
			sc = *p.Spec.StorageClassName
		}

		am := ""
		if len(p.Spec.AccessModes) > 0 {
			am = string(p.Spec.AccessModes[0])
		}

		pvcs = append(pvcs, storage.GroupPVC{
			ID:           pvcID,
			Name:         p.Name, // Using K8s name as display name fallback
			GroupID:      groupID,
			Namespace:    p.Namespace,
			PVCName:      p.Name,
			Size:         fmt.Sprintf("%dGi", capacity),
			Capacity:     capacity,
			StorageClass: sc,
			AccessMode:   am,
			Status:       string(p.Status.Phase),
		})
	}

	sm.setCachedPVCs(groupID, pvcs, pvcCacheTTL)
	return pvcs, nil
}

// DeleteGroupPVC deletes a persistent volume claim from a group.
func (sm *StorageManager) DeleteGroupPVC(ctx context.Context, pvcID string) error {
	// Extract group ID from PVC ID: group-{gid}-{suffix}
	parts := strings.Split(pvcID, "-")
	if len(parts) < 3 || parts[0] != "group" {
		return fmt.Errorf("invalid PVC ID format: %s", pvcID)
	}
	// Since gid is NanoID (maybe containing - ?) No, NanoID alphabet I used is lowercase+numbers.
	// But standard NanoID has - and _. I used custom alphabet without - inside `generateGroupPVCID` (storage_manager).
	// So splitting by - is safe for `gid` IF `gid` itself doesn't contain `-`.
	// WARNING: GroupID (from Group model) uses standard NanoID `gonanoid.New()` probably.
	// `internal/domain/group/model.go` uses `nanoid.New()`. `nanoid` contains `_` and `-`.
	// So `parts := strings.Split(id, "-")` is DANGEROUS if GroupID has hyphen.

	// If GroupID has hyphen, `group-{gid}-{suffix}` parsing fails using Split.
	// ID structure is `group` `gid` `suffix`.
	// Suffix is fixed length (10).
	// So we can parse from right?
	// `suffix` = last 10 chars.
	// `gid` = substring from index 6 to len-11.

	if len(pvcID) < 6+10+2 { // group- + - + suffix
		return fmt.Errorf("invalid PVC ID format (too short): %s", pvcID)
	}

	// check hyphen before suffix at position len-11
	if pvcID[len(pvcID)-11] != '-' {
		return fmt.Errorf("invalid PVC ID format (separator): %s", pvcID)
	}

	// groupID is between 'group-' and '-suffix'
	groupID := pvcID[6 : len(pvcID)-11]

	ns := sm.getGroupNamespace(groupID)

	// pvcName is pvc-{suffix} ? No, earlier I did: strings.Replace(pvcID, "group", "pvc", 1)
	// so pvcName = pvc-{gid}-{suffix}.
	pvcName := strings.Replace(pvcID, "group", "pvc", 1)

	if k8sclient.Clientset != nil {
		err := k8sclient.Clientset.CoreV1().PersistentVolumeClaims(ns).Delete(ctx, pvcName, metav1.DeleteOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to delete K8s PVC: %w", err)
		}
	}

	sm.invalidatePVCsCache(ctx, groupID)
	return nil
}
