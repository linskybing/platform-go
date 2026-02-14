package image

import "testing"

func TestImageAllowListStructure(t *testing.T) {
	tests := []struct {
		name     string
		setupAL  func() ImageAllowList
		verify   func(ImageAllowList) bool
		scenario string
	}{
		{
			name: "global_allow_rule",
			setupAL: func() ImageAllowList {
				return ImageAllowList{
					ProjectID:    nil,
					TagID:        ptrString("10"),
					RepositoryID: "5",
					IsEnabled:    true,
				}
			},
			verify: func(al ImageAllowList) bool {
				return al.ProjectID == nil && al.IsEnabled
			},
			scenario: "Global allow list rule",
		},
		{
			name: "project_specific_rule",
			setupAL: func() ImageAllowList {
				projectID := "100"
				return ImageAllowList{
					ProjectID:    &projectID,
					TagID:        ptrString("20"),
					RepositoryID: "6",
					IsEnabled:    true,
				}
			},
			verify: func(al ImageAllowList) bool {
				return al.ProjectID != nil && *al.ProjectID == "100"
			},
			scenario: "Project-specific allow rule",
		},
		{
			name: "disabled_rule",
			setupAL: func() ImageAllowList {
				return ImageAllowList{
					ProjectID:    nil,
					TagID:        ptrString("30"),
					RepositoryID: "7",
					IsEnabled:    false,
				}
			},
			verify: func(al ImageAllowList) bool {
				return !al.IsEnabled
			},
			scenario: "Disabled allow rule",
		},
		{
			name: "rule_with_creator",
			setupAL: func() ImageAllowList {
				return ImageAllowList{
					ProjectID:    nil,
					TagID:        ptrString("40"),
					RepositoryID: "8",
					CreatedBy:    "999",
					IsEnabled:    true,
				}
			},
			verify: func(al ImageAllowList) bool {
				return al.CreatedBy == "999"
			},
			scenario: "Rule with creator information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			al := tt.setupAL()
			if !tt.verify(al) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}
