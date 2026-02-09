package configfile

import (
	"time"

	"github.com/linskybing/platform-go/internal/domain/project"
	nanoid "github.com/matoous/go-nanoid/v2"
	"gorm.io/gorm"
)

type ConfigFile struct {
	CFID      string           `gorm:"primaryKey;column:cf_id;size:21"`
	Filename  string           `gorm:"size:200;not null"`
	Content   string           `gorm:"size:10000"`
	ProjectID string           `gorm:"not null;size:21;index;constraint:OnDelete:CASCADE,OnUpdate:CASCADE"`
	CreatedAt time.Time        `gorm:"column:create_at"`
	UpdatedAt time.Time        `gorm:"column:update_at;autoUpdateTime"`
	Project   *project.Project `json:"-" gorm:"foreignKey:ProjectID;references:PID"`
}

func (cf *ConfigFile) BeforeCreate(tx *gorm.DB) (err error) {
	if cf.CFID == "" {
		cf.CFID, err = nanoid.New()
	}
	return
}
