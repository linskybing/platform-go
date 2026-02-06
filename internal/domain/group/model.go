package group

import (
	"time"

	"github.com/linskybing/platform-go/internal/domain/user"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
)

type Group struct {
	GID         string    `gorm:"primaryKey;column:g_id;size:20"`
	GroupName   string    `gorm:"size:100;not null"`
	Description string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"column:create_at"`
	UpdatedAt   time.Time `gorm:"column:update_at"`
}

func (g *Group) BeforeCreate(tx *gorm.DB) (err error) {
	if g.GID == "" {
		g.GID, err = gonanoid.New(12)
	}
	return
}

func (Group) TableName() string {
	return "group_list"
}

type UserGroup struct {
	UID       string    `gorm:"primaryKey;column:u_id;size:20;foreignKey:UID;references:UID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	GID       string    `gorm:"primaryKey;column:g_id;size:20;foreignKey:GID;references:GID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	Role      string    `gorm:"type:user_role;default:user;not null"` // ENUM
	CreatedAt time.Time `gorm:"column:create_at"`
	UpdatedAt time.Time `gorm:"column:update_at"`
	// Relationships
	User  *user.User `json:"-" gorm:"foreignKey:UID;references:UID"`
	Group *Group     `json:"-" gorm:"foreignKey:GID;references:GID"`
}

func (UserGroup) TableName() string {
	return "user_group"
}
