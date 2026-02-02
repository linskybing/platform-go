package storage

import "time"

// ===== DTOs for Group Storage =====

// GroupPVCSpec represents a PVC specification for a group
type GroupPVCSpec struct {
	ID           string    `json:"id"`            // Unique identifier: group-{gid}-{uuid}
	GroupID      uint      `json:"group_id"`      // Group ID
	Name         string    `json:"name"`          // User-friendly PVC name
	PVCName      string    `json:"pvc_name"`      // K8s PVC name (for direct deletion)
	Capacity     int       `json:"capacity"`      // Size in Gi
	StorageClass string    `json:"storage_class"` // Storage class name
	Status       string    `json:"status"`        // Bound, Pending, etc
	AccessMode   string    `json:"access_mode"`   // ReadWriteMany, ReadWriteOnce, etc
	CreatedAt    time.Time `json:"created_at"`    // Creation timestamp
	CreatedBy    uint      `json:"created_by"`    // User ID who created this PVC
}

// CreateGroupStorageRequest defines the payload for creating group storage.
// Supports both simple (ID, name, capacity) and extended (storage class) formats.
type CreateGroupStorageRequest struct {
	GroupID      uint    `json:"group_id" binding:"required"`
	GroupName    string  `json:"group_name" binding:"required"`     // Group display name
	Name         string  `json:"name" binding:"required"`           // PVC display name
	Capacity     int     `json:"capacity" binding:"required,min=1"` // Size in Gi
	StorageClass *string `json:"storage_class,omitempty"`           // Optional storage class
}

// ListGroupStorageResponse represents the response for listing group storage
type ListGroupStorageResponse struct {
	GroupID uint           `json:"group_id"`
	PVCs    []GroupPVCSpec `json:"pvcs"`
	Total   int            `json:"total"`
}

// ExpandGroupStorageRequest defines the payload for expanding group storage
type ExpandGroupStorageRequest struct {
	PVCSpecID   string `json:"pvc_spec_id" binding:"required"`        // The unique PVC ID
	NewCapacity int    `json:"new_capacity" binding:"required,min=1"` // New size in Gi
}

// DeleteGroupStorageRequest defines the payload for deleting group storage
type DeleteGroupStorageRequest struct {
	PVCName string `json:"pvc_name" binding:"required"` // K8s PVC name for direct deletion
}

// GroupStorageInfo represents group storage information
type GroupStorageInfo struct {
	ID         uint      `json:"id"`
	GroupID    uint      `json:"group_id"`
	GroupName  string    `json:"group_name"`
	Namespace  string    `json:"namespace"`
	PVCName    string    `json:"pvc_name"`
	Capacity   string    `json:"capacity"`
	Size       string    `json:"size,omitempty"`
	Status     string    `json:"status"`
	AccessMode string    `json:"access_mode"`
	Role       string    `json:"role,omitempty"` // admin/manager/member
	CreatedAt  time.Time `json:"created_at"`
}

// ===== DTOs for User Storage (Legacy) =====

// ExpandStorageRequest represents a request to expand storage
type ExpandStorageRequest struct {
	StorageName  string  `json:"storage_name" binding:"required"`
	NewSize      string  `json:"new_size" binding:"required"`
	StorageClass *string `json:"storage_class,omitempty"`
}

// UserStorageInfo represents user storage information
type UserStorageInfo struct {
	Username  string    `json:"username"`
	Namespace string    `json:"namespace"`
	PVCName   string    `json:"pvc_name"`
	Size      string    `json:"size"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// GroupPVCOutput represents the output for group PVC information
type GroupPVCOutput struct {
	GroupID   uint      `json:"group_id"`
	Name      string    `json:"name"`
	Capacity  string    `json:"capacity"`
	Status    string    `json:"status"`
	Role      string    `json:"role,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ExpandStorageInput represents the input for expanding storage
type ExpandStorageInput struct {
	StorageName string `json:"storage_name" binding:"required"`
	NewSize     string `json:"new_size" binding:"required"`
}
