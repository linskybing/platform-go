package group

import (
	"time"

	"github.com/linskybing/platform-go/internal/domain/common"
	"github.com/linskybing/platform-go/internal/domain/user"
)

// Group represents a resource-owning collective.
type Group struct {
	ID            string               `gorm:"primaryKey;type:uuid;column:id;default:uuid_generate_v4()"`
	GID           string               `gorm:"-"` // Alias for ID
	Name          string               `gorm:"size:100;not null;column:name"`
	GroupName     string               `gorm:"-"` // Alias for Name
	Description   string               `gorm:"type:text;column:description"`
	ParentGroupID *string              `gorm:"type:uuid;column:parent_group_id"`
	Owner         common.ResourceOwner `gorm:"foreignKey:ID;references:ID"`
	CreatedAt     time.Time            `gorm:"autoCreateTime;column:created_at"`
}

func (Group) TableName() string { return "groups" }

// UserGroup defines the membership relationship.
type UserGroup struct {
	UID       string    `gorm:"primaryKey;type:uuid;column:user_id"`
	GID       string    `gorm:"primaryKey;type:uuid;column:group_id"`
	Role      string    `gorm:"size:50;not null;default:'user';column:role"`
	User      user.User `gorm:"foreignKey:UID"` // Added back for preloading
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

func (UserGroup) TableName() string { return "user_group" }
