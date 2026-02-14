//go:build integration
// +build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProjectHandler_Integration(t *testing.T) {
	ctx := GetTestContext()

	var projectID string

	// Arrange + Act + Assert
	t.Run("CreateProject - Success as Admin", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)

		formData := map[string]string{
			"project_name": "integration-project",
			"description":  "project created by integration test",
			"g_id":         ctx.TestGroup.ID,
		}

		resp, err := client.POSTForm("/projects", formData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var created project.Project
		err = resp.DecodeData(&created)
		require.NoError(t, err)
		assert.NotEmpty(t, created.ID)
		assert.Equal(t, ctx.TestGroup.ID, *created.OwnerID)
		projectID = created.ID
	})

	t.Run("GetProjectByID - Success", func(t *testing.T) {
		if projectID == "" {
			t.Skip("No project created")
		}

		client := NewHTTPClient(ctx.Router, ctx.UserToken)
		path := fmt.Sprintf("/projects/%s", projectID)

		resp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var got project.Project
		err = resp.DecodeData(&got)
		require.NoError(t, err)
		assert.Equal(t, projectID, got.ID)
	})

	t.Run("UpdateProject - Success", func(t *testing.T) {
		if projectID == "" {
			t.Skip("No project created")
		}

		client := NewHTTPClient(ctx.Router, ctx.ManagerToken)
		path := fmt.Sprintf("/projects/%s", projectID)

		formData := map[string]string{
			"project_name": "integration-project-updated",
			"description":  "updated by integration test",
		}

		resp, err := client.PUTForm(path, formData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		getResp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, getResp.StatusCode)

		var updated project.Project
		err = getResp.DecodeData(&updated)
		require.NoError(t, err)
		assert.Equal(t, "integration-project-updated", updated.Name)
	})

	t.Run("GetProjectsByUser - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		resp, err := client.GET("/projects/by-user")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.NotEmpty(t, resp.Body)
	})

	t.Run("DeleteProject - Success as Admin", func(t *testing.T) {
		if projectID == "" {
			t.Skip("No project created")
		}

		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := fmt.Sprintf("/projects/%s", projectID)

		resp, err := client.DELETE(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		getResp, err := client.GET(path)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, getResp.StatusCode)
	})
}
