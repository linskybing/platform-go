package k8s

import (
	"context"
	"fmt"

	"github.com/linskybing/platform-go/internal/domain/storage"
)

// CreateGroupPVC creates a new persistent volume claim for a group.
func (s *K8sService) CreateGroupPVC(ctx context.Context, groupID uint, req *storage.CreateGroupStorageRequest, createdByID uint) (*storage.GroupPVC, error) {
	if groupID == 0 {
		return nil, fmt.Errorf("invalid group ID: %w", ErrInvalidID)
	}
	if req == nil {
		return nil, fmt.Errorf("storage request is required: %w", ErrNilRequest)
	}
	if createdByID == 0 {
		return nil, fmt.Errorf("invalid creator ID: %w", ErrInvalidID)
	}

	pvc, err := s.storageManager.CreateGroupPVC(ctx, groupID, req, createdByID)
	if err != nil {
		return nil, fmt.Errorf("failed to create PVC for group %d: %w", groupID, err)
	}
	return pvc, nil
}

// ListGroupPVCs returns all persistent volume claims for a group.
func (s *K8sService) ListGroupPVCs(ctx context.Context, groupID uint) ([]storage.GroupPVC, error) {
	if groupID == 0 {
		return nil, fmt.Errorf("invalid group ID: %w", ErrInvalidID)
	}

	pvcs, err := s.storageManager.ListGroupPVCs(ctx, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to list PVCs for group %d: %w", groupID, err)
	}
	return pvcs, nil
}

// DeleteGroupPVC deletes a persistent volume claim from a group.
func (s *K8sService) DeleteGroupPVC(ctx context.Context, pvcID string) error {
	if pvcID == "" {
		return fmt.Errorf("PVC ID is required: %w", ErrMissingField)
	}

	if err := s.storageManager.DeleteGroupPVC(ctx, pvcID); err != nil {
		return fmt.Errorf("failed to delete PVC %s: %w", pvcID, err)
	}
	return nil
}

// DeleteGroupStorage deletes all storage resources for a group.
func (s *K8sService) DeleteGroupStorage(ctx context.Context, groupID uint) error {
	if groupID == 0 {
		return fmt.Errorf("invalid group ID: %w", ErrInvalidID)
	}

	// List all PVCs for the group
	pvcs, err := s.ListGroupPVCs(ctx, groupID)
	if err != nil {
		return fmt.Errorf("failed to list PVCs for group %d: %w", groupID, err)
	}

	// Delete each PVC
	for _, pvc := range pvcs {
		if err := s.DeleteGroupPVC(ctx, pvc.ID); err != nil {
			return fmt.Errorf("failed to delete PVC %s for group %d: %w", pvc.ID, groupID, err)
		}
	}

	return nil
}

// ListAllGroupStorages returns all group storage resources (for backward compatibility with handlers).
// Returns interface{} for handler flexibility.
func (s *K8sService) ListAllGroupStorages(ctx context.Context) (interface{}, error) {
	// Note: Full implementation requires listing all group namespaces and aggregating PVCs.
	// Current limitation: Returns error until namespace discovery is implemented.
	// Workaround: Use per-group storage queries with known group IDs.
	return []interface{}{}, fmt.Errorf("group storage listing requires namespace discovery: %w", ErrDeprecated)
}

// CreateGroupPVC creates a new PVC for a group (handler-compatible wrapper).
// This version accepts individual parameters for backward compatibility with handlers.
func (s *K8sService) CreateGroupPVCFromParams(ctx context.Context, groupID uint, groupName string, name string, capacityGi int, storageClass *string) (interface{}, error) {
	if groupID == 0 {
		return nil, fmt.Errorf("invalid group ID: %w", ErrInvalidID)
	}

	req := &storage.CreateGroupStorageRequest{
		GroupID:      groupID,
		GroupName:    groupName,
		Name:         name,
		Capacity:     capacityGi,
		StorageClass: storageClass,
	}

	// Note: Creator ID defaults to 0 (system). Future: Extract from context.Context user claims.
	pvc, err := s.CreateGroupPVC(ctx, groupID, req, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to create PVC for group %d: %w", groupID, err)
	}

	// Return the PVC name for handler compatibility
	if pvc != nil {
		return pvc.Name, nil
	}
	return nil, fmt.Errorf("PVC creation returned nil: %w", ErrMissingField)
}
