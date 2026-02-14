package job

import (
	"time"
)

// Job represents a job execution record in the high-concurrency queue.
type Job struct {
	ID              string     `gorm:"primaryKey;type:uuid"`
	ConfigCommitID  string     `gorm:"type:varchar(21);not null;index"`
	ProjectID       string     `gorm:"type:uuid;not null;index"`
	UserID          string     `gorm:"type:uuid;not null;index"`
	Status          string     `gorm:"size:50;not null;default:'PENDING'"` // PENDING, RUNNING, COMPLETED, FAILED, PREEMPTED
	SubmitType      string     `gorm:"size:20"`                            // Added back for executor support
	QueueName       string     `gorm:"size:50"`                            // Added back for executor support
	PriorityClassID *string    `gorm:"type:uuid;index"`                    // Pointer because it can be null (default priority)
	PriorityValue   int        `gorm:"not null;default:0;index"`           // Denormalized priority value
	RequiredGPU     int        `gorm:"default:0"`
	AssignedNode    string     `gorm:"size:100"`
	Namespace       string     `gorm:"size:100;not null"`
	ErrorMessage    string     `gorm:"type:text"`
	CreatedAt       time.Time  `gorm:"autoCreateTime"`
	StartedAt       *time.Time `gorm:"index"`
	CompletedAt     *time.Time
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

func (Job) TableName() string { return "jobs" }
