//go:build integration
// +build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigFileRead(t *testing.T) {
	ctx := GetTestContext()
	created := createConfigFileAsManager(t, ctx, "apiVersion: v1\nkind: Pod\nmetadata:\n  name: read-pod", "read commit")

	t.Run("GetConfigFile - Success as Member", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		path := fmt.Sprintf("/configfiles/%s", created.Commit.ID)
		resp, err := client.GET(path)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var cf configCommitResponse
		err = resp.DecodeData(&cf)
		require.NoError(t, err)
		assert.Equal(t, created.Commit.ID, cf.Commit.ID)
	})

	t.Run("ListConfigFiles - Admin and User", func(t *testing.T) {
		adminClient := NewHTTPClient(ctx.Router, ctx.AdminToken)
		resp, err := adminClient.GET("/configfiles")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		userClient := NewHTTPClient(ctx.Router, ctx.UserToken)
		resp, err = userClient.GET("/configfiles")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ListConfigFilesByProject - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		path := fmt.Sprintf("/projects/%s/config-files", ctx.TestProject.ID)
		resp, err := client.GET(path)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var configs []configfile.ConfigCommit
		err = resp.DecodeData(&configs)
		require.NoError(t, err)
	})
}
