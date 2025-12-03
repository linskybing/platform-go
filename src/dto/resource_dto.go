package dto

import "gorm.io/datatypes"

type ResourceUpdateDTO struct {
	Type        *string         `json:"type,omitempty"`
	Name        *string         `json:"name,omitempty"`
	ParsedYAML  *datatypes.JSON `json:"parsed_yaml,omitempty"`
	Description *string         `json:"description,omitempty"`
}
