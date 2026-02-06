//go:build integration
// +build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linskybing/platform-go/internal/domain/user"
)

func TestUserGroupHandler_Integration(t *testing.T) {
	ctx := GetTestContext()
	generator := NewTestDataGenerator()
	cleaner := NewDatabaseCleaner()
	t.Cleanup(func() {
		_ = cleaner.Cleanup()
	})

	testUser := generator.GenerateUser("usergroup-test")
	require.NoError(t, generator.CreateTestUser(testUser))
	cleaner.RegisterUser(testUser.UID)

	testGroup := generator.GenerateGroup("usergroup-group")
	require.NoError(t, generator.CreateTestGroup(testGroup))
	cleaner.RegisterGroup(testGroup.GID)

	t.Run("AddUserToGroup - Success as Admin", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		formData := map[string]string{
			"user_id":  testUser.UID,
			"group_id": testGroup.GID,
			"role":     string(user.UserRoleUser),
		}

		resp, err := client.POSTForm("/user-groups", formData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("AddUserToGroup - Duplicate Entry", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		formData := map[string]string{
			"user_id":  testUser.UID,
			"group_id": testGroup.GID,
			"role":     string(user.UserRoleUser),
		}

		resp, err := client.POSTForm("/user-groups", formData)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetUserGroupsByUID - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)
		path := fmt.Sprintf("/users/%s/groups", testUser.UID)

		resp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetUserGroupsByGID - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)
		path := fmt.Sprintf("/groups/%s/members", testGroup.GID)

		resp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("UpdateUserRole - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := fmt.Sprintf("/user-groups/%s/%s", testUser.UID, testGroup.GID)

		formData := map[string]string{
			"role": string(user.UserRoleManager),
		}

		resp, err := client.PUTForm(path, formData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("UpdateUserRole - Invalid Role", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := fmt.Sprintf("/user-groups/%s/%s", testUser.UID, testGroup.GID)

		formData := map[string]string{
			"role": "invalid_role",
		}

		resp, err := client.PUTForm(path, formData)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetGroupMembers - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)
		path := fmt.Sprintf("/groups/%s/users", testGroup.GID)

		resp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("RemoveUserFromGroup - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := fmt.Sprintf("/user-groups/%s/%s", testUser.UID, testGroup.GID)

		resp, err := client.DELETE(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	})

	t.Run("RemoveUserFromGroup - Already Removed", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := fmt.Sprintf("/user-groups/%s/%s", testUser.UID, testGroup.GID)

		resp, err := client.DELETE(path)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusNoContent, resp.StatusCode)
	})
}
