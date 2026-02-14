//go:build integration
// +build integration

package integration

import (
	"net/http"
	"testing"

	"github.com/linskybing/platform-go/internal/domain/configfile"
	"github.com/stretchr/testify/require"
)

type configCommitResponse struct {
	Commit  configfile.ConfigCommit `json:"commit"`
	Content string                  `json:"content"`
}

func createConfigFileAsManager(t *testing.T, ctx *TestContext, rawYaml, message string) configCommitResponse {
	t.Helper()
	client := NewHTTPClient(ctx.Router, ctx.ManagerToken)

	formData := map[string]string{
		"project_id": ctx.TestProject.ID,
		"raw_yaml":   rawYaml,
	}
	if message != "" {
		formData["message"] = message
	}

	resp, err := client.POSTFormRaw("/configfiles", formData)
	require.NoError(t, err)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}

	var created configCommitResponse
	err = resp.DecodeData(&created)
	require.NoError(t, err)
	return created
}
