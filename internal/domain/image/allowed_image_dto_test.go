package image

import "testing"

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
				ProjectID: ptrString("50"),
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
