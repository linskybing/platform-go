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

func TestGroupUpdate(t *testing.T) {
	ctx := GetTestContext()

	t.Run("UpdateGroup - Success as Admin", func(t *testing.T) {
		created := createGroupAsAdmin(t, ctx, "group-update-test")
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		updateDTO := map[string]string{
			"group_name": "updated-group-name",
		}

		path := fmt.Sprintf("/groups/%s", created.ID)
		resp, err := client.PUTForm(path, updateDTO)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		getResp, err := client.GET(path)
		require.NoError(t, err)

		var updated group.Group
		err = getResp.DecodeData(&updated)
		require.NoError(t, err)
		assert.Equal(t, "updated-group-name", updated.Name)
	})

	t.Run("UpdateGroup - Forbidden for Regular User", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		updateDTO := map[string]string{
			"group_name": "hacked",
		}

		path := fmt.Sprintf("/groups/%s", ctx.TestGroup.ID)
		resp, err := client.PUTForm(path, updateDTO)

		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("UpdateGroup - Cannot Update Reserved Group", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		updateDTO := map[string]string{
			"group_name": "new-super-name",
		}

		path := fmt.Sprintf("/groups/%s", ctx.SuperGroup.ID)
		resp, err := client.PUTForm(path, updateDTO)

		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		getResp, err := client.GET(path)
		require.NoError(t, err)

		var g group.Group
		err = getResp.DecodeData(&g)
		require.NoError(t, err)
		assert.Equal(t, config.ReservedGroupName, g.Name, "Super group name should remain unchanged")
	})
}
