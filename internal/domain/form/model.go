package form

import (
	"time"

	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/linskybing/platform-go/internal/domain/user"
	gonanoid "github.com/matoous/go-nanoid/v2"
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
	ID          string `json:"id" gorm:"primaryKey;size:21"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt   `gorm:"index"`
	UserID      string           `json:"user_id" gorm:"size:21;index;foreignKey:UserID;references:UID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	ProjectID   *string          `json:"project_id" gorm:"size:21;index;foreignKey:ProjectID;references:PID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"` // Optional
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Tag         string           `json:"tag"` // Single-select tag configured by backend
	Status      FormStatus       `json:"status" gorm:"default:'Pending'"`
	User        *user.User       `json:"user" gorm:"foreignKey:UserID;references:UID"`
	Project     *project.Project `json:"project" gorm:"foreignKey:ProjectID;references:PID"`
	Messages    []FormMessage    `json:"messages" gorm:"foreignKey:FormID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
}

func (m *Form) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID, err = gonanoid.New()
	}
	return
}

// FormMessage represents a comment on a form. Both admin and requester can post.
type FormMessage struct {
	ID        string     `json:"id" gorm:"primaryKey;size:21"`
	FormID    string     `json:"form_id" gorm:"index;size:21;foreignKey:FormID;references:ID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	UserID    string     `json:"user_id" gorm:"size:21;index;foreignKey:UserID;references:UID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Content   string     `json:"content" gorm:"type:text"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`
	Form      *Form      `json:"-" gorm:"foreignKey:FormID;references:ID"`
	User      *user.User `json:"-" gorm:"foreignKey:UserID;references:UID"`
}

func (m *FormMessage) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID, err = gonanoid.New()
	}
	return
}
