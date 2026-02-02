package resource

import (
	"time"

	"gorm.io/datatypes"
)

type ResourceUpdateDTO struct {
	Type        *string         `json:"type,omitempty"`
	Name        *string         `json:"name,omitempty"`
	ParsedYAML  *datatypes.JSON `json:"parsed_yaml,omitempty"`
	Description *string         `json:"description,omitempty"`
}

type CreatePVCDTO struct {
	Namespace        string `form:"namespace" binding:"required"`
	Name             string `form:"name" binding:"required"`
	StorageClassName string `form:"storageClassName" binding:"required"`
	Size             string `form:"size" binding:"required"`
}

type ExpandPVCDTO struct {
	Namespace string `form:"namespace" binding:"required"`
	Name      string `form:"name" binding:"required"`
	Size      string `form:"size" binding:"required"`
}

type ExpandStorageInput struct {
	NewSize string `json:"new_size" binding:"required" example:"1Ti"`
}

type PVC struct {
	Name      string `json:"name" example:"my-pvc"`
	Namespace string `json:"namespace" example:"default"`
	Status    string `json:"status" example:"Bound"`
	Size      string `json:"size" example:"1Gi"`
	IsGlobal  bool   `json:"isGlobal" example:"false"`
}

// CreateGroupStorageRequest defines the payload for creating group storage.
// 定義建立群組儲存空間的請求參數
type CreateGroupStorageRequest struct {
	GroupID   uint   `json:"groupId" binding:"required"`
	GroupName string `json:"groupName" binding:"required"`
	Capacity    int    `json:"capacity" binding:"required,min=1"` // In Gi
	Name        string `json:"name" binding:"required"`
}

// GroupPVCOutput defines the response structure for listing storages.
type GroupPVCOutput struct {
	ID          string    `json:"id"`          // The Group ID (string format to prevent frontend conversion issues)
	PVCName     string    `json:"pvcName"`     // The K8s PVC Name
	GroupName string    `json:"groupName"` // Human readable name
	Namespace   string    `json:"namespace"`   // K8s Namespace
	Capacity    string    `json:"capacity"`    // e.g., "10Gi"
	Status      string    `json:"status"`      // e.g., "Bound"
	Role        string    `json:"role"`        // [NEW] User's role in the group (admin/manager/member)
	AccessMode  string    `json:"accessmode"`
	CreatedAt   time.Time `json:"createdAt"` // Creation timestamp
}

type StartFileBrowserDTO struct {
	Namespace string `json:"namespace" binding:"required"`
	PVCName   string `json:"pvc_name" binding:"required"`
}

type StopFileBrowserDTO struct {
	Namespace string `json:"namespace" binding:"required"`
	PVCName   string `json:"pvc_name" binding:"required"`
}
