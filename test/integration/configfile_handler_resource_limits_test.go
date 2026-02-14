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

func TestConfigFileHandler_ResourceLimits(t *testing.T) {
	ctx := GetTestContext()

	tests := []struct {
		name        string
		cpuRequest  string
		memRequest  string
		cpuLimit    string
		memLimit    string
		expectError bool
		description string
	}{
		{
			name:        "Valid small resources",
			cpuRequest:  "100m",
			memRequest:  "128Mi",
			cpuLimit:    "200m",
			memLimit:    "256Mi",
			expectError: false,
		},
		{
			name:        "Valid large resources",
			cpuRequest:  "2",
			memRequest:  "4Gi",
			cpuLimit:    "4",
			memLimit:    "8Gi",
			expectError: false,
		},
		{
			name:        "Limit less than request (CPU)",
			cpuRequest:  "1000m",
			memRequest:  "1Gi",
			cpuLimit:    "500m",
			memLimit:    "2Gi",
			expectError: true,
			description: "CPU limit should be >= request",
		},
		{
			name:        "Limit less than request (Memory)",
			cpuRequest:  "500m",
			memRequest:  "2Gi",
			cpuLimit:    "1000m",
			memLimit:    "1Gi",
			expectError: true,
			description: "Memory limit should be >= request",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewHTTPClient(ctx.Router, ctx.ManagerToken)

			formData := map[string]string{
				"project_id": ctx.TestProject.PID,
				"filename":   "resource-test-" + tt.name + ".yaml",
				"raw_yaml":   fmt.Sprintf("apiVersion: v1\nkind: Pod\nmetadata:\n  name: resource-test\nspec:\n  containers:\n  - name: test\n    image: nginx:latest\n    resources:\n      requests:\n        cpu: %s\n        memory: %s\n      limits:\n        cpu: %s\n        memory: %s", tt.cpuRequest, tt.memRequest, tt.cpuLimit, tt.memLimit),
			}

			resp, err := client.POSTFormRaw("/configfiles", formData)
			require.NoError(t, err)

			if tt.expectError {
				assert.GreaterOrEqual(t, resp.StatusCode, 400, tt.description)
			} else {
				assert.Equal(t, http.StatusCreated, resp.StatusCode)
			}
		})
	}
}
