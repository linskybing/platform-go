package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/stretchr/testify/assert"
)

// ---------- CreateUserGroup ----------
func TestCreateUserGroup_Success(t *testing.T) {
	svc, ugRepo, userRepo, projectRepo, _, ctx := setupUserGroupService(t)

	ug := &group.UserGroup{UID: "1", GID: "2"}
	projects := []project.Project{{PID: "100", ProjectName: "p1"}}

	ugRepo.createUserGroup = func(ctx context.Context, ug *group.UserGroup) error { return nil }
	userRepo.getUsernameByID = func(ctx context.Context, id string) (string, error) { return "admin", nil }
	projectRepo.listProjectsByGroup = func(ctx context.Context, gid string) ([]project.Project, error) { return projects, nil }

	res, err := svc.CreateUserGroup(ctx, ug)

	assert.NoError(t, err)
	assert.Equal(t, ug, res)
}

func TestCreateUserGroup_Fail_CreateRepo(t *testing.T) {
	svc, ugRepo, _, _, _, ctx := setupUserGroupService(t)

	ug := &group.UserGroup{UID: "1", GID: "2"}
	ugRepo.createUserGroup = func(ctx context.Context, ug *group.UserGroup) error { return errors.New("db error") }

	res, err := svc.CreateUserGroup(ctx, ug)

	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestCreateUserGroup_Fail_GetUser(t *testing.T) {
	svc, ugRepo, userRepo, _, _, ctx := setupUserGroupService(t)

	ug := &group.UserGroup{UID: "1", GID: "2"}
	ugRepo.createUserGroup = func(ctx context.Context, ug *group.UserGroup) error { return nil }
	userRepo.getUsernameByID = func(ctx context.Context, id string) (string, error) { return "", errors.New("user not found") }

	res, err := svc.CreateUserGroup(ctx, ug)

	assert.Error(t, err)
	assert.Nil(t, res)
}
