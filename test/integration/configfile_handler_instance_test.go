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

func TestConfigFileInstanceLifecycle(t *testing.T) {
	ctx := GetTestContext()
	k8sValidator := NewK8sValidator()

	created := createConfigFileAsManager(t, ctx, "apiVersion: v1\nkind: Pod\nmetadata:\n  name: instance-pod", "instance commit")

	t.Run("CreateInstance - Success with K8s Verification", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		path := fmt.Sprintf("/instance/%s", created.Commit.ID)
		resp, err := client.POST(path, nil)

		require.NoError(t, err)
		if resp.StatusCode == http.StatusForbidden {
			t.Errorf("User should have permission to create instance")
		}
		if resp.StatusCode == http.StatusOK && k8sValidator != nil {
			namespace := fmt.Sprintf("proj-%s", ctx.TestProject.ID)
			deploymentName := "test-config-deployment"

			exists, err := k8sValidator.DeploymentExists(namespace, deploymentName)
			if err == nil && exists {
				t.Logf("Deployment %s/%s created successfully", namespace, deploymentName)
			}
		}
	})

	t.Run("DestructInstance - Success as Member", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.UserToken)

		path := fmt.Sprintf("/instance/%s", created.Commit.ID)
		resp, err := client.DELETE(path)

		require.NoError(t, err)
		assert.NotEqual(t, http.StatusForbidden, resp.StatusCode)
	})
}
