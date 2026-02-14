package utils

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestGetUserIDFromContext(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		username    string
		hasClaims   bool
		validType   bool
		wantErr     bool
		wantUserID  string
		description string
	}{
		{
			name:        "valid claims",
			userID:      "123",
			username:    "testuser",
			hasClaims:   true,
			validType:   true,
			wantErr:     false,
			wantUserID:  "123",
			description: "should extract user ID from valid claims",
		},
		{
			name:        "no claims in context",
			userID:      "",
			username:    "",
			hasClaims:   false,
			validType:   false,
			wantErr:     true,
			wantUserID:  "",
			description: "should error when claims not in context",
		},
		{
			name:        "invalid claims type",
			userID:      "",
			username:    "",
			hasClaims:   true,
			validType:   false,
			wantErr:     true,
			wantUserID:  "",
			description: "should error on type assertion failure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)

			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			if tt.hasClaims {
				if tt.validType {
					c.Set("claims", &types.Claims{UserID: tt.userID, Username: tt.username})
				} else {
					c.Set("claims", "invalid")
				}
			}

			got, err := GetUserIDFromContext(c)

			if tt.wantErr {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.wantUserID, got, tt.description)
			}
		})
	}
}

func TestGetUserNameFromContext(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		username    string
		hasClaims   bool
		validType   bool
		wantErr     bool
		wantName    string
		description string
	}{
		{
			name:        "valid claims",
			userID:      "123",
			username:    "alice",
			hasClaims:   true,
			validType:   true,
			wantErr:     false,
			wantName:    "alice",
			description: "should extract username from valid claims",
		},
		{
			name:        "empty username",
			userID:      "456",
			username:    "",
			hasClaims:   true,
			validType:   true,
			wantErr:     false,
			wantName:    "",
			description: "should handle empty username",
		},
		{
			name:        "no claims",
			userID:      "",
			username:    "",
			hasClaims:   false,
			validType:   false,
			wantErr:     true,
			wantName:    "",
			description: "should error when claims not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/", nil)

			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(w)
			c.Request = req

			if tt.hasClaims {
				if tt.validType {
					c.Set("claims", &types.Claims{UserID: tt.userID, Username: tt.username})
				} else {
					c.Set("claims", 123)
				}
			}

			got, err := GetUserNameFromContext(c)

			if tt.wantErr {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.wantName, got, tt.description)
			}
		})
	}
}
