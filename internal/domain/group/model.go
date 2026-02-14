package group

import (
	"time"

	"github.com/google/uuid"
	"github.com/linskybing/platform-go/internal/domain/common"
	"github.com/linskybing/platform-go/internal/domain/user"
	"gorm.io/gorm"
)

// Group represents a resource-owning collective.
type Group struct {
	ID            string    `gorm:"primaryKey;type:uuid;column:id;default:uuid_generate_v4();references:resource_owners(id);constraint:OnDelete:CASCADE"`
	Name          string    `gorm:"size:100;not null;column:name"`
	Description   string    `gorm:"type:text;column:description"`
	ParentGroupID *string   `gorm:"type:uuid;column:parent_group_id"`
	CreatedAt     time.Time `gorm:"autoCreateTime;column:created_at"`
}

func (Group) TableName() string { return "groups" }

func (g *Group) BeforeCreate(tx *gorm.DB) error {
	if g.ID == "" {
		g.ID = uuid.NewString()
	}
	owner := common.ResourceOwner{
		ID:        g.ID,
		OwnerType: "GROUP",
	}
	return tx.Table("resource_owners").Save(&owner).Error
}

// UserGroup defines the membership relationship.
type UserGroup struct {
	UserID    string    `gorm:"primaryKey;type:uuid;column:user_id"`
	GroupID   string    `gorm:"primaryKey;type:uuid;column:group_id"`
	Role      string    `gorm:"size:50;not null;default:'user';column:role"`
	User      user.User `gorm:"foreignKey:UserID"`  // Added back for preloading
	Group     Group     `gorm:"foreignKey:GroupID"` // Added for preloading
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

func (UserGroup) TableName() string { return "user_group" }
