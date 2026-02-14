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

func TestConfigFileDelete(t *testing.T) {
	ctx := GetTestContext()

	t.Run("DeleteConfigFile - Success as Manager", func(t *testing.T) {
		created := createConfigFileAsManager(t, ctx, "apiVersion: v1\nkind: Pod\nmetadata:\n  name: config-to-delete\nspec:\n  containers:\n  - name: nginx\n    image: nginx:latest\n", "delete commit")

		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)
		path := fmt.Sprintf("/configfiles/%s", created.Commit.ID)
		deleteResp, err := client.DELETE(path)

		require.NoError(t, err)
		assert.Contains(t, []int{http.StatusOK, http.StatusNoContent}, deleteResp.StatusCode)

		getResp, err := client.GET(path)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, getResp.StatusCode)
	})

	t.Run("DeleteConfigFile - Forbidden for Regular User", func(t *testing.T) {
		created := createConfigFileAsManager(t, ctx, "apiVersion: v1\nkind: Pod\nmetadata:\n  name: config-forbidden-delete", "delete commit")

		client := NewHTTPClient(ctx.Router, ctx.UserToken)
		path := fmt.Sprintf("/configfiles/%s", created.Commit.ID)
		resp, err := client.DELETE(path)

		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
