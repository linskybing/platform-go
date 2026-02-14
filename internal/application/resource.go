package application

import (
	"context"
	"errors"

	"github.com/linskybing/platform-go/internal/domain/resource"
	"github.com/linskybing/platform-go/internal/repository"
)

var ErrResourceNotFound = errors.New("resource not found")

type ResourceService struct {
	Repos *repository.Repos
}

func NewResourceService(repos *repository.Repos) *ResourceService {
	return &ResourceService{
		Repos: repos,
	}
}

func (s *ResourceService) ListResourcesByProjectID(projectID string) ([]resource.Resource, error) {
	return s.Repos.Resource.ListResourcesByProjectID(context.Background(), projectID)
}

func (s *ResourceService) ListResourcesByCommitID(commitID string) ([]resource.Resource, error) {
	return s.Repos.Resource.ListResourcesByCommitID(context.Background(), commitID)
}
