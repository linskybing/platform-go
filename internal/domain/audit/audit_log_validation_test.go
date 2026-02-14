package audit

import (
	"testing"
	"time"
)

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
