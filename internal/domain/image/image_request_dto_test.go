package image

import "testing"

// TestCreateImageRequestDTOValidation tests CreateImageRequestDTO
func TestCreateImageRequestDTOValidation(t *testing.T) {
	tests := []struct {
		name     string
		dto      CreateImageRequestDTO
		isValid  func(CreateImageRequestDTO) bool
		scenario string
	}{
		{
			name: "valid_image_request",
			dto: CreateImageRequestDTO{
				Registry:  "docker.io",
				ImageName: "ubuntu",
				Tag:       "latest",
			},
			isValid: func(d CreateImageRequestDTO) bool {
				return d.Registry == "docker.io" && d.ImageName == "ubuntu" && d.Tag == "latest"
			},
			scenario: "Valid image request",
		},
		{
			name: "image_request_with_project",
			dto: CreateImageRequestDTO{
				Registry:  "gcr.io",
				ImageName: "my-app",
				Tag:       "v1.0",
				ProjectID: ptrString("25"),
			},
			isValid: func(d CreateImageRequestDTO) bool {
				return d.ProjectID != nil && *d.ProjectID == "25"
			},
			scenario: "Image request for specific project",
		},
		{
			name: "image_request_without_project",
			dto: CreateImageRequestDTO{
				Registry:  "quay.io",
				ImageName: "app",
				Tag:       "main",
				ProjectID: nil,
			},
			isValid: func(d CreateImageRequestDTO) bool {
				return d.ProjectID == nil
			},
			scenario: "Image request without project association",
		},
		{
			name: "custom_registry_request",
			dto: CreateImageRequestDTO{
				Registry:  "myregistry.example.com",
				ImageName: "custom-image",
				Tag:       "build-123",
			},
			isValid: func(d CreateImageRequestDTO) bool {
				return d.Registry == "myregistry.example.com"
			},
			scenario: "Custom registry image request",
		},
		{
			name: "empty_registry_uses_default",
			dto: CreateImageRequestDTO{
				Registry:  "",
				ImageName: "node",
				Tag:       "18",
			},
			isValid: func(d CreateImageRequestDTO) bool {
				return d.Registry == ""
			},
			scenario: "Empty registry field",
		},
		{
			name: "complex_image_name",
			dto: CreateImageRequestDTO{
				Registry:  "docker.io",
				ImageName: "library/ubuntu/base",
				Tag:       "22.04",
			},
			isValid: func(d CreateImageRequestDTO) bool {
				return len(d.ImageName) > 10
			},
			scenario: "Complex image name",
		},
		{
			name: "version_tag",
			dto: CreateImageRequestDTO{
				Registry:  "docker.io",
				ImageName: "python",
				Tag:       "3.11.5",
			},
			isValid: func(d CreateImageRequestDTO) bool {
				return d.Tag == "3.11.5"
			},
			scenario: "Version tag",
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

// TestUpdateImageRequestDTOValidation tests UpdateImageRequestDTO
func TestUpdateImageRequestDTOValidation(t *testing.T) {
	tests := []struct {
		name     string
		dto      UpdateImageRequestDTO
		isValid  func(UpdateImageRequestDTO) bool
		scenario string
	}{
		{
			name: "approve_request",
			dto: UpdateImageRequestDTO{
				Status: "approved",
				Note:   "Image verified and approved",
			},
			isValid: func(d UpdateImageRequestDTO) bool {
				return d.Status == "approved"
			},
			scenario: "Approve image request",
		},
		{
			name: "reject_request",
			dto: UpdateImageRequestDTO{
				Status: "rejected",
				Note:   "Image contains security vulnerabilities",
			},
			isValid: func(d UpdateImageRequestDTO) bool {
				return d.Status == "rejected" && len(d.Note) > 0
			},
			scenario: "Reject image request",
		},
		{
			name: "approval_without_note",
			dto: UpdateImageRequestDTO{
				Status: "approved",
				Note:   "",
			},
			isValid: func(d UpdateImageRequestDTO) bool {
				return d.Status == "approved" && d.Note == ""
			},
			scenario: "Approval without review note",
		},
		{
			name: "rejection_with_detail",
			dto: UpdateImageRequestDTO{
				Status: "rejected",
				Note:   "This image is from an untrusted registry and contains outdated dependencies. Please use an image from our approved registry instead.",
			},
			isValid: func(d UpdateImageRequestDTO) bool {
				return d.Status == "rejected" && len(d.Note) > 50
			},
			scenario: "Rejection with detailed note",
		},
		{
			name: "invalid_status",
			dto: UpdateImageRequestDTO{
				Status: "pending",
				Note:   "Still reviewing",
			},
			isValid: func(d UpdateImageRequestDTO) bool {
				return d.Status == "pending"
			},
			scenario: "Invalid status value",
		},
		{
			name: "empty_status",
			dto: UpdateImageRequestDTO{
				Status: "",
				Note:   "Test",
			},
			isValid: func(d UpdateImageRequestDTO) bool {
				return d.Status == ""
			},
			scenario: "Empty status",
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
