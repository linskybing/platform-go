package form

import (
	"time"

	"github.com/google/uuid"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/user"
	"gorm.io/gorm"
)

type FormStatus string

const (
	FormStatusPending    FormStatus = "Pending"
	FormStatusProcessing FormStatus = "Processing"
	FormStatusCompleted  FormStatus = "Completed"
	FormStatusRejected   FormStatus = "Rejected"
)

type Form struct {
	ID          string `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt   `gorm:"index"`
	UserID      string           `json:"user_id" gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ProjectID   *string          `json:"project_id" gorm:"type:uuid;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"` // Optional
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Tag         string           `json:"tag"` // Single-select tag configured by backend
	Status      FormStatus       `json:"status" gorm:"default:'Pending'"`
	User        *user.User       `json:"user" gorm:"foreignKey:UserID;references:ID"`
	Project     *project.Project `json:"project" gorm:"foreignKey:ProjectID;references:ID"`
	Messages    []FormMessage    `json:"messages" gorm:"foreignKey:FormID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

func (m *Form) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}

// FormMessage represents a comment on a form. Both admin and requester can post.
type FormMessage struct {
	ID        string     `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	FormID    string     `json:"form_id" gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	UserID    string     `json:"user_id" gorm:"type:uuid;not null;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Content   string     `json:"content" gorm:"type:text"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	Form      *Form      `json:"-" gorm:"foreignKey:FormID;references:ID"`
	User      *user.User `json:"-" gorm:"foreignKey:UserID;references:ID"`
}

func (m *FormMessage) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = uuid.NewString()
	}
	return nil
}
