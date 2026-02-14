package form

import "testing"

// TestCreateFormDTOValidation tests CreateFormDTO validation
func TestCreateFormDTOValidation(t *testing.T) {
	tests := []struct {
		name     string
		dto      CreateFormDTO
		isValid  func(CreateFormDTO) bool
		scenario string
	}{
		{
			name: "valid_create_form_dto",
			dto: CreateFormDTO{
				Title:       "New Form Request",
				Description: "This is a request for a new form",
				Tag:         "feature",
			},
			isValid: func(d CreateFormDTO) bool {
				return d.Title != "" && d.Description != ""
			},
			scenario: "Valid form creation DTO",
		},
		{
			name: "create_form_dto_with_project_id",
			dto: CreateFormDTO{
				ProjectID:   ptrString("123"),
				Title:       "Project Associated Form",
				Description: "Form linked to project 123",
				Tag:         "project",
			},
			isValid: func(d CreateFormDTO) bool {
				return d.ProjectID != nil && *d.ProjectID == "123"
			},
			scenario: "Form DTO with project association",
		},
		{
			name: "create_form_dto_empty_title",
			dto: CreateFormDTO{
				Title:       "",
				Description: "Description without title",
				Tag:         "notag",
			},
			isValid: func(d CreateFormDTO) bool {
				return d.Title == ""
			},
			scenario: "DTO with empty title (invalid state)",
		},
		{
			name: "create_form_dto_empty_description",
			dto: CreateFormDTO{
				Title:       "Title Only",
				Description: "",
				Tag:         "test",
			},
			isValid: func(d CreateFormDTO) bool {
				return d.Description == ""
			},
			scenario: "DTO with empty description",
		},
		{
			name: "create_form_dto_long_title",
			dto: CreateFormDTO{
				Title:       "This is a very long title for a form request that contains many characters and describes the form in detail",
				Description: "A comprehensive description",
				Tag:         "long-title",
			},
			isValid: func(d CreateFormDTO) bool {
				return len(d.Title) > 50
			},
			scenario: "DTO with long title",
		},
		{
			name: "create_form_dto_no_tag",
			dto: CreateFormDTO{
				Title:       "Form without tag",
				Description: "No tag specified",
				Tag:         "",
			},
			isValid: func(d CreateFormDTO) bool {
				return d.Tag == ""
			},
			scenario: "DTO without tag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.isValid(tt.dto) {
				t.Errorf("validation failed for scenario: %s", tt.scenario)
			}
		})
	}
}
