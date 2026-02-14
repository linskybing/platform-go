//go:build integration
// +build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoragePermissionHandler_Integration(t *testing.T) {
	ctx := GetTestContext()
	pvcID := fmt.Sprintf("group-%s-testpvc", ctx.TestGroup.ID)
	generator := NewTestDataGenerator()
	cleaner := NewDatabaseCleaner()
	t.Cleanup(func() {
		_ = cleaner.Cleanup()
	})

	testUser := generator.GenerateUser("storage-perm-test")
	require.NoError(t, generator.CreateTestUser(testUser))
	cleaner.RegisterUser(testUser.ID)

	testProject := generator.GenerateProject("storage-project", ctx.TestGroup.ID)
	require.NoError(t, generator.CreateTestProject(testProject))
	cleaner.RegisterProject(testProject.ID)

	t.Run("SetPermission - Success as Admin", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		body := map[string]string{
			"user_id":    testUser.ID,
			"group_id":   ctx.TestGroup.ID,
			"pvc_id":     pvcID,
			"permission": "write",
		}

		resp, err := client.POST("/storage/permissions", body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("SetPermission - Invalid Permission Type", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		body := map[string]string{
			"user_id":    testUser.ID,
			"group_id":   ctx.TestGroup.ID,
			"pvc_id":     pvcID,
			"permission": "invalid",
		}

		resp, err := client.POST("/storage/permissions", body)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetUserPermission - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)
		path := fmt.Sprintf("/storage/permissions/group/%s/pvc/%s", ctx.TestGroup.ID, pvcID)

		resp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("BatchSetPermissions - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		testUser2 := generator.GenerateUser("storage-test-2")
		require.NoError(t, generator.CreateTestUser(testUser2))
		cleaner.RegisterUser(testUser2.ID)

		body := map[string]interface{}{
			"group_id": ctx.TestGroup.ID,
			"pvc_id":   pvcID,
			"permissions": []map[string]string{
				{"user_id": testUser.ID, "permission": "read"},
				{"user_id": testUser2.ID, "permission": "read"},
			},
		}

		resp, err := client.POST("/storage/permissions/batch", body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("SetAccessPolicy - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := "/storage/policies"

		body := map[string]interface{}{
			"group_id":           ctx.TestGroup.ID,
			"pvc_id":             pvcID,
			"default_permission": "read",
			"admin_only":         false,
		}

		resp, err := client.POST(path, body)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("SetAccessPolicy - Invalid Policy", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := "/storage/policies"

		body := map[string]interface{}{
			"group_id":           ctx.TestGroup.ID,
			"pvc_id":             pvcID,
			"default_permission": "invalid_policy",
			"admin_only":         false,
		}

		resp, err := client.POST(path, body)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})
}
