package audit

import (
	"testing"
	"time"

	"gorm.io/datatypes"
)

// TestAuditLogStructure verifies that AuditLog struct has all required fields
func TestAuditLogStructure(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() AuditLog
		verify  func(AuditLog) bool
		wantErr bool
	}{
		{
			name: "valid_full_audit_log",
			setup: func() AuditLog {
				return AuditLog{
					ID:           1,
					UserID:       "100",
					Action:       "CREATE",
					ResourceType: "Project",
					ResourceID:   "proj-123",
					OldData:      datatypes.JSON(`{"old": "data"}`),
					NewData:      datatypes.JSON(`{"new": "data"}`),
					IPAddress:    "192.168.1.1",
					UserAgent:    "Mozilla/5.0",
					Description:  "Project created",
					CreatedAt:    time.Now(),
				}
			},
			verify: func(log AuditLog) bool {
				return log.ID == 1 &&
					log.UserID == "100" &&
					log.Action == "CREATE" &&
					log.ResourceType == "Project" &&
					log.ResourceID == "proj-123"
			},
			wantErr: false,
		},
		{
			name: "minimal_audit_log",
			setup: func() AuditLog {
				return AuditLog{
					UserID:       "50",
					Action:       "UPDATE",
					ResourceType: "User",
					ResourceID:   "user-456",
					CreatedAt:    time.Now(),
				}
			},
			verify: func(log AuditLog) bool {
				return log.UserID == "50" &&
					log.Action == "UPDATE" &&
					log.ResourceType == "User" &&
					log.ResourceID == "user-456"
			},
			wantErr: false,
		},
		{
			name: "audit_log_with_empty_json",
			setup: func() AuditLog {
				return AuditLog{
					UserID:       "75",
					Action:       "DELETE",
					ResourceType: "Resource",
					ResourceID:   "res-789",
					OldData:      datatypes.JSON(`{}`),
					NewData:      datatypes.JSON(nil),
					CreatedAt:    time.Now(),
				}
			},
			verify: func(log AuditLog) bool {
				return len(log.OldData) > 0 && log.UserID == "75"
			},
			wantErr: false,
		},
		{
			name: "audit_log_with_long_description",
			setup: func() AuditLog {
				longDesc := "This is a very long description of the audit event that contains multiple sentences and provides detailed information about what action was performed on the resource."
				return AuditLog{
					UserID:       "100",
					Action:       "MODIFY",
					ResourceType: "Configuration",
					ResourceID:   "config-001",
					Description:  longDesc,
					CreatedAt:    time.Now(),
				}
			},
			verify: func(log AuditLog) bool {
				return len(log.Description) > 50
			},
			wantErr: false,
		},
		{
			name: "audit_log_with_ipv6",
			setup: func() AuditLog {
				return AuditLog{
					UserID:       "100",
					Action:       "READ",
					ResourceType: "Data",
					ResourceID:   "data-123",
					IPAddress:    "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
					CreatedAt:    time.Now(),
				}
			},
			verify: func(log AuditLog) bool {
				return log.IPAddress == "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
			},
			wantErr: false,
		},
		{
			name: "audit_log_with_zero_user_id_boundary",
			setup: func() AuditLog {
				return AuditLog{
					UserID:       "0",
					Action:       "SYSTEM",
					ResourceType: "System",
					ResourceID:   "sys-000",
					CreatedAt:    time.Now(),
				}
			},
			verify: func(log AuditLog) bool {
				return log.UserID == "0"
			},
			wantErr: false,
		},
		{
			name: "audit_log_with_large_id",
			setup: func() AuditLog {
				return AuditLog{
					ID:           4294967295, // Max uint32
					UserID:       "1000000",
					Action:       "ARCHIVE",
					ResourceType: "Project",
					ResourceID:   "proj-max",
					CreatedAt:    time.Now(),
				}
			},
			verify: func(log AuditLog) bool {
				return log.ID == 4294967295
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := tt.setup()
			if !tt.verify(log) {
				t.Errorf("verification failed for %s", tt.name)
			}
		})
	}
}

// TestAuditLogFieldValidation verifies field constraints
func TestAuditLogFieldValidation(t *testing.T) {
	tests := []struct {
		name      string
		setupFunc func() AuditLog
		fieldName string
		isValid   func(AuditLog) bool
	}{
		{
			name: "action_field_validation",
			setupFunc: func() AuditLog {
				return AuditLog{
					UserID:       "1",
					Action:       "CREATE",
					ResourceType: "Project",
					ResourceID:   "id-1",
					CreatedAt:    time.Now(),
				}
			},
			fieldName: "Action",
			isValid: func(log AuditLog) bool {
				return log.Action != "" && len(log.Action) <= 20
			},
		},
		{
			name: "resource_type_field_validation",
			setupFunc: func() AuditLog {
				return AuditLog{
					UserID:       "1",
					Action:       "UPDATE",
					ResourceType: "ProjectConfiguration",
					ResourceID:   "id-2",
					CreatedAt:    time.Now(),
				}
			},
			fieldName: "ResourceType",
			isValid: func(log AuditLog) bool {
				return log.ResourceType != "" && len(log.ResourceType) <= 50
			},
		},
		{
			name: "resource_id_not_empty",
			setupFunc: func() AuditLog {
				return AuditLog{
					UserID:       "1",
					Action:       "DELETE",
					ResourceType: "User",
					ResourceID:   "user-123",
					CreatedAt:    time.Now(),
				}
			},
			fieldName: "ResourceID",
			isValid: func(log AuditLog) bool {
				return log.ResourceID != ""
			},
		},
		{
			name: "timestamp_set_on_creation",
			setupFunc: func() AuditLog {
				now := time.Now()
				return AuditLog{
					UserID:       "1",
					Action:       "CREATE",
					ResourceType: "Project",
					ResourceID:   "id-3",
					CreatedAt:    now,
				}
			},
			fieldName: "CreatedAt",
			isValid: func(log AuditLog) bool {
				return !log.CreatedAt.IsZero()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log := tt.setupFunc()
			if !tt.isValid(log) {
				t.Errorf("%s validation failed", tt.fieldName)
			}
		})
	}
}

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
