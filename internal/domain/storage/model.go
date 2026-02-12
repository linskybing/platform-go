package storage

import (
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"

	group "github.com/linskybing/platform-go/internal/domain/group"
	user "github.com/linskybing/platform-go/internal/domain/user"
)

// TableName specifies the database table name
func (UserStorage) TableName() string {
	return "user_storage"
}

// BeforeCreate hook to generate a unique ID
func (p *UserStorage) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == "" {
		p.ID, err = gonanoid.New()
	}
	return
}

// UserStorage represents a user-specific persistent volume claim
type UserStorage struct {
	ID           string    `gorm:"primaryKey;column:id;size:21"`
	Name         string    `gorm:"size:100;not null;index"`                     // Display name
	UserID       string    `gorm:"column:user_id;size:21;not null;index"`       // User ID
	PVCName      string    `gorm:"size:100;not null;uniqueIndex:uidx_pvc_user"` // K8s PVC name, unique for user storage
	Capacity     int       `gorm:"not null"`                                    // Capacity in Gi (numeric)
	StorageClass string    `gorm:"size:100;default:'longhorn'"`                 // Storage class name
	CreatedBy    string    `gorm:"not null;size:21"`                            // User ID who created
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;index"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`

	User *user.User `gorm:"foreignKey:UserID;references:UID"` // GORM association to User model
}

// GroupStorage represents a group-specific persistent volume claim
type GroupStorage struct {
	ID           string    `gorm:"primaryKey;column:id;size:21"`
	Name         string    `gorm:"size:100;not null;index"`                      // Display name
	GroupID      string    `gorm:"column:group_id;size:21;not null;index"`       // Group ID
	PVCName      string    `gorm:"size:100;not null;uniqueIndex:uidx_pvc_group"` // K8s PVC name, unique for group storage
	Capacity     int       `gorm:"not null"`                                     // Capacity in Gi (numeric)
	StorageClass string    `gorm:"size:100;default:'longhorn'"`                  // Storage class name
	CreatedBy    string    `gorm:"not null;size:21"`                             // User ID who created
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime;index"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`

	Group *group.Group `gorm:"foreignKey:GroupID;references:GID"` // GORM association to Group model
}

// TableName specifies the database table name for GroupStorage
func (GroupStorage) TableName() string {
	return "group_storage"
}

// BeforeCreate hook to generate a unique ID for GroupStorage
func (g *GroupStorage) BeforeCreate(tx *gorm.DB) (err error) {
	if g.ID == "" {
		g.ID, err = gonanoid.New()
	}
	return
}
