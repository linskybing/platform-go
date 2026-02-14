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

func TestConfigFileUpdate(t *testing.T) {
	ctx := GetTestContext()
	created := createConfigFileAsManager(t, ctx, "apiVersion: v1\nkind: Pod\nmetadata:\n  name: original-pod", "initial commit")

	t.Run("Success as Manager", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)

		formData := map[string]string{
			"raw_yaml": "apiVersion: v1\nkind: Pod\nmetadata:\n  name: updated-pod",
			"message":  "update commit",
		}

		path := fmt.Sprintf("/configfiles/%s", created.Commit.ID)
		resp, err := client.PUTForm(path, formData)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var updatedCommit configCommitResponse
		err = resp.DecodeJSON(&updatedCommit)
		require.NoError(t, err)
		assert.NotEmpty(t, updatedCommit.Commit.ID)
		assert.Contains(t, updatedCommit.Content, "updated-pod")

		getResp, err := client.GET(fmt.Sprintf("/configfiles/%s", updatedCommit.Commit.ID))
		require.NoError(t, err)

		var fetched configCommitResponse
		err = getResp.DecodeJSON(&fetched)
		require.NoError(t, err)
		assert.Contains(t, fetched.Content, "updated-pod")
	})

	t.Run("Forbidden for Regular User", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		formData := map[string]string{
			"filename": "forbidden-update.yaml",
		}

		path := fmt.Sprintf("/configfiles/%s", created.Commit.ID)
		resp, err := client.PUTForm(path, formData)

		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}
