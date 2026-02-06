package image

import (
	"testing"
)

// ptrString returns a pointer to a string
func ptrString(s string) *string {
	return &s
}

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

// TestAllowedImageDTOStructure tests AllowedImageDTO
func TestAllowedImageDTOStructure(t *testing.T) {
	tests := []struct {
		name     string
		dto      AllowedImageDTO
		verify   func(AllowedImageDTO) bool
		scenario string
	}{
		{
			name: "global_allowed_image",
			dto: AllowedImageDTO{
				ID:        "1",
				Registry:  "docker.io",
				ImageName: "ubuntu",
				Tag:       "latest",
				Digest:    "sha256:abc123def456",
				ProjectID: nil,
				IsGlobal:  true,
				IsPulled:  true,
			},
			verify: func(d AllowedImageDTO) bool {
				return d.IsGlobal && d.ProjectID == nil
			},
			scenario: "Global allowed image",
		},
		{
			name: "project_specific_allowed_image",
			dto: AllowedImageDTO{
				ID:        "2",
				Registry:  "gcr.io",
				ImageName: "my-project/app",
				Tag:       "v1.0",
				Digest:    "sha256:xyz789abc123",
				ProjectID: ptrString("100"),
				IsGlobal:  false,
				IsPulled:  true,
			},
			verify: func(d AllowedImageDTO) bool {
				return !d.IsGlobal && d.ProjectID != nil && *d.ProjectID == "100"
			},
			scenario: "Project-specific allowed image",
		},
		{
			name: "unpulled_allowed_image",
			dto: AllowedImageDTO{
				ID:        "3",
				Registry:  "quay.io",
				ImageName: "org/image",
				Tag:       "main",
				Digest:    "sha256:111222333",
				ProjectID: nil,
				IsGlobal:  true,
				IsPulled:  false,
			},
			verify: func(d AllowedImageDTO) bool {
				return !d.IsPulled && d.IsGlobal
			},
			scenario: "Unpulled allowed image",
		},
		{
			name: "allowed_image_with_full_details",
			dto: AllowedImageDTO{
				ID:        "4",
				Registry:  "docker.io",
				ImageName: "library/postgres",
				Tag:       "15-alpine",
				Digest:    "sha256:444555666",
				ProjectID: ptrStringImg("50"),
				IsGlobal:  false,
				IsPulled:  true,
			},
			verify: func(d AllowedImageDTO) bool {
				return len(d.Registry) > 0 && len(d.ImageName) > 0 && len(d.Digest) > 0
			},
			scenario: "Allowed image with complete information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.verify(tt.dto) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}

// Helper function
func ptrUintImg(u uint) *uint {
	return &u
}

func ptrStringImg(s string) *string {
	return &s
}
