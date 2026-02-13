package user

import (
	"time"

	"gorm.io/gorm"
)

// UserSettings stores per-user preferences.
type UserSettings struct {
	UserID               string    `gorm:"primaryKey;size:20" json:"user_id"`
	Theme                string    `gorm:"size:10;default:'light'" json:"theme"`
	Language             string    `gorm:"size:10;default:'en'" json:"language"`
	ReceiveNotifications bool      `gorm:"default:true" json:"receiveNotifications"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (UserSettings) TableName() string {
	return "user_settings"
}

// DefaultSettings returns default settings for a new user.
func DefaultSettings(userID string) *UserSettings {
	return &UserSettings{
		UserID:               userID,
		Theme:                "light",
		Language:             "en",
		ReceiveNotifications: true,
	}
}

// BeforeCreate ensures defaults are set.
func (s *UserSettings) BeforeCreate(tx *gorm.DB) error {
	if s.Theme == "" {
		s.Theme = "light"
	}
	if s.Language == "" {
		s.Language = "en"
	}
	return nil
}
