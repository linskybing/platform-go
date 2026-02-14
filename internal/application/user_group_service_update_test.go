package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/stretchr/testify/assert"
)

// ---------- UpdateUserGroup ----------
func TestUpdateUserGroup_Success(t *testing.T) {
	svc, ugRepo, _, _, groupRepo, ctx := setupUserGroupService(t)

	oldUG := group.UserGroup{UID: "1", GID: "1", Role: "user"}
	newUG := &group.UserGroup{UID: "1", GID: "1", Role: "admin"}

	groupRepo.getGroupByID = func(ctx context.Context, id string) (*group.Group, error) { return &group.Group{}, nil }
	ugRepo.updateUserGroup = func(ctx context.Context, ug *group.UserGroup) error { return nil }

	res, err := svc.UpdateUserGroup(ctx, newUG, oldUG)
	assert.NoError(t, err)
	assert.Equal(t, newUG, res)
}

func TestUpdateUserGroup_Fail_UpdateRepo(t *testing.T) {
	svc, ugRepo, _, _, groupRepo, ctx := setupUserGroupService(t)

	oldUG := group.UserGroup{UID: "1", GID: "1"}
	newUG := &group.UserGroup{UID: "1", GID: "1"}

	groupRepo.getGroupByID = func(ctx context.Context, id string) (*group.Group, error) { return &group.Group{}, nil }
	ugRepo.updateUserGroup = func(ctx context.Context, ug *group.UserGroup) error { return errors.New("update fail") }

	res, err := svc.UpdateUserGroup(ctx, newUG, oldUG)

	assert.Nil(t, res)
	assert.EqualError(t, err, "update fail")
}
