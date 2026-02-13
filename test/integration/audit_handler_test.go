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

func TestAuditHandler_Integration(t *testing.T) {
	ctx := GetTestContext()
	generator := NewTestDataGenerator()
	cleaner := NewDatabaseCleaner()
	t.Cleanup(func() {
		_ = cleaner.Cleanup()
	})

	testUser := generator.GenerateUser("audit-test")
	require.NoError(t, generator.CreateTestUser(testUser))
	cleaner.RegisterUser(testUser.UID)

	testGroup := generator.GenerateGroup("audit-group")
	require.NoError(t, generator.CreateTestGroup(testGroup))
	cleaner.RegisterGroup(testGroup.GID)

	t.Run("GetAuditLogs - Success as Admin", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		resp, err := client.GET("/audit/logs")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, resp.Body)
	})

	t.Run("GetAuditLogs - With Pagination", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		resp, err := client.GET("/audit/logs?page=1&limit=10")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetAuditLogs - Filter by User", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := fmt.Sprintf("/audit/logs?user_id=%s", testUser.UID)

		resp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetAuditLogs - Filter by Action", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		resp, err := client.GET("/audit/logs?action=create")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("GetAuditLogs - Forbidden for Regular User", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		resp, err := client.GET("/audit/logs")
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})
}
