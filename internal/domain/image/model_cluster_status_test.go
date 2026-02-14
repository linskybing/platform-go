package image

import "testing"

func TestClusterImageStatusStructure(t *testing.T) {
	tests := []struct {
		name     string
		setupCIS func() ClusterImageStatus
		verify   func(ClusterImageStatus) bool
		scenario string
	}{
		{
			name: "unpulled_image",
			setupCIS: func() ClusterImageStatus {
				return ClusterImageStatus{
					TagID:    "4",
					IsPulled: false,
				}
			},
			verify: func(cis ClusterImageStatus) bool {
				return cis.TagID == "4" && !cis.IsPulled
			},
			scenario: "Image not yet pulled",
		},
		{
			name: "successfully_pulled_image",
			setupCIS: func() ClusterImageStatus {
				return ClusterImageStatus{
					TagID:    "101",
					IsPulled: true,
				}
			},
			verify: func(cis ClusterImageStatus) bool {
				return cis.IsPulled
			},
			scenario: "Successfully pulled image",
		},
		{
			name: "failed_pull_attempt",
			setupCIS: func() ClusterImageStatus {
				return ClusterImageStatus{
					TagID:    "102",
					IsPulled: false,
				}
			},
			verify: func(cis ClusterImageStatus) bool {
				return !cis.IsPulled
			},
			scenario: "Failed pull attempt",
		},
		{
			name: "image_with_last_pull_time",
			setupCIS: func() ClusterImageStatus {
				return ClusterImageStatus{
					TagID:    "103",
					IsPulled: true,
				}
			},
			verify: func(cis ClusterImageStatus) bool {
				return cis.IsPulled
			},
			scenario: "Image with pull history",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cis := tt.setupCIS()
			if !tt.verify(cis) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}
