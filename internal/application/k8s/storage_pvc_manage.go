package k8s

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	k8sclient "github.com/linskybing/platform-go/pkg/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ExpandGroupPVC expands a group's PVC by PVC name directly.
// This approach avoids expensive database lookups.
// Performance: O(1) instead of O(n) list operation
func (sm *StorageManager) ExpandGroupPVC(ctx context.Context, pvcName string, newCapacity int) error {
	if k8sclient.Clientset == nil {
		slog.Debug("k8s client not initialized, skipping PVC expansion")
		return nil
	}

	// Extract groupID from PVC name format: group-{gid}-{namespace}-pvc
	groupID, namespace, err := parseGroupPVCName(pvcName)
	if err != nil {
		slog.Error("failed to parse PVC name", "pvc_name", pvcName, "error", err)
		return fmt.Errorf("invalid PVC name format: %w", err)
	}

	newSize := fmt.Sprintf("%dGi", newCapacity)

	if err := k8sclient.ExpandPVC(namespace, pvcName, newSize); err != nil {
		slog.Error("failed to expand PVC",
			"pvc_name", pvcName,
			"namespace", namespace,
			"new_capacity_gi", newCapacity,
			"error", err)
		return fmt.Errorf("failed to expand PVC %s: %w", pvcName, err)
	}

	// Invalidate cache for this group
	sm.invalidatePVCsCache(ctx, groupID)

	slog.Info("PVC expanded successfully",
		"pvc_name", pvcName,
		"new_capacity_gi", newCapacity,
		"group_id", groupID)

	return nil
}

// DeleteGroupPVC deletes a PVC from a group by PVC name directly.
// This approach avoids expensive database lookups and unnecessary iterations.
// Performance: O(1) instead of O(n) list operation
func (sm *StorageManager) DeleteGroupPVC(ctx context.Context, pvcName string) error {
	if k8sclient.Clientset == nil {
		slog.Debug("k8s client not initialized, skipping PVC deletion")
		return nil
	}

	// Extract groupID and namespace from PVC name
	groupID, namespace, err := parseGroupPVCName(pvcName)
	if err != nil {
		slog.Error("failed to parse PVC name", "pvc_name", pvcName, "error", err)
		return fmt.Errorf("invalid PVC name format: %w", err)
	}

	// Delete PVC directly from K8s
	if err := k8sclient.Clientset.CoreV1().
		PersistentVolumeClaims(namespace).
		Delete(ctx, pvcName, metav1.DeleteOptions{}); err != nil {
		slog.Error("failed to delete PVC from k8s",
			"pvc_name", pvcName,
			"namespace", namespace,
			"error", err)
		return fmt.Errorf("failed to delete PVC %s: %w", pvcName, err)
	}

	// Invalidate cache for this group
	sm.invalidatePVCsCache(ctx, groupID)

	slog.Info("PVC deleted successfully",
		"pvc_name", pvcName,
		"namespace", namespace,
		"group_id", groupID)

	return nil
}

// parseGroupPVCName extracts groupID and namespace from PVC name.
// Expected format: group-{gid}-{namespace}-pvc
func parseGroupPVCName(pvcName string) (groupID uint, namespace string, err error) {
	parts := strings.Split(pvcName, "-")
	if len(parts) < 4 || parts[0] != "group" {
		return 0, "", fmt.Errorf("invalid PVC name format, expected: group-{gid}-{namespace}-pvc")
	}

	// Parse group ID
	if _, err := fmt.Sscanf(parts[1], "%d", &groupID); err != nil {
		return 0, "", fmt.Errorf("invalid group ID in PVC name: %w", err)
	}

	// Reconstruct namespace (handles namespaces with hyphens)
	// Join all parts except first, second, and last
	namespaceParts := parts[2 : len(parts)-1]
	namespace = strings.Join(namespaceParts, "-")

	if namespace == "" {
		return 0, "", fmt.Errorf("unable to extract namespace from PVC name")
	}

	return groupID, namespace, nil
}
