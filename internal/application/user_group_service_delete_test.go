package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/stretchr/testify/assert"
)

// ---------- DeleteUserGroup ----------
func TestDeleteUserGroup_Success(t *testing.T) {
	svc, ugRepo, userRepo, projectRepo, groupRepo, ctx := setupUserGroupService(t)

	oldUG := group.UserGroup{UID: "1", GID: "2"}
	ugRepo.getUserGroup = func(ctx context.Context, uid, gid string) (*group.UserGroup, error) { return &oldUG, nil }
	groupRepo.getGroupByID = func(ctx context.Context, id string) (*group.Group, error) { return &group.Group{}, nil }
	ugRepo.deleteUserGroup = func(ctx context.Context, uid, gid string) error { return nil }
	userRepo.getUsernameByID = func(ctx context.Context, id string) (string, error) { return "admin", nil }
	projectRepo.listProjectsByGroup = func(ctx context.Context, gid string) ([]project.Project, error) {
		return []project.Project{{PID: "100"}}, nil
	}

	err := svc.DeleteUserGroup(ctx, "1", "2")

	assert.NoError(t, err)
}

func TestDeleteUserGroup_Fail_DeleteRepo(t *testing.T) {
	svc, ugRepo, _, _, groupRepo, ctx := setupUserGroupService(t)

	oldUG := group.UserGroup{UID: "1", GID: "2"}
	ugRepo.getUserGroup = func(ctx context.Context, uid, gid string) (*group.UserGroup, error) { return &oldUG, nil }
	groupRepo.getGroupByID = func(ctx context.Context, id string) (*group.Group, error) { return &group.Group{}, nil }
	ugRepo.deleteUserGroup = func(ctx context.Context, uid, gid string) error { return errors.New("delete fail") }

	err := svc.DeleteUserGroup(ctx, "1", "2")

	assert.Error(t, err)
}
