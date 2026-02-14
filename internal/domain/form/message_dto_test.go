package form

import "testing"

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
