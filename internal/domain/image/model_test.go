package image

import (
	"testing"
	"time"
)

// TestContainerRepositoryStructure verifies ContainerRepository model
func TestContainerRepositoryStructure(t *testing.T) {
	tests := []struct {
		name      string
		setupRepo func() ContainerRepository
		verify    func(ContainerRepository) bool
		scenario  string
	}{
		{
			name: "minimal_repository",
			setupRepo: func() ContainerRepository {
				return ContainerRepository{
					Registry:  "docker.io",
					Namespace: "library",
					Name:      "ubuntu",
					FullName:  "docker.io/library/ubuntu",
				}
			},
			verify: func(r ContainerRepository) bool {
				return r.Registry == "docker.io" &&
					r.Name == "ubuntu" &&
					r.FullName == "docker.io/library/ubuntu"
			},
			scenario: "Create container repository",
		},
		{
			name: "custom_registry",
			setupRepo: func() ContainerRepository {
				return ContainerRepository{
					Registry:  "gcr.io",
					Namespace: "my-project",
					Name:      "my-image",
					FullName:  "gcr.io/my-project/my-image",
				}
			},
			verify: func(r ContainerRepository) bool {
				return r.Registry == "gcr.io" && r.Namespace == "my-project"
			},
			scenario: "Custom registry repository",
		},
		{
			name: "with_tags",
			setupRepo: func() ContainerRepository {
				return ContainerRepository{
					Registry:  "docker.io",
					Namespace: "library",
					Name:      "python",
					FullName:  "docker.io/library/python",
					Tags: []ContainerTag{
						{
							Name:   "latest",
							Digest: "sha256:abcd1234",
						},
						{
							Name:   "3.11",
							Digest: "sha256:efgh5678",
						},
					},
				}
			},
			verify: func(r ContainerRepository) bool {
				return len(r.Tags) == 2
			},
			scenario: "Repository with tags",
		},
		{
			name: "empty_tags",
			setupRepo: func() ContainerRepository {
				return ContainerRepository{
					Registry:  "quay.io",
					Namespace: "org",
					Name:      "app",
					FullName:  "quay.io/org/app",
					Tags:      []ContainerTag{},
				}
			},
			verify: func(r ContainerRepository) bool {
				return len(r.Tags) == 0
			},
			scenario: "Repository with empty tags",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.setupRepo()
			if !tt.verify(repo) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}

// TestContainerTagStructure verifies ContainerTag model
func TestContainerTagStructure(t *testing.T) {
	tests := []struct {
		name     string
		setupTag func() ContainerTag
		verify   func(ContainerTag) bool
		scenario string
	}{
		{
			name: "minimal_tag",
			setupTag: func() ContainerTag {
				return ContainerTag{
					RepositoryID: "1",
					Name:         "latest",
					Digest:       "sha256:1234567890abcdef",
				}
			},
			verify: func(tag ContainerTag) bool {
				return tag.RepositoryID == "1" && tag.Name == "latest"
			},
			scenario: "Create container tag",
		},
		{
			name: "tag_with_size",
			setupTag: func() ContainerTag {
				return ContainerTag{
					RepositoryID: "2",
					Name:         "v1.0",
					Digest:       "sha256:abcdef1234567890",
					Size:         1073741824, // 1GB
				}
			},
			verify: func(tag ContainerTag) bool {
				return tag.Size == 1073741824
			},
			scenario: "Tag with size information",
		},
		{
			name: "tag_with_push_time",
			setupTag: func() ContainerTag {
				pushedAt := time.Now().Add(-24 * time.Hour)
				return ContainerTag{
					RepositoryID: "3",
					Name:         "stable",
					Digest:       "sha256:xyz789abc",
					PushedAt:     &pushedAt,
				}
			},
			verify: func(tag ContainerTag) bool {
				return tag.PushedAt != nil
			},
			scenario: "Tag with push timestamp",
		},
		{
			name: "tag_without_push_time",
			setupTag: func() ContainerTag {
				return ContainerTag{
					RepositoryID: "4",
					Name:         "dev",
					Digest:       "sha256:qwe123rty",
					PushedAt:     nil,
				}
			},
			verify: func(tag ContainerTag) bool {
				return tag.PushedAt == nil
			},
			scenario: "Tag without push timestamp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := tt.setupTag()
			if !tt.verify(tag) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}

// TestImageAllowListStructure verifies ImageAllowList model
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

// TestImageRequestStructure verifies ImageRequest model
func TestImageRequestStructure(t *testing.T) {
	tests := []struct {
		name     string
		setupIR  func() ImageRequest
		verify   func(ImageRequest) bool
		scenario string
	}{
		{
			name: "pending_request",
			setupIR: func() ImageRequest {
				return ImageRequest{
					UserID:         "1",
					ProjectID:      ptrString("50"),
					InputRegistry:  "docker.io",
					InputImageName: "ubuntu",
					InputTag:       "latest",
					Status:         "pending",
				}
			},
			verify: func(ir ImageRequest) bool {
				return ir.Status == "pending" && ir.InputRegistry == "docker.io"
			},
			scenario: "Create pending image request",
		},
		{
			name: "approved_request",
			setupIR: func() ImageRequest {
				reviewedAt := time.Now()
				return ImageRequest{
					UserID:         "2",
					ProjectID:      nil,
					InputRegistry:  "gcr.io",
					InputImageName: "my-app",
					InputTag:       "v1.0",
					Status:         "approved",
					ReviewerID:     ptrString("100"),
					ReviewedAt:     &reviewedAt,
					ReviewerNote:   "Approved for production",
				}
			},
			verify: func(ir ImageRequest) bool {
				return ir.Status == "approved" && ir.ReviewerID != nil
			},
			scenario: "Approved image request with review",
		},
		{
			name: "rejected_request",
			setupIR: func() ImageRequest {
				reviewedAt := time.Now()
				return ImageRequest{
					UserID:         "3",
					ProjectID:      ptrString("51"),
					InputRegistry:  "quay.io",
					InputImageName: "untrusted-image",
					InputTag:       "v0.1",
					Status:         "rejected",
					ReviewerID:     ptrString("101"),
					ReviewedAt:     &reviewedAt,
					ReviewerNote:   "Security concerns detected",
				}
			},
			verify: func(ir ImageRequest) bool {
				return ir.Status == "rejected" && len(ir.ReviewerNote) > 0
			},
			scenario: "Rejected image request",
		},
		{
			name: "unreviewed_request",
			setupIR: func() ImageRequest {
				return ImageRequest{
					UserID:         "4",
					ProjectID:      nil,
					InputRegistry:  "docker.io",
					InputImageName: "postgres",
					InputTag:       "15",
					Status:         "pending",
					ReviewerID:     nil,
					ReviewedAt:     nil,
				}
			},
			verify: func(ir ImageRequest) bool {
				return ir.ReviewerID == nil && ir.ReviewedAt == nil
			},
			scenario: "Unreviewed request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupIR()
			if !tt.verify(req) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}

// TestClusterImageStatusStructure verifies ClusterImageStatus model
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

// Helper functions
func ptrUint(u uint) *uint {
	return &u
}
