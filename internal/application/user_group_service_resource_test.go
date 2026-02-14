package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/stretchr/testify/assert"
)

// ---------- AllocateGroupResource ----------
func TestAllocateGroupResource_Success(t *testing.T) {
	svc, _, _, projectRepo, _, _ := setupUserGroupService(t)

	projectRepo.listProjectsByGroup = func(ctx context.Context, gid string) ([]project.Project, error) {
		return []project.Project{{PID: "100"}}, nil
	}

	err := svc.AllocateGroupResource("1", "admin")

	assert.NoError(t, err)
}

func TestAllocateGroupResource_Fail_ListProjects(t *testing.T) {
	svc, _, _, projectRepo, _, _ := setupUserGroupService(t)

	projectRepo.listProjectsByGroup = func(ctx context.Context, gid string) ([]project.Project, error) {
		return nil, errors.New("db fail")
	}

	err := svc.AllocateGroupResource("1", "admin")

	assert.Error(t, err)
}

// ---------- RemoveGroupResource ----------
func TestRemoveGroupResource_Success(t *testing.T) {
	svc, _, _, projectRepo, _, _ := setupUserGroupService(t)

	projectRepo.listProjectsByGroup = func(ctx context.Context, gid string) ([]project.Project, error) {
		return []project.Project{{PID: "100"}}, nil
	}

	err := svc.RemoveGroupResource("1", "admin")

	assert.NoError(t, err)
}
