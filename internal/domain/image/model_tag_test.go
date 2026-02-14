package image

import (
	"testing"
	"time"
)

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
					Size:         1073741824,
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
