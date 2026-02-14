//go:build integration
// +build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupCreate(t *testing.T) {
	ctx := GetTestContext()

	t.Run("CreateGroup - Success as Admin", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		createDTO := map[string]string{
			"group_name": "test-integration-group",
		}

		resp, err := client.POSTForm("/groups", createDTO)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
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
			"group_name": ctx.TestGroup.Name,
		}

		resp, err := client.POSTForm("/groups", createDTO)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, resp.StatusCode, 400)
	})
}
