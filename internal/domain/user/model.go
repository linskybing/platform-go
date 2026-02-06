package user

import (
	"time"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
)

type UserStatus string

const (
	UserStatusOnline  UserStatus = "online"
	UserStatusOffline UserStatus = "offline"
	UserStatusDelete  UserStatus = "delete"
)

type UserType string

const (
	UserTypeOrigin UserType = "origin"
	UserTypeOauth2 UserType = "oauth2"
)

type UserRole string

const (
	UserRoleAdmin   UserRole = "admin"
	UserRoleManager UserRole = "manager"
	UserRoleUser    UserRole = "user"
)

type User struct {
	UID       string    `gorm:"primaryKey;column:u_id;size:20"`
	Username  string    `gorm:"size:50;not null;unique" json:"Username"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Email     *string   `gorm:"size:100"`
	FullName  *string   `gorm:"size:50"`
	Type      string    `gorm:"type:user_type;default:'origin';not null"`
	Status    string    `gorm:"type:user_status;default:'offline';not null"`
	CreatedAt time.Time `gorm:"column:create_at;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:update_at;autoUpdateTime"`
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.UID == "" {
		u.UID, err = gonanoid.New(12)
		if err != nil {
			return err
		}
	}
	return
}

type UserWithSuperAdmin struct {
	UID          string    `gorm:"column:u_id" json:"UID"`
	Username     string    `gorm:"column:username" json:"Username"`
	Password     string    `gorm:"column:password" json:"-"`
	Email        string    `gorm:"column:email" json:"Email"`
	FullName     string    `gorm:"column:full_name" json:"FullName"`
	Type         string    `gorm:"column:type" json:"Type"`
	Status       string    `gorm:"column:status" json:"Status"`
	CreatedAt    time.Time `gorm:"column:create_at" json:"CreatedAt"`
	UpdatedAt    time.Time `gorm:"column:update_at" json:"UpdatedAt"`
	IsSuperAdmin bool      `gorm:"column:is_super_admin" json:"IsSuperAdmin"`
}

// TableName specifies the table name for UserWithSuperAdmin
func (UserWithSuperAdmin) TableName() string {
	return "users"
}
