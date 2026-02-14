//go:build integration
// +build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/linskybing/platform-go/internal/domain/user"
)

func TestUserHandler_Integration(t *testing.T) {
	ctx := GetTestContext()
	generator := NewTestDataGenerator()
	cleaner := NewDatabaseCleaner()
	t.Cleanup(func() {
		_ = cleaner.Cleanup()
	})

	testUser := generator.GenerateUser("integration")
	require.NoError(t, generator.CreateTestUser(testUser))
	cleaner.RegisterUser(testUser.ID)

	t.Run("GetUserByID - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := fmt.Sprintf("/users/%s", testUser.ID)

		resp, err := client.GET(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		
		var u user.User
		err = resp.DecodeData(&u)
		require.NoError(t, err)
		assert.Equal(t, testUser.Username, u.Username)
	})

	t.Run("UpdateUser - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := fmt.Sprintf("/users/%s", testUser.ID)

		newEmail := "updated-integration@test.com"
		formData := map[string]string{
			"email":  newEmail,
			"status": string(user.UserStatusOnline),
		}

		resp, err := client.PUTForm(path, formData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("DeleteUser - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.AdminToken)
		path := fmt.Sprintf("/users/%s", testUser.ID)

		resp, err := client.DELETE(path)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		getResp, err := client.GET(path)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, getResp.StatusCode)
	})
}
