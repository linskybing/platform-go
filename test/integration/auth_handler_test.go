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

func TestAuthAndMiscHandlers_Integration(t *testing.T) {
	ctx := GetTestContext()

	t.Run("AuthStatus - Success with Valid Token", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.RegularToken)

		resp, err := client.GET("/auth/status")
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("AuthStatus - Unauthorized without Token", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, "")

		resp, err := client.GET("/auth/status")
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("AuthStatus - Invalid Token", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, "invalid-token")

		resp, err := client.GET("/auth/status")
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})
}

func TestUserRegistrationAndLogin_Integration(t *testing.T) {
	ctx := GetTestContext()
	generator := NewTestDataGenerator()
	cleaner := NewDatabaseCleaner()
	t.Cleanup(func() {
		_ = cleaner.Cleanup()
	})

	username := fmt.Sprintf("register-test-%d", generator.RandomInt(10000, 99999))
	password := "testPass123!"

	t.Run("Register - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, "")

		formData := map[string]string{
			"username":  username,
			"password":  password,
			"email":     fmt.Sprintf("%s@test.com", username),
			"full_name": "Test User",
		}

		resp, err := client.POSTForm("/register", formData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = resp.DecodeJSON(&result)
		require.NoError(t, err)

		if uid, ok := result["user_id"].(string); ok {
			cleaner.RegisterUser(uid)
		}
	})

	t.Run("Register - Duplicate Username", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, "")

		formData := map[string]string{
			"username":  username,
			"password":  password,
			"email":     "another@test.com",
			"full_name": "Test User",
		}

		resp, err := client.POSTForm("/register", formData)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Login - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, "")

		formData := map[string]string{
			"username": username,
			"password": password,
		}

		resp, err := client.POSTForm("/login", formData)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = resp.DecodeJSON(&result)
		require.NoError(t, err)
		assert.NotEmpty(t, result["token"])
	})

	t.Run("Login - Wrong Password", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, "")

		formData := map[string]string{
			"username": username,
			"password": "wrongpassword",
		}

		resp, err := client.POSTForm("/login", formData)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Login - Nonexistent User", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, "")

		formData := map[string]string{
			"username": "nonexistent-user",
			"password": "anypassword",
		}

		resp, err := client.POSTForm("/login", formData)
		require.NoError(t, err)
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Logout - Success", func(t *testing.T) {
		client := NewHTTPClient(ctx.Router, ctx.RegularToken)

		resp, err := client.POST("/logout", nil)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
