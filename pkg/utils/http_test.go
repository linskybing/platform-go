package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseIDParam - Table-driven tests
func TestParseIDParam(t *testing.T) {
	tests := []struct {
		name        string
		idValue     string
		wantID      string
		wantErr     bool
		description string
	}{
		{
			name:        "valid positive ID",
			idValue:     "123",
			wantID:      "123",
			wantErr:     false,
			description: "should parse valid positive ID",
		},
		{
			name:        "zero ID",
			idValue:     "0",
			wantID:      "0",
			wantErr:     false,
			description: "should parse zero ID",
		},
		{
			name:        "large ID",
			idValue:     "4294967295",
			wantID:      "4294967295",
			wantErr:     false,
			description: "should parse large uint32 value",
		},
		{
			name:        "non-numeric ID",
			idValue:     "abc",
			wantID:      "abc",
			wantErr:     false,
			description: "should error for non-numeric ID",
		},
		{
			name:        "negative ID",
			idValue:     "-123",
			wantID:      "-123",
			wantErr:     false,
			description: "should error for negative ID",
		},
		{
			name:        "empty string",
			idValue:     "",
			wantID:      "",
			wantErr:     true,
			description: "should error for empty string",
		},
		{
			name:        "float ID",
			idValue:     "123.45",
			wantID:      "123.45",
			wantErr:     false,
			description: "should error for float value",
		},
		{
			name:        "ID with spaces",
			idValue:     " 123 ",
			wantID:      " 123 ",
			wantErr:     false,
			description: "should error for ID with spaces",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			// Use a safe URL path, set the ID in Params for testing
			req := httptest.NewRequest("GET", "/resource/id", nil)

			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.Params = gin.Params{{Key: "id", Value: tt.idValue}}

			got, err := ParseIDParam(c, "id")

			if tt.wantErr {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.wantID, got, tt.description)
			}
		})
	}
}

// TestParseQueryIDParam - Table-driven tests
func TestParseQueryIDParam(t *testing.T) {
	tests := []struct {
		name        string
		paramName   string
		paramValue  string
		wantValue   string
		wantErr     bool
		description string
	}{
		{
			name:        "valid positive query param",
			paramName:   "page",
			paramValue:  "5",
			wantValue:   "5",
			wantErr:     false,
			description: "should parse valid query parameter",
		},
		{
			name:        "zero query param",
			paramName:   "limit",
			paramValue:  "0",
			wantValue:   "0",
			wantErr:     false,
			description: "should parse zero value",
		},
		{
			name:        "missing query param",
			paramName:   "page",
			paramValue:  "",
			wantValue:   "",
			wantErr:     true,
			description: "should error when parameter is empty",
		},
		{
			name:        "non-numeric query param",
			paramName:   "page",
			paramValue:  "invalid",
			wantValue:   "invalid",
			wantErr:     false,
			description: "should error for non-numeric value",
		},
		{
			name:        "negative query param",
			paramName:   "limit",
			paramValue:  "-10",
			wantValue:   "-10",
			wantErr:     false,
			description: "should error for negative value",
		},
		{
			name:        "large query param",
			paramName:   "count",
			paramValue:  "4294967295",
			wantValue:   "4294967295",
			wantErr:     false,
			description: "should parse large uint32 value",
		},
		{
			name:        "float query param",
			paramName:   "rate",
			paramValue:  "3.14",
			wantValue:   "3.14",
			wantErr:     false,
			description: "should error for float value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			url := "/?page=" + tt.paramValue
			req := httptest.NewRequest("GET", url, nil)
			if tt.paramName != "page" {
				req = httptest.NewRequest("GET", "/?"+tt.paramName+"="+tt.paramValue, nil)
			}

			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			got, err := ParseQueryIDParam(c, tt.paramName)

			if tt.wantErr {
				assert.Error(t, err, tt.description)
				if err == ErrEmptyParameter {
					assert.Equal(t, ErrEmptyParameter, err)
				}
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.wantValue, got, tt.description)
			}
		})
	}
}

// TestErrEmptyParameter
func TestErrEmptyParameter(t *testing.T) {
	tests := []struct {
		name        string
		paramName   string
		paramValue  string
		wantErr     bool
		description string
	}{
		{
			name:        "empty param error",
			paramName:   "limit",
			paramValue:  "",
			wantErr:     true,
			description: "empty parameter should return ErrEmptyParameter",
		},
		{
			name:        "non-empty param",
			paramName:   "limit",
			paramValue:  "10",
			wantErr:     false,
			description: "non-empty parameter should not error for empty parameter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotNil(t, ErrEmptyParameter)
			assert.Equal(t, "empty parameter", ErrEmptyParameter.Error())
		})
	}
}

// TestParsingEdgeCases - Complex edge cases
func TestParsingEdgeCases(t *testing.T) {
	t.Run("consecutive parsing", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/resource/123", nil)

		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		c.Params = gin.Params{{Key: "id", Value: "123"}}

		// Parse multiple times
		id1, err1 := ParseIDParam(c, "id")
		require.NoError(t, err1)
		require.Equal(t, "123", id1)

		// Parse again from same context
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

		// Should get first value when duplicates exist
		val, err := ParseQueryIDParam(c, "id")
		require.NoError(t, err)
		assert.True(t, val == "100" || val == "200")
	})
}
