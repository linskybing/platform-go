//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/linskybing/platform-go/internal/domain/group"
	"github.com/stretchr/testify/require"
)

func createGroupAsAdmin(t *testing.T, ctx *TestContext, name string) group.Group {
	t.Helper()
	client := NewHTTPClient(ctx.Router, ctx.AdminToken)

	createDTO := map[string]string{
		"group_name": name,
	}

	resp, err := client.POSTForm("/groups", createDTO)
	require.NoError(t, err)

	var created group.Group
	err = resp.DecodeData(&created)
	require.NoError(t, err)
	return created
}
