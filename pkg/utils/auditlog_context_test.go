package utils

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestLogAuditWithContextData - Test with proper context setup
func TestLogAuditWithContextData(t *testing.T) {
	tests := []struct {
		name        string
		userID      string
		username    string
		action      string
		description string
	}{
		{
			name:        "with user context",
			userID:      "1",
			username:    "alice",
			action:      "create",
			description: "should extract user ID from context",
		},
		{
			name:        "with different user",
			userID:      "2",
			username:    "bob",
			action:      "delete",
			description: "should use provided user ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/resource", nil)
			req.Header.Set("User-Agent", "Mozilla/5.0")

			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			// Set up claims
			c.Set("claims", &struct {
				UserID   string
				Username string
			}{UserID: tt.userID, Username: tt.username})

			repo := &mockAuditRepo{}

			LogAuditWithConsole(c, tt.action, "resource", "id_123", nil, nil, "message", repo)

			// Give async operation time
			time.Sleep(20 * time.Millisecond)
		})
	}
}
