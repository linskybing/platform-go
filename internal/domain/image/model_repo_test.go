package image

import "testing"

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
