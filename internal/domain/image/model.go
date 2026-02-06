package image

import (
	"time"

	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/user"
	nanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
)

type ContainerRepository struct {
	ID        string `gorm:"primaryKey;size:21"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	Registry  string         `gorm:"size:255;default:'docker.io'"`
	Namespace string         `gorm:"size:255"`
	Name      string         `gorm:"size:255;index"`
	FullName  string         `gorm:"uniqueIndex;size:512"`
	Tags      []ContainerTag `gorm:"foreignKey:RepositoryID"`
}

func (m *ContainerRepository) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID, err = nanoid.New()
	}
	return
}

type ContainerTag struct {
	ID           string `gorm:"primaryKey;size:21"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
	RepositoryID string         `gorm:"index;not null;size:21;foreignKey:RepositoryID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Name         string         `gorm:"size:128;index"`
	Digest       string         `gorm:"size:255"`
	Size         int64
	PushedAt     *time.Time
	Repository   *ContainerRepository `json:"-" gorm:"foreignKey:RepositoryID;references:ID"`
}

func (m *ContainerTag) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID, err = nanoid.New()
	}
	return
}

type ImageAllowList struct {
	ID           string `gorm:"primaryKey;size:21"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt       `gorm:"index"`
	ProjectID    *string              `gorm:"index;size:21;foreignKey:ProjectID;references:PID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	TagID        *string              `gorm:"index;size:21;foreignKey:TagID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	RepositoryID string               `gorm:"index;not null;size:21;foreignKey:RepositoryID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	RequestID    *string              `gorm:"size:21;foreignKey:RequestID;references:ID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE"`
	CreatedBy    string               `gorm:"size:21;index;foreignKey:CreatedBy;references:UID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	IsEnabled    bool                 `gorm:"default:true"`
	Repository   *ContainerRepository `json:"-" gorm:"foreignKey:RepositoryID;references:ID"`
	Tag          *ContainerTag        `json:"-" gorm:"foreignKey:TagID;references:ID"`
	Project      *project.Project     `json:"-" gorm:"foreignKey:ProjectID;references:PID"`
	Request      *ImageRequest        `json:"-" gorm:"foreignKey:RequestID;references:ID"`
	CreatorUser  *user.User           `json:"-" gorm:"foreignKey:CreatedBy;references:UID"`
}

func (m *ImageAllowList) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID, err = nanoid.New()
	}
	return
}

type ImageRequest struct {
	ID             string `gorm:"primaryKey;size:21"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      gorm.DeletedAt `gorm:"index"`
	UserID         string         `gorm:"index;size:21;foreignKey:UserID;references:UID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ProjectID      *string        `gorm:"index;size:21;foreignKey:ProjectID;references:PID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	InputRegistry  string
	InputImageName string
	InputTag       string
	Status         string  `gorm:"size:32;default:'pending';index"`
	ReviewerID     *string `gorm:"size:21;index;foreignKey:ReviewerID;references:UID;constraint:OnDelete:SET NULL,OnUpdate:CASCADE"`
	ReviewedAt     *time.Time
	ReviewerNote   string           `gorm:"type:text"`
	User           *user.User       `json:"-" gorm:"foreignKey:UserID;references:UID"`
	Project        *project.Project `json:"-" gorm:"foreignKey:ProjectID;references:PID"`
	Reviewer       *user.User       `json:"-" gorm:"foreignKey:ReviewerID;references:UID"`
}

func (m *ImageRequest) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID, err = nanoid.New()
	}
	return
}

type ClusterImageStatus struct {
	ID        string `gorm:"primaryKey;size:21"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	TagID     string         `gorm:"uniqueIndex;size:21;foreignKey:TagID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	IsPulled  bool           `gorm:"default:false"`
	Tag       *ContainerTag  `json:"-" gorm:"foreignKey:TagID;references:ID"`
}

func (m *ClusterImageStatus) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID, err = nanoid.New()
	}
	return
}
