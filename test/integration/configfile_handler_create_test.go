//go:build integration
// +build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateConfigFile(t *testing.T) {
	ctx := GetTestContext()

	t.Run("Success as Manager", func(t *testing.T) {
		created := createConfigFileAsManager(t, ctx, "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test-pod", "initial commit")
		assert.Equal(t, ctx.TestProject.PID, created.Commit.ProjectID)
		assert.NotEmpty(t, created.Commit.ID)
		assert.Contains(t, created.Content, "test-pod")
	})

	t.Run("Forbidden for Regular User", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		formData := map[string]string{
			"project_id": ctx.TestProject.PID,
			"filename":   "unauthorized-config.yaml",
			"raw_yaml":   "apiVersion: v1\nkind: Pod\nmetadata:\n  name: unauthorized-pod",
		}

		resp, err := client.POSTFormRaw("/configfiles", formData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Invalid Input Validation", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)

		tests := []struct {
			name  string
			input map[string]interface{}
		}{
			{
				name: "Missing project_id",
				input: map[string]interface{}{
					"raw_yaml": "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test-pod",
				},
			},
			{
				name: "Missing raw_yaml",
				input: map[string]interface{}{
					"project_id": ctx.TestProject.PID,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := client.POST("/configfiles", tt.input)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, resp.StatusCode, 400)
			})
		}
	})
}
