package image

import (
	"time"

	"github.com/google/uuid"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/user"
	"gorm.io/gorm"
)

type ContainerRepository struct {
	ID        string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Registry  string         `gorm:"size:255;default:'docker.io'"`
	Namespace string         `gorm:"size:255"`
	Name      string         `gorm:"size:255;index"`
	FullName  string         `gorm:"uniqueIndex;size:512"`
	Tags      []ContainerTag `gorm:"foreignKey:RepositoryID"`
}

func (m *ContainerRepository) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}

type ContainerTag struct {
	ID           string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	RepositoryID string         `gorm:"index;not null;type:uuid;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Name         string         `gorm:"size:128;index"`
	Digest       string         `gorm:"size:255"`
	Size         int64
	PushedAt     *time.Time
	Repository   *ContainerRepository `json:"-" gorm:"foreignKey:RepositoryID;references:ID"`
}

func (m *ContainerTag) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}

type ImageAllowList struct {
	ID           string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt       `gorm:"index"`
	ProjectID    *string              `gorm:"type:uuid;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	TagID        *string              `gorm:"index;type:uuid;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	RepositoryID string               `gorm:"index;not null;type:uuid;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	RequestID    *string              `gorm:"type:uuid;constraint:OnDelete:SET NULL,OnUpdate:CASCADE"`
	CreatedBy    string               `gorm:"type:uuid;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	IsEnabled    bool                 `gorm:"default:true"`
	Repository   *ContainerRepository `json:"-" gorm:"foreignKey:RepositoryID;references:ID"`
	Tag          *ContainerTag        `json:"-" gorm:"foreignKey:TagID;references:ID"`
	Project      *project.Project     `json:"-" gorm:"foreignKey:ProjectID;references:ID"`
	Request      *ImageRequest        `json:"-" gorm:"foreignKey:RequestID;references:ID"`
	CreatorUser  *user.User           `json:"-" gorm:"foreignKey:CreatedBy;references:ID"`
}

func (m *ImageAllowList) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}

type ImageRequest struct {
	ID             string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	UserID         string         `gorm:"type:uuid;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ProjectID      *string        `gorm:"type:uuid;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	InputRegistry  string
	InputImageName string
	InputTag       string
	Status         string  `gorm:"size:32;default:'pending';index"`
	ReviewerID     *string `gorm:"type:uuid;index;constraint:OnDelete:SET NULL,OnUpdate:CASCADE"`
	ReviewedAt     *time.Time
	ReviewerNote   string           `gorm:"type:text"`
	User           *user.User       `json:"-" gorm:"foreignKey:UserID;references:ID"`
	Project        *project.Project `json:"-" gorm:"foreignKey:ProjectID;references:ID"`
	Reviewer       *user.User       `json:"-" gorm:"foreignKey:ReviewerID;references:ID"`
}

func (m *ImageRequest) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}

type ClusterImageStatus struct {
	ID        string `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	TagID     string         `gorm:"uniqueIndex;type:uuid;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	IsPulled  bool           `gorm:"default:false"`
	Tag       *ContainerTag  `json:"-" gorm:"foreignKey:TagID;references:ID"`
}

func (m *ClusterImageStatus) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}
