package job

import (
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
)

// Job represents a job execution record
type Job struct {
	ID           string     `gorm:"primaryKey;size:21"`
	ConfigFileID string     `gorm:"not null;index;size:21"`
	ProjectID    string     `gorm:"not null;index;size:21"`
	Namespace    string     `gorm:"size:100;not null"`
	UserID       string     `gorm:"not null;index;size:21"`
	Status       string     `gorm:"type:varchar(20);not null;default:'submitted'"`
	SubmitType   string     `gorm:"type:varchar(20);not null;default:'job'"`
	QueueName    string     `gorm:"size:100"`
	Priority     int32      `gorm:"default:0"`
	ErrorMessage string     `gorm:"type:text"`
	SubmittedAt  time.Time  `gorm:"autoCreateTime"`
	StartedAt    *time.Time `gorm:"index"`
	CompletedAt  *time.Time
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}

func (m *Job) BeforeCreate(tx *gorm.DB) (err error) {
	if m.ID == "" {
		m.ID, err = gonanoid.New()
	}
	return
}

// TableName specifies the database table name
func (Job) TableName() string {
	return "jobs"
}
