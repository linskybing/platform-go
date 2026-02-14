package common

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// ResourceOwner is the base for all entities that can own resources.
type ResourceOwner struct {
	ID        string    `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	OwnerType string    `gorm:"size:50;not null"` // USER, GROUP
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (ResourceOwner) TableName() string { return "resource_owners" }

// Ltree represents the PostgreSQL ltree type for hierarchical paths.
type Ltree string

func (l Ltree) Value() (driver.Value, error) { return string(l), nil }

func (l *Ltree) Scan(value interface{}) error {
	if value == nil {
		*l = ""
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan Ltree: %v", value)
	}
	*l = Ltree(s)
	return nil
}

// Int4Range represents the PostgreSQL int4range type.
type Int4Range string

func (r Int4Range) Value() (driver.Value, error) { return string(r), nil }

func (r *Int4Range) Scan(value interface{}) error {
	if value == nil {
		*r = ""
		return nil
	}
	s, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan Int4Range: %v", value)
	}
	*r = Int4Range(s)
	return nil
}
