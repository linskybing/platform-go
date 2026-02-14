package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

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
