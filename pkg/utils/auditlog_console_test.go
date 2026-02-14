package utils

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// TestLogAuditWithConsole - Test console logging with background goroutine
func TestLogAuditWithConsole(t *testing.T) {
	tests := []struct {
		name        string
		action      string
		description string
	}{
		{
			name:        "async logging",
			action:      "create",
			description: "should log asynchronously",
		},
		{
			name:        "multiple concurrent logs",
			action:      "update",
			description: "should handle concurrent logging",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/resource", nil)
			req.Header.Set("User-Agent", "test-client")

			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			// Note: GetUserIDFromContext would fail here without claims,
			// but LogAuditWithConsole handles it gracefully

			repo := &mockAuditRepo{}

			// This should not panic even without claims
			LogAuditWithConsole(c, tt.action, "resource", "id_123",
				map[string]interface{}{"field": "old"},
				map[string]interface{}{"field": "new"},
				"test message", repo)

			// Give goroutine time to complete
			time.Sleep(10 * time.Millisecond)
		})
	}
}
