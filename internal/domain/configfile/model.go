package configfile

import (
	"time"

	"gorm.io/datatypes"
)

// ConfigBlob stores deduplicated configuration content.
type ConfigBlob struct {
	Hash      string         `gorm:"primaryKey;size:64"`
	Content   datatypes.JSON `gorm:"type:jsonb;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
}

func (ConfigBlob) TableName() string { return "config_blobs" }

// ConfigCommit records the version history of configurations.
type ConfigCommit struct {
	ID        string    `gorm:"primaryKey;type:varchar(21)"`
	ProjectID string    `gorm:"type:varchar(21);not null;index"`
	BlobHash  string    `gorm:"size:64;not null"`
	Message   string    `gorm:"not null"`
	AuthorID  string    `gorm:"type:varchar(21);not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (ConfigCommit) TableName() string { return "config_commits" }
