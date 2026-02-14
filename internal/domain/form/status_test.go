package form

import "testing"

// TestFormStatusConstants verifies form status constants
func TestFormStatusConstants(t *testing.T) {
	tests := []struct {
		name     string
		status   FormStatus
		expected string
	}{
		{
			name:     "pending_status",
			status:   FormStatusPending,
			expected: "Pending",
		},
		{
			name:     "processing_status",
			status:   FormStatusProcessing,
			expected: "Processing",
		},
		{
			name:     "completed_status",
			status:   FormStatusCompleted,
			expected: "Completed",
		},
		{
			name:     "rejected_status",
			status:   FormStatusRejected,
			expected: "Rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.status != FormStatus(tt.expected) {
				t.Errorf("expected %s, got %s", tt.expected, tt.status)
			}
		})
	}
}
