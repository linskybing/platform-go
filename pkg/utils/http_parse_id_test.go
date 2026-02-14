package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
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
