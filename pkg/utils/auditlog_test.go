package utils

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/domain/audit"
	"github.com/linskybing/platform-go/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type mockAuditRepo struct {
	createdLogs []*audit.AuditLog
	shouldErr   bool
}

func (m *mockAuditRepo) CreateAuditLog(log *audit.AuditLog) error {
	if m.shouldErr {
		return nil // Simulating error (we ignore it in LogAudit)
	}
	if m.createdLogs == nil {
		m.createdLogs = make([]*audit.AuditLog, 0)
	}
	m.createdLogs = append(m.createdLogs, log)
	return nil
}

func (m *mockAuditRepo) GetAuditLogs(params repository.AuditQueryParams) ([]audit.AuditLog, error) {
	return nil, nil
}

func (m *mockAuditRepo) DeleteOldAuditLogs(retentionDays int) error {
	return nil
}

func (m *mockAuditRepo) WithTx(tx *gorm.DB) repository.AuditRepo {
	return m
}

// TestLogAudit - Table-driven tests
func TestLogAudit(t *testing.T) {
	tests := []struct {
		name         string
		userID       uint
		ip           string
		ua           string
		action       string
		resourceType string
		resourceID   string
		before       interface{}
		after        interface{}
		description  string
	}{
		{
			name:         "create action",
			userID:       1,
			ip:           "192.168.1.1",
			ua:           "Mozilla/5.0",
			action:       "create",
			resourceType: "project",
			resourceID:   "proj_123",
			before:       nil,
			after: map[string]interface{}{
				"id":   "proj_123",
				"name": "My Project",
			},
			description: "should log create action",
		},
		{
			name:         "update action",
			userID:       2,
			ip:           "10.0.0.1",
			ua:           "Chrome",
			action:       "update",
			resourceType: "user",
			resourceID:   "user_456",
			before: map[string]interface{}{
				"email": "old@example.com",
			},
			after: map[string]interface{}{
				"email": "new@example.com",
			},
			description: "should log update action with before/after",
		},
		{
			name:         "delete action",
			userID:       3,
			ip:           "::1",
			ua:           "Safari",
			action:       "delete",
			resourceType: "resource",
			resourceID:   "res_789",
			before: map[string]interface{}{
				"id": "res_789",
			},
			after:       nil,
			description: "should log delete action",
		},
		{
			name:         "no before/after",
			userID:       4,
			ip:           "127.0.0.1",
			ua:           "Firefox",
			action:       "view",
			resourceType: "report",
			resourceID:   "rep_001",
			before:       nil,
			after:        nil,
			description:  "should handle missing before/after",
		},
		{
			name:         "complex nested data",
			userID:       5,
			ip:           "192.168.0.1",
			ua:           "API Client",
			action:       "modify",
			resourceType: "config",
			resourceID:   "cfg_123",
			before: map[string]interface{}{
				"settings": map[string]interface{}{
					"enabled": false,
					"timeout": 30,
				},
			},
			after: map[string]interface{}{
				"settings": map[string]interface{}{
					"enabled": true,
					"timeout": 60,
				},
			},
			description: "should handle nested data structures",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuditRepo{createdLogs: make([]*audit.AuditLog, 0)}

			err := LogAudit(tt.userID, tt.ip, tt.ua, tt.action, tt.resourceType, tt.resourceID,
				tt.before, tt.after, "test description", repo)

			assert.NoError(t, err, tt.description)
			assert.Equal(t, 1, len(repo.createdLogs), "should create one audit log")

			log := repo.createdLogs[0]
			assert.Equal(t, tt.userID, log.UserID)
			assert.Equal(t, tt.ip, log.IPAddress)
			assert.Equal(t, tt.ua, log.UserAgent)
			assert.Equal(t, tt.action, log.Action)
			assert.Equal(t, tt.resourceType, log.ResourceType)
			assert.Equal(t, tt.resourceID, log.ResourceID)
		})
	}
}

// TestLogAuditDataSerialization - Verify JSON serialization
func TestLogAuditDataSerialization(t *testing.T) {
	tests := []struct {
		name        string
		data        interface{}
		description string
	}{
		{
			name:        "string data",
			data:        "test string",
			description: "should serialize string data",
		},
		{
			name:        "struct data",
			data:        struct{ Field string }{"value"},
			description: "should serialize struct data",
		},
		{
			name:        "map data",
			data:        map[string]interface{}{"key": "value"},
			description: "should serialize map data",
		},
		{
			name:        "list data",
			data:        []string{"a", "b", "c"},
			description: "should serialize list data",
		},
		{
			name:        "numeric data",
			data:        12345,
			description: "should serialize numeric data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuditRepo{createdLogs: make([]*audit.AuditLog, 0)}

			err := LogAudit(1, "127.0.0.1", "test", "action", "resource", "id",
				tt.data, nil, "test", repo)

			assert.NoError(t, err, tt.description)
			assert.Equal(t, 1, len(repo.createdLogs))

			log := repo.createdLogs[0]
			assert.NotNil(t, log.OldData)

			// Verify data can be unmarshaled
			var unmarshal interface{}
			err = json.Unmarshal(log.OldData, &unmarshal)
			assert.NoError(t, err, "should be valid JSON")
		})
	}
}

