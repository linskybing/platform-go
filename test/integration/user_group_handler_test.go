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
	cleaner.RegisterUser(testUser.ID)

	testGroup := generator.GenerateGroup("usergroup-group")
	require.NoError(t, generator.CreateTestGroup(testGroup))
	cleaner.RegisterGroup(testGroup.ID)

	t.Run("AddUserToGroup - Success as Admin", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		payload := map[string]string{
			"uid":  testUser.ID,
			"gid":  testGroup.ID,
			"role": string(user.UserRoleUser),
		}

		resp, err := client.POST("/user-groups", payload)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("AddUserToGroup - Duplicate Entry", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		payload := map[string]string{
			"uid":  testUser.ID,
			"gid":  testGroup.ID,
			"role": string(user.UserRoleUser),
		}

		resp, err := client.POST("/user-groups", payload)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetUserGroupsByUID - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)
		path := fmt.Sprintf("/user-groups/by-user?u_id=%s", testUser.ID)

		resp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetUserGroupsByGID - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)
		path := fmt.Sprintf("/user-groups/by-group?g_id=%s", testGroup.ID)

		resp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("UpdateUserRole - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := "/user-groups"

		payload := map[string]string{
			"uid":  testUser.ID,
			"gid":  testGroup.ID,
			"role": string(user.UserRoleManager),
		}

		resp, err := client.PUT(path, payload)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("UpdateUserRole - Invalid Role", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := "/user-groups"

		payload := map[string]string{
			"uid":  testUser.ID,
			"gid":  testGroup.ID,
			"role": "invalid_role",
		}

		resp, err := client.PUT(path, payload)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetGroupMembers - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)
		path := fmt.Sprintf("/user-groups/%s/members", testGroup.ID)

		resp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("RemoveUserFromGroup - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		// Use query params for DELETE as per handler implementation
		path := fmt.Sprintf("/user-groups?u_id=%s&g_id=%s", testUser.ID, testGroup.ID)

		resp, err := client.DELETE(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("RemoveUserFromGroup - Already Removed", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := fmt.Sprintf("/user-groups?u_id=%s&g_id=%s", testUser.ID, testGroup.ID)

		resp, err := client.DELETE(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
