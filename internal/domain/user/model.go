package user

import (
	"time"

	"github.com/linskybing/platform-go/internal/domain/common"
)

// User represents a system user inheriting from ResourceOwner.
type User struct {
	ID           string               `gorm:"primaryKey;type:uuid;references:resource_owners(id)"`
	UID          string               `gorm:"-"` // Legacy Alias
	Username     string               `gorm:"size:100;not null;unique"`
	Email        string               `gorm:"size:255;not null;unique"`
	PasswordHash string               `gorm:"size:255;not null"`
	Password     string               `gorm:"-"`        // Legacy support
	FullName     string               `gorm:"size:100"` // Added back
	Role         string               `gorm:"size:50;default:'USER'"`
	Type         string               `gorm:"size:50;default:'origin'"`  // Added back
	Status       string               `gorm:"size:50;default:'offline'"` // Added back
	Owner        common.ResourceOwner `gorm:"foreignKey:ID"`
	CreatedAt    time.Time            `gorm:"autoCreateTime"`
}

func (User) TableName() string { return "users" }

type UserWithSuperAdmin struct {
	User
	IsSuperAdmin bool `gorm:"column:is_super_admin"`
}
