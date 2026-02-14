package resource

import (
	"time"

	"gorm.io/datatypes"
)

type ResourceType string

const (
	ResourcePod        ResourceType = "Pod"
	ResourceService    ResourceType = "Service"
	ResourceDeployment ResourceType = "Deployment"
	ResourceConfigMap  ResourceType = "ConfigMap"
	ResourceIngress    ResourceType = "Ingress"
	ResourceJob        ResourceType = "Job"
)

// Resource represents a K8s resource (Pod, Service, etc.) managed by the platform.
type Resource struct {
	ID             string         `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ConfigCommitID string         `gorm:"column:config_commit_id;type:uuid;not null;index"`
	Type           ResourceType   `gorm:"size:50;not null"` // Pod, Service, etc.
	Name           string         `gorm:"size:100;not null"`
	ParsedYAML     datatypes.JSON `gorm:"type:jsonb;not null"`
	Description    string         `gorm:"type:text"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
}

func (Resource) TableName() string { return "resources" }
