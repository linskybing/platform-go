package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLogAudit - Table-driven tests
func TestLogAudit(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
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
			userID:       "1",
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
			userID:       "2",
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
			userID:       "3",
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
			userID:       "4",
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
			userID:       "5",
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
			repo := &mockAuditRepo{}

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
			repo := &mockAuditRepo{}

			err := LogAudit("1", "127.0.0.1", "test", "action", "resource", "id",
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
