//go:build integration
// +build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/linskybing/platform-go/internal/config"
	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupHandler_Integration(t *testing.T) {
	ctx := GetTestContext()

	var testGroupID string

	t.Run("GetGroups - Success for All Users", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)
		resp, err := client.GET("/groups")

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var groups []group.Group
		err = resp.DecodeJSON(&groups)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(groups), 1)
	})

	t.Run("CreateGroup - Success as Admin", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		createDTO := map[string]string{
			"group_name": "test-integration-group",
		}

		resp, err := client.POSTForm("/groups", createDTO)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var created group.Group
		err = resp.DecodeJSON(&created)
		require.NoError(t, err)
		assert.Equal(t, "test-integration-group", created.GroupName)
		assert.NotEmpty(t, created.GID)
		testGroupID = created.GID
	})

	t.Run("CreateGroup - Forbidden for Regular User", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		createDTO := map[string]string{
			"group_name": "unauthorized-group",
		}

		resp, err := client.POSTForm("/groups", createDTO)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("CreateGroup - Forbidden for Manager", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)

		createDTO := map[string]string{
			"group_name": "manager-group",
		}

		resp, err := client.POSTForm("/groups", createDTO)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("CreateGroup - Empty Name Validation", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		createDTO := map[string]string{
			"group_name": "",
		}

		resp, err := client.POSTForm("/groups", createDTO)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, resp.StatusCode, 400)
	})

	t.Run("CreateGroup - Duplicate Name", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		createDTO := map[string]string{
			"group_name": ctx.TestGroup.GroupName,
		}

		resp, err := client.POSTForm("/groups", createDTO)
		require.NoError(t, err)
		assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode >= 400)
	})

	t.Run("GetGroupByID - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		path := fmt.Sprintf("/groups/%s", ctx.TestGroup.GID)
		resp, err := client.GET(path)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var g group.Group
		err = resp.DecodeJSON(&g)
		require.NoError(t, err)
		assert.Equal(t, ctx.TestGroup.GID, g.GID)
	})

	t.Run("GetGroupByID - Not Found", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		resp, err := client.GET("/groups/99999")
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("UpdateGroup - Success as Admin", func(t *testing.T) {
		if testGroupID == "" {
			t.Skip("No test group created")
		}

		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		updateDTO := map[string]string{
			"group_name": "updated-group-name",
		}

		path := fmt.Sprintf("/groups/%s", testGroupID)
		resp, err := client.PUTForm(path, updateDTO)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Verify update
		getResp, err := client.GET(path)
		require.NoError(t, err)

		var updated group.Group
		err = getResp.DecodeJSON(&updated)
		require.NoError(t, err)
		assert.Equal(t, "updated-group-name", updated.GroupName)
	})

	t.Run("UpdateGroup - Forbidden for Regular User", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		updateDTO := map[string]string{
			"group_name": "hacked",
		}

		path := fmt.Sprintf("/groups/%s", ctx.TestGroup.GID)
		resp, err := client.PUTForm(path, updateDTO)

		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("UpdateGroup - Cannot Update Reserved Group", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		updateDTO := map[string]string{
			"group_name": "new-super-name",
		}

		path := fmt.Sprintf("/groups/%s", ctx.SuperGroup.GID)
		resp, err := client.PUTForm(path, updateDTO)

		require.NoError(t, err)
		// Should fail with forbidden status because super group cannot be renamed
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		// Verify super group name is unchanged
		getResp, err := client.GET(path)
		require.NoError(t, err)

		var g group.Group
		err = getResp.DecodeJSON(&g)
		require.NoError(t, err)
		assert.Equal(t, config.ReservedGroupName, g.GroupName, "Super group name should remain unchanged")
	})

	t.Run("DeleteGroup - Success as Admin", func(t *testing.T) {
		// Create a group to delete
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		createDTO := map[string]string{
			"group_name": "group-to-delete",
		}

		createResp, err := client.POSTForm("/groups", createDTO)
		require.NoError(t, err)

		var created group.Group
		err = createResp.DecodeJSON(&created)
		require.NoError(t, err)

		// Delete it
		path := fmt.Sprintf("/groups/%s", created.GID)
		deleteResp, err := client.DELETE(path)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, deleteResp.StatusCode)

		// Verify deletion
		getResp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})

	t.Run("DeleteGroup - Forbidden for Manager", func(t *testing.T) {
		if testGroupID == "" {
			t.Skip("No test group")
		}

		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)

		path := fmt.Sprintf("/groups/%s", testGroupID)
		resp, err := client.DELETE(path)

		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("DeleteGroup - Cannot Delete Reserved Group", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		path := fmt.Sprintf("/groups/%s", ctx.SuperGroup.GID)
		resp, err := client.DELETE(path)

		require.NoError(t, err)
		// Should fail with appropriate error
		assert.GreaterOrEqual(t, resp.StatusCode, 400)
	})
}
