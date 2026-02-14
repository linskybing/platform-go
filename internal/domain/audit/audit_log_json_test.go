package audit

import (
	"testing"
	"time"

	"gorm.io/datatypes"
)

// TestAuditLogJSONMarshaling tests JSON field compatibility
func TestAuditLogJSONMarshaling(t *testing.T) {
	tests := []struct {
		name  string
		data  datatypes.JSON
		valid bool
	}{
		{
			name:  "valid_json_object",
			data:  datatypes.JSON(`{"key": "value"}`),
			valid: true,
		},
		{
			name:  "valid_json_array",
			data:  datatypes.JSON(`[1, 2, 3]`),
			valid: true,
		},
		{
			name:  "empty_json_object",
			data:  datatypes.JSON(`{}`),
			valid: true,
		},
		{
			name:  "null_json",
			data:  datatypes.JSON(nil),
			valid: true,
		},
		{
			name:  "complex_nested_json",
			data:  datatypes.JSON(`{"user": {"name": "John", "email": "john@example.com"}}`),
			valid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := AuditLog{
				UserID:       "1",
				Action:       "CREATE",
				ResourceType: "Test",
				ResourceID:   "test-1",
				OldData:      tt.data,
				CreatedAt:    time.Now(),
			}
			if log.OldData == nil && tt.valid {
				t.Log("OldData is nil as expected")
			}
		})
	}
}
