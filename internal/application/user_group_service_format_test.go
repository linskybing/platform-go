package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/linskybing/platform-go/internal/domain/user"
	"github.com/stretchr/testify/assert"
)

// ---------- Formatter ----------
func TestFormatByUID(t *testing.T) {
	svc, _, mockUserRepo, _, mockGroupRepo, _ := setupUserGroupService(t)

	mockUserRepo.getUsernameByID = func(ctx context.Context, id string) (string, error) {
		if id == "1" {
			return "user1", nil
		}
		return "user2", nil
	}

	mockGroupRepo.getGroupByID = func(ctx context.Context, id string) (*group.Group, error) {
		switch id {
		case "10":
			return &group.Group{GID: "10", GroupName: "Group10"}, nil
		case "11":
			return &group.Group{GID: "11", GroupName: "Group11"}, nil
		default:
			return nil, errors.New("not found")
		}
	}

	records := []group.UserGroup{
		{UID: "1", GID: "10", Role: "user"},
		{UID: "1", GID: "11", Role: "admin"},
		{UID: "2", GID: "10", Role: "user"},
	}

	res := svc.FormatByUID(records)

	assert.Len(t, res, 2)
	assert.NotNil(t, res["1"])
	assert.NotNil(t, res["2"])

	userData1 := res["1"]
	assert.Equal(t, "1", userData1["UID"])
	assert.Equal(t, "user1", userData1["UserName"])
	groups1 := userData1["Groups"].([]map[string]interface{})
	assert.Len(t, groups1, 2)

	userData2 := res["2"]
	assert.Equal(t, "2", userData2["UID"])
	assert.Equal(t, "user2", userData2["UserName"])
	groups2 := userData2["Groups"].([]map[string]interface{})
	assert.Len(t, groups2, 1)
}

func TestFormatByGID(t *testing.T) {
	svc, _, _, _, mockGroupRepo, _ := setupUserGroupService(t)

	mockGroupRepo.getGroupByID = func(ctx context.Context, id string) (*group.Group, error) {
		return &group.Group{GID: "10", GroupName: "TestGroup"}, nil
	}

	records := []group.UserGroup{
		{UID: "1", GID: "10", Role: "user", User: user.User{Username: "user1"}},
		{UID: "2", GID: "10", Role: "admin", User: user.User{Username: "user2"}},
	}

	res := svc.FormatByGID(records)

	assert.Len(t, res, 1)
	assert.NotNil(t, res["10"])

	groupData := res["10"]
	assert.Equal(t, "10", groupData["GID"])
	assert.Equal(t, "TestGroup", groupData["GroupName"])
	assert.NotNil(t, groupData["Users"])

	users := groupData["Users"].([]map[string]interface{})
	assert.Len(t, users, 2)
}

func TestFormatByUID_Empty(t *testing.T) {
	svc, _, _, _, _, _ := setupUserGroupService(t)
	res := svc.FormatByUID([]group.UserGroup{})
	assert.Len(t, res, 0)
}

func TestFormatByGID_Empty(t *testing.T) {
	svc, _, _, _, _, _ := setupUserGroupService(t)
	res := svc.FormatByGID([]group.UserGroup{})
	assert.Len(t, res, 0)
}
