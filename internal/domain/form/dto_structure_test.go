package form

import "testing"

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
				ProjectID:   ptrString("10"),
				Title:       "Project Form",
				Description: "Form for project",
				Tag:         "project-tag",
			},
			verifyCreate: func(dto *CreateFormDTO) bool {
				return dto.ProjectID != nil && *dto.ProjectID == "10"
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
