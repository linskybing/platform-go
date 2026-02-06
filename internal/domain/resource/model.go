package resource

import (
	"time"

	"github.com/linskybing/platform-go/internal/domain/configfile"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ResourceType defines the type of Kubernetes resource
type ResourceType string

const (
	ResourcePod        ResourceType = "pod"        // Kubernetes Pod
	ResourceService    ResourceType = "service"    // Kubernetes Service
	ResourceDeployment ResourceType = "deployment" // Kubernetes Deployment
	ResourceConfigMap  ResourceType = "configmap"  // Kubernetes ConfigMap
	ResourceIngress    ResourceType = "ingress"    // Kubernetes Ingress
	ResourceJob        ResourceType = "job"        // Kubernetes Job
)

// Resource represents a Kubernetes resource configuration
type Resource struct {
	RID         string                 `gorm:"primaryKey;column:r_id;size:21"`
	CFID        string                 `gorm:"not null;column:cf_id;size:21;index;foreignKey:CFID;references:CFID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"` // ConfigFile ID
	Type        ResourceType           `gorm:"type:resource_type;not null"`
	Name        string                 `gorm:"size:50;not null"`
	ParsedYAML  datatypes.JSON         `gorm:"type:jsonb;not null;"`
	Description *string                `gorm:"type:text"`
	CreatedAt   time.Time              `gorm:"column:create_at;autoCreateTime"`
	ConfigFile  *configfile.ConfigFile `json:"-" gorm:"foreignKey:CFID;references:CFID"`
}

func (m *Resource) BeforeCreate(tx *gorm.DB) (err error) {
	if m.RID == "" {
		m.RID, err = gonanoid.New()
	}
	return
}
