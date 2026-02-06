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
	generator := NewTestDataGenerator()
	cleaner := NewDatabaseCleaner()
	t.Cleanup(func() {
		_ = cleaner.Cleanup()
	})

	testUser := generator.GenerateUser("storage-perm-test")
	require.NoError(t, generator.CreateTestUser(testUser))
	cleaner.RegisterUser(testUser.UID)

	testProject := generator.GenerateProject("storage-project", ctx.TestGroup.GID)
	require.NoError(t, generator.CreateTestProject(testProject))
	cleaner.RegisterProject(testProject.PID)

	t.Run("SetPermission - Success as Admin", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		formData := map[string]string{
			"user_id":    testUser.UID,
			"project_id": testProject.PID,
			"permission": "rw",
		}

		resp, err := client.POSTForm("/storage/permissions", formData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("SetPermission - Invalid Permission Type", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		formData := map[string]string{
			"user_id":    testUser.UID,
			"project_id": testProject.PID,
			"permission": "invalid",
		}

		resp, err := client.POSTForm("/storage/permissions", formData)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetUserPermission - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)
		path := fmt.Sprintf("/storage/permissions/%s/%s", testUser.UID, testProject.PID)

		resp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("BatchSetPermissions - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		testUser2 := generator.GenerateUser("storage-test-2")
		require.NoError(t, generator.CreateTestUser(testUser2))
		cleaner.RegisterUser(testUser2.UID)

		formData := map[string]string{
			"user_ids":   fmt.Sprintf("%s,%s", testUser.UID, testUser2.UID),
			"project_id": testProject.PID,
			"permission": "r",
		}

		resp, err := client.POSTForm("/storage/permissions/batch", formData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("SetAccessPolicy - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := fmt.Sprintf("/storage/projects/%s/policy", testProject.PID)

		formData := map[string]string{
			"policy": "private",
		}

		resp, err := client.POSTForm(path, formData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("SetAccessPolicy - Invalid Policy", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := fmt.Sprintf("/storage/projects/%s/policy", testProject.PID)

		formData := map[string]string{
			"policy": "invalid_policy",
		}

		resp, err := client.POSTForm(path, formData)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})
}
