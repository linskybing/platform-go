// src/dto/storage_dto.go
package dto

import "time"

// CreateProjectStorageRequest defines the payload for creating project storage.
// 定義建立專案儲存空間的請求參數
type CreateProjectStorageRequest struct {
	ProjectID   uint   `json:"projectId" binding:"required"`
	ProjectName string `json:"projectName" binding:"required"`
	Capacity    int    `json:"capacity" binding:"required,min=1"` // In Gi
}

// ProjectPVCOutput defines the response structure for listing storages.
// 定義回傳給前端的 PVC 列表資料結構
type ProjectPVCOutput struct {
	ID          string    `json:"id"`          // The Project ID
	PVCName     string    `json:"pvcName"`     // The K8s PVC Name
	ProjectName string    `json:"projectName"` // Human readable name
	Namespace   string    `json:"namespace"`
	Capacity    string    `json:"capacity"`
	Status      string    `json:"status"`
	AccessMode  string    `json:"accessMode"`
	CreatedAt   time.Time `json:"createdAt"`
}