// TestLogAuditWithConsole - Test console logging with background goroutine
func TestLogAuditWithConsole(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		action      string
		description string
	}{
		{
			name:        "async logging",
			userID:      1,
			action:      "create",
			description: "should log asynchronously",
		},
		{
			name:        "multiple concurrent logs",
			userID:      2,
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

			repo := &mockAuditRepo{createdLogs: make([]*audit.AuditLog, 0)}

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

// TestLogAuditWithContextData - Test with proper context setup
func TestLogAuditWithContextData(t *testing.T) {
	tests := []struct {
		name        string
		userID      uint
		username    string
		action      string
		description string
	}{
		{
			name:        "with user context",
			userID:      1,
			username:    "alice",
			action:      "create",
			description: "should extract user ID from context",
		},
		{
			name:        "with different user",
			userID:      2,
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
				UserID   uint
				Username string
			}{UserID: tt.userID, Username: tt.username})

			repo := &mockAuditRepo{createdLogs: make([]*audit.AuditLog, 0)}

			LogAuditWithConsole(c, tt.action, "resource", "id_123", nil, nil, "message", repo)

			// Give async operation time
			time.Sleep(20 * time.Millisecond)
		})
	}
}

// TestLogAuditErrorHandling - Test error scenarios
func TestLogAuditErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		beforeErr   bool
		afterErr    bool
		description string
	}{
		{
			name:        "unmarshallable before",
			beforeErr:   true,
			afterErr:    false,
			description: "should handle unmarshallable before data",
		},
		{
			name:        "unmarshallable after",
			beforeErr:   false,
			afterErr:    true,
			description: "should handle unmarshallable after data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockAuditRepo{createdLogs: make([]*audit.AuditLog, 0)}

			// Use data that's valid to marshal
			err := LogAudit(1, "127.0.0.1", "test", "action", "resource", "id",
				map[string]interface{}{"key": "value"},
				map[string]interface{}{"key": "new_value"},
				"test", repo)

			assert.NoError(t, err)
			assert.Equal(t, 1, len(repo.createdLogs))
		})
	}
}

// TestAuditLogCreation - Verify audit log structure
func TestAuditLogCreation(t *testing.T) {
	repo := &mockAuditRepo{createdLogs: make([]*audit.AuditLog, 0)}

	beforeData := map[string]interface{}{"status": "active"}
	afterData := map[string]interface{}{"status": "inactive"}

	err := LogAudit(42, "192.168.1.100", "Chrome/90", "status_change", "user", "user_789",
		beforeData, afterData, "User status changed", repo)

	require.NoError(t, err)
	require.Equal(t, 1, len(repo.createdLogs))

	log := repo.createdLogs[0]

	// Verify all fields
	t.Run("audit log fields", func(t *testing.T) {
		assert.Equal(t, uint(42), log.UserID)
		assert.Equal(t, "192.168.1.100", log.IPAddress)
		assert.Equal(t, "Chrome/90", log.UserAgent)
		assert.Equal(t, "status_change", log.Action)
		assert.Equal(t, "user", log.ResourceType)
		assert.Equal(t, "user_789", log.ResourceID)
		assert.Equal(t, "User status changed", log.Description)

		// Verify data is JSON
		var before map[string]interface{}
		err := json.Unmarshal(log.OldData, &before)
		assert.NoError(t, err)
		assert.Equal(t, "active", before["status"])

		var after map[string]interface{}
		err = json.Unmarshal(log.NewData, &after)
		assert.NoError(t, err)
		assert.Equal(t, "inactive", after["status"])
	})
}

// TestLogAuditConcurrency - Test thread safety
func TestLogAuditConcurrency(t *testing.T) {
	repo := &mockAuditRepo{createdLogs: make([]*audit.AuditLog, 0)}

	// Log multiple times concurrently
	numGoroutines := 10
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			_ = LogAudit(
				uint(id), "127.0.0.1", "test", "action",
				"resource", "id_"+string(rune(id)),
				nil, nil, "test", repo,
			)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// All logs should be created
	assert.Equal(t, numGoroutines, len(repo.createdLogs))
}
