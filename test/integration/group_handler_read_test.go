//go:build integration
// +build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupRead(t *testing.T) {
	ctx := GetTestContext()

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
}
