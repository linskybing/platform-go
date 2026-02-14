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

func TestGroupDelete(t *testing.T) {
	ctx := GetTestContext()

	t.Run("DeleteGroup - Success as Admin", func(t *testing.T) {
		created := createGroupAsAdmin(t, ctx, "group-to-delete")
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		path := fmt.Sprintf("/groups/%s", created.ID)
		deleteResp, err := client.DELETE(path)

		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, deleteResp.StatusCode)

		getResp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, getResp.StatusCode)
	})

	t.Run("DeleteGroup - Forbidden for Manager", func(t *testing.T) {
		created := createGroupAsAdmin(t, ctx, "manager-forbidden-delete")
		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)

		path := fmt.Sprintf("/groups/%s", created.ID)
		resp, err := client.DELETE(path)

		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("DeleteGroup - Cannot Delete Reserved Group", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		path := fmt.Sprintf("/groups/%s", ctx.SuperGroup.ID)
		resp, err := client.DELETE(path)

		require.NoError(t, err)
		assert.GreaterOrEqual(t, resp.StatusCode, 400)
	})
}
