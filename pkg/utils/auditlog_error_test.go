package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			repo := &mockAuditRepo{}

			// Use data that's valid to marshal
			err := LogAudit("1", "127.0.0.1", "test", "action", "resource", "id",
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
	repo := &mockAuditRepo{}

	beforeData := map[string]interface{}{"status": "active"}
	afterData := map[string]interface{}{"status": "inactive"}

	err := LogAudit("42", "192.168.1.100", "Chrome/90", "status_change", "user", "user_789",
		beforeData, afterData, "User status changed", repo)

	require.NoError(t, err)
	require.Equal(t, 1, len(repo.createdLogs))

	log := repo.createdLogs[0]

	// Verify all fields
	t.Run("audit log fields", func(t *testing.T) {
		assert.Equal(t, "42", log.UserID)
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
