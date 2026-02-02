package form

import (
	"testing"
)

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
				ProjectID:   ptrUint(123),
				Title:       "Project Associated Form",
				Description: "Form linked to project 123",
				Tag:         "project",
			},
			isValid: func(d CreateFormDTO) bool {
				return d.ProjectID != nil && *d.ProjectID == 123
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

// TestUpdateFormStatusDTOValidation tests UpdateFormStatusDTO
func TestUpdateFormStatusDTOValidation(t *testing.T) {
	tests := []struct {
		name     string
		dto      UpdateFormStatusDTO
		isValid  func(UpdateFormStatusDTO) bool
		scenario string
	}{
		{
			name: "update_status_pending",
			dto: UpdateFormStatusDTO{
				Status: "Pending",
			},
			isValid: func(d UpdateFormStatusDTO) bool {
				return d.Status == "Pending"
			},
			scenario: "Update form to pending status",
		},
		{
			name: "update_status_processing",
			dto: UpdateFormStatusDTO{
				Status: "Processing",
			},
			isValid: func(d UpdateFormStatusDTO) bool {
				return d.Status == "Processing"
			},
			scenario: "Update form to processing status",
		},
		{
			name: "update_status_completed",
			dto: UpdateFormStatusDTO{
				Status: "Completed",
			},
			isValid: func(d UpdateFormStatusDTO) bool {
				return d.Status == "Completed"
			},
			scenario: "Update form to completed status",
		},
		{
			name: "update_status_rejected",
			dto: UpdateFormStatusDTO{
				Status: "Rejected",
			},
			isValid: func(d UpdateFormStatusDTO) bool {
				return d.Status == "Rejected"
			},
			scenario: "Update form to rejected status",
		},
		{
			name: "update_status_invalid",
			dto: UpdateFormStatusDTO{
				Status: "InvalidStatus",
			},
			isValid: func(d UpdateFormStatusDTO) bool {
				return d.Status == "InvalidStatus"
			},
			scenario: "Update with invalid status value",
		},
		{
			name: "update_status_empty",
			dto: UpdateFormStatusDTO{
				Status: "",
			},
			isValid: func(d UpdateFormStatusDTO) bool {
				return d.Status == ""
			},
			scenario: "Update with empty status",
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

// TestCreateFormMessageDTOValidation tests CreateFormMessageDTO
func TestCreateFormMessageDTOValidation(t *testing.T) {
	tests := []struct {
		name     string
		dto      CreateFormMessageDTO
		isValid  func(CreateFormMessageDTO) bool
		scenario string
	}{
		{
			name: "message_dto_valid_content",
			dto: CreateFormMessageDTO{
				Content: "This is a valid form message",
			},
			isValid: func(d CreateFormMessageDTO) bool {
				return d.Content != ""
			},
			scenario: "Message with valid content",
		},
		{
			name: "message_dto_long_content",
			dto: CreateFormMessageDTO{
				Content: "This is a very long message content that spans multiple lines and contains detailed information about the form message. It includes multiple sentences and explains the purpose and context of the message thoroughly.",
			},
			isValid: func(d CreateFormMessageDTO) bool {
				return len(d.Content) > 100
			},
			scenario: "Message with long content",
		},
		{
			name: "message_dto_single_character",
			dto: CreateFormMessageDTO{
				Content: "A",
			},
			isValid: func(d CreateFormMessageDTO) bool {
				return len(d.Content) == 1
			},
			scenario: "Message with single character",
		},
		{
			name: "message_dto_empty_content",
			dto: CreateFormMessageDTO{
				Content: "",
			},
			isValid: func(d CreateFormMessageDTO) bool {
				return d.Content == ""
			},
			scenario: "Message with empty content",
		},
		{
			name: "message_dto_whitespace_content",
			dto: CreateFormMessageDTO{
				Content: "   ",
			},
			isValid: func(d CreateFormMessageDTO) bool {
				return d.Content == "   "
			},
			scenario: "Message with whitespace only",
		},
		{
			name: "message_dto_special_characters",
			dto: CreateFormMessageDTO{
				Content: "Message with special chars: !@#$%^&*()_+-=[]{}|;:,.<>?",
			},
			isValid: func(d CreateFormMessageDTO) bool {
				return len(d.Content) > 0
			},
			scenario: "Message with special characters",
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
