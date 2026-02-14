package user

import (
	"time"

	"github.com/google/uuid"
	"github.com/linskybing/platform-go/internal/domain/common"
	"gorm.io/gorm"
)

// UserRole defines the possible roles for a user.
type UserRole string

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleManager UserRole = "manager"
	UserRoleUser    UserRole = "user"
)

// UserStatus defines the possible status for a user.
type UserStatus string

const (
	UserStatusOnline  UserStatus = "online"
	UserStatusOffline UserStatus = "offline"
	UserStatusDelete  UserStatus = "delete"
)

// User represents a system user inheriting from ResourceOwner.
type User struct {
	ID           string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4();references:resource_owners(id);constraint:OnDelete:CASCADE"`
	Username     string    `gorm:"size:100;not null;unique"`
	Email        string    `gorm:"size:255;not null;unique"`
	PasswordHash string    `gorm:"size:255;not null"`
	FullName     string    `gorm:"size:100"`
	Role         string    `gorm:"size:50;default:'USER'"`
	IsSuperAdmin bool      `gorm:"default:false;column:is_super_admin"`
	Type         string    `gorm:"size:50;default:'origin'"`
	Status       string    `gorm:"size:50;default:'offline'"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
}

func (User) TableName() string { return "users" }

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.NewString()
	}
	// Ensure base ResourceOwner exists
	owner := common.ResourceOwner{
		ID:        u.ID,
		OwnerType: "USER",
	}
	return tx.Table("resource_owners").Save(&owner).Error
}

type UserWithSuperAdmin struct {
	User
	IsSuperAdmin bool `gorm:"column:is_super_admin"`
}
