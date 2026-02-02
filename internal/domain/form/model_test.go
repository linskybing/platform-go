package form

import (
	"testing"
	"time"
)

// TestFormStatusConstants verifies form status constants
func TestFormStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   FormStatus
		expected string
	}{
		{
			name:     "pending_status",
			status:   FormStatusPending,
			expected: "Pending",
		},
		{
			name:     "processing_status",
			status:   FormStatusProcessing,
			expected: "Processing",
		},
		{
			name:     "completed_status",
			status:   FormStatusCompleted,
			expected: "Completed",
		},
		{
			name:     "rejected_status",
			status:   FormStatusRejected,
			expected: "Rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status != FormStatus(tt.expected) {
				t.Errorf("expected %s, got %s", tt.expected, tt.status)
			}
		})
	}
}

// TestFormStructure verifies Form struct initialization
func TestFormStructure(t *testing.T) {
	tests := []struct {
		name      string
		setupForm func() Form
		verify    func(Form) bool
		scenario  string
	}{
		{
			name: "minimal_form_creation",
			setupForm: func() Form {
				return Form{
					UserID:      1,
					Title:       "Test Form",
					Description: "This is a test form",
					Tag:         "general",
					Status:      FormStatusPending,
				}
			},
			verify: func(f Form) bool {
				return f.UserID == 1 && f.Title == "Test Form" && f.Status == FormStatusPending
			},
			scenario: "Create form with minimum required fields",
		},
		{
			name: "form_with_project_id",
			setupForm: func() Form {
				projectID := uint(42)
				return Form{
					UserID:      2,
					ProjectID:   &projectID,
					Title:       "Project Form",
					Description: "Form linked to project",
					Tag:         "project",
					Status:      FormStatusPending,
				}
			},
			verify: func(f Form) bool {
				return f.ProjectID != nil && *f.ProjectID == 42
			},
			scenario: "Create form with optional project reference",
		},
		{
			name: "form_without_project_id",
			setupForm: func() Form {
				return Form{
					UserID:      3,
					ProjectID:   nil,
					Title:       "Standalone Form",
					Description: "Form without project",
					Tag:         "other",
					Status:      FormStatusPending,
				}
			},
			verify: func(f Form) bool {
				return f.ProjectID == nil && f.UserID == 3
			},
			scenario: "Create standalone form",
		},
		{
			name: "form_status_transitions",
			setupForm: func() Form {
				return Form{
					UserID:      4,
					Title:       "Processing Form",
					Description: "Form being processed",
					Tag:         "urgent",
					Status:      FormStatusProcessing,
				}
			},
			verify: func(f Form) bool {
				return f.Status == FormStatusProcessing
			},
			scenario: "Form in processing status",
		},
		{
			name: "form_with_empty_messages",
			setupForm: func() Form {
				return Form{
					UserID:      5,
					Title:       "Form with Messages",
					Description: "Form that will have messages",
					Messages:    []FormMessage{},
					Status:      FormStatusPending,
				}
			},
			verify: func(f Form) bool {
				return len(f.Messages) == 0
			},
			scenario: "Form initialized with empty messages",
		},
		{
			name: "form_with_long_description",
			setupForm: func() Form {
				longDesc := "This is a very long description that contains multiple lines and detailed information about the form purpose, requirements, and expectations. It demonstrates proper handling of long text fields in the form structure."
				return Form{
					UserID:      6,
					Title:       "Complex Form",
					Description: longDesc,
					Tag:         "detailed",
					Status:      FormStatusPending,
				}
			},
			verify: func(f Form) bool {
				return len(f.Description) > 100
			},
			scenario: "Form with long description text",
		},
		{
			name: "form_rejected_status",
			setupForm: func() Form {
				return Form{
					UserID:      7,
					Title:       "Rejected Form",
					Description: "This form was rejected",
					Tag:         "rejected",
					Status:      FormStatusRejected,
				}
			},
			verify: func(f Form) bool {
				return f.Status == FormStatusRejected
			},
			scenario: "Form with rejected status",
		},
		{
			name: "form_completed_status",
			setupForm: func() Form {
				return Form{
					UserID:      8,
					Title:       "Completed Form",
					Description: "This form is completed",
					Tag:         "done",
					Status:      FormStatusCompleted,
				}
			},
			verify: func(f Form) bool {
				return f.Status == FormStatusCompleted
			},
			scenario: "Form with completed status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := tt.setupForm()
			if !tt.verify(form) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}

// TestFormMessageStructure verifies FormMessage struct
func TestFormMessageStructure(t *testing.T) {
	tests := []struct {
		name      string
		setupMsg  func() FormMessage
		verify    func(FormMessage) bool
		scenario  string
	}{
		{
			name: "message_minimal",
			setupMsg: func() FormMessage {
				return FormMessage{
					FormID:  1,
					UserID:  100,
					Content: "This is a test message",
				}
			},
			verify: func(m FormMessage) bool {
				return m.FormID == 1 && m.UserID == 100 && m.Content != ""
			},
			scenario: "Create minimal form message",
		},
		{
			name: "message_with_id",
			setupMsg: func() FormMessage {
				return FormMessage{
					ID:      99,
					FormID:  2,
					UserID:  101,
					Content: "Message with ID",
				}
			},
			verify: func(m FormMessage) bool {
				return m.ID == 99 && m.FormID == 2
			},
			scenario: "Message with predefined ID",
		},
		{
			name: "message_with_timestamp",
			setupMsg: func() FormMessage {
				now := time.Now()
				return FormMessage{
					ID:        1,
					FormID:    3,
					UserID:    102,
					Content:   "Timestamped message",
					CreatedAt: now,
				}
			},
			verify: func(m FormMessage) bool {
				return !m.CreatedAt.IsZero()
			},
			scenario: "Message with timestamp",
		},
		{
			name: "message_long_content",
			setupMsg: func() FormMessage {
				longContent := "This is a very long message that spans multiple lines and contains detailed information. It tests the ability of the system to handle longer text content in form messages without truncation or errors."
				return FormMessage{
					FormID:  4,
					UserID:  103,
					Content: longContent,
				}
			},
			verify: func(m FormMessage) bool {
				return len(m.Content) > 100
			},
			scenario: "Message with long content",
		},
		{
			name: "message_empty_content_boundary",
			setupMsg: func() FormMessage {
				return FormMessage{
					FormID:  5,
					UserID:  104,
					Content: "",
				}
			},
			verify: func(m FormMessage) bool {
				return m.Content == ""
			},
			scenario: "Message with empty content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := tt.setupMsg()
			if !tt.verify(msg) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}

// TestFormDTOStructure verifies DTO structures
func TestFormDTOStructure(t *testing.T) {
	tests := []struct {
		name         string
		createDTO    *CreateFormDTO
		verifyCreate func(*CreateFormDTO) bool
	}{
		{
			name: "create_form_dto_minimal",
			createDTO: &CreateFormDTO{
				Title:       "Test",
				Description: "Test Description",
			},
			verifyCreate: func(dto *CreateFormDTO) bool {
				return dto.Title == "Test" && dto.Description == "Test Description"
			},
		},
		{
			name: "create_form_dto_with_project",
			createDTO: &CreateFormDTO{
				ProjectID:   ptrUint(10),
				Title:       "Project Form",
				Description: "Form for project",
				Tag:         "project-tag",
			},
			verifyCreate: func(dto *CreateFormDTO) bool {
				return dto.ProjectID != nil && *dto.ProjectID == 10
			},
		},
		{
			name: "create_form_dto_without_project",
			createDTO: &CreateFormDTO{
				Title:       "Standalone",
				Description: "No project",
			},
			verifyCreate: func(dto *CreateFormDTO) bool {
				return dto.ProjectID == nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.verifyCreate(tt.createDTO) {
				t.Error("DTO verification failed")
			}
		})
	}
}

// Helper function for pointer conversion
func ptrUint(u uint) *uint {
	return &u
}
