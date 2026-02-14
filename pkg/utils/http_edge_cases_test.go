package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParsingEdgeCases - Complex edge cases
func TestParsingEdgeCases(t *testing.T) {
	t.Run("consecutive parsing", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/resource/123", nil)

		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "123"}}

		id1, err1 := ParseIDParam(c, "id")
		require.NoError(t, err1)
		require.Equal(t, "123", id1)

		id2, err2 := ParseIDParam(c, "id")
		require.NoError(t, err2)
		require.Equal(t, "123", id2)
		require.Equal(t, id1, id2)
	})

	t.Run("multiple query parameters", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/?page=2&limit=10&sort=id", nil)

		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		page, err1 := ParseQueryIDParam(c, "page")
		require.NoError(t, err1)
		require.Equal(t, "2", page)

		limit, err2 := ParseQueryIDParam(c, "limit")
		require.NoError(t, err2)
		require.Equal(t, "10", limit)
	})

	t.Run("parameter value overrides", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/?id=100&id=200", nil)

		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(w)
		c.Request = req

		val, err := ParseQueryIDParam(c, "id")
		require.NoError(t, err)
		assert.True(t, val == "100" || val == "200")
	})
}
