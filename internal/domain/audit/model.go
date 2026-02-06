package audit

import (
	"time"

	"github.com/linskybing/platform-go/internal/domain/user"
	"gorm.io/datatypes"
)

type AuditLog struct {
	ID           uint           `gorm:"primaryKey;autoIncrement"`
	UserID       string         `gorm:"not null;index;size:21;foreignKey:UserID;references:UID;constraint:OnDelete:CASCADE,OnUpdate:CASCADE" json:"user_id"`
	Action       string         `gorm:"type:varchar(20);not null;index" json:"action"`
	ResourceType string         `gorm:"type:varchar(50);not null;index" json:"resource_type"`
	ResourceID   string         `gorm:"not null;index" json:"resource_id"`
	OldData      datatypes.JSON `gorm:"type:jsonb" json:"old_data"`
	NewData      datatypes.JSON `gorm:"type:jsonb" json:"new_data"`
	IPAddress    string         `gorm:"type:varchar(45)" json:"ip_address"`
	UserAgent    string         `gorm:"type:text" json:"user_agent"`
	Description  string         `gorm:"type:text" json:"description"`
	CreatedAt    time.Time      `gorm:"autoCreateTime;index" json:"created_at"`
	User         *user.User     `json:"-" gorm:"foreignKey:UserID;references:UID"`
}
