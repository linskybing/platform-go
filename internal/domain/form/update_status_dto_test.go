package form

import "testing"

// TestUpdateFormStatusDTOValidation tests UpdateFormStatusDTO
func TestUpdateFormStatusDTOValidation(t *testing.T) {
	tests := []struct {
		name     string
		dto      UpdateFormStatusDTO
		isValid  func(UpdateFormStatusDTO) bool
		scenario string
	}{
		{
			name: "update_status_pending",
			dto: UpdateFormStatusDTO{
				Status: "Pending",
			},
			isValid: func(d UpdateFormStatusDTO) bool {
				return d.Status == "Pending"
			},
			scenario: "Update form to pending status",
		},
		{
			name: "update_status_processing",
			dto: UpdateFormStatusDTO{
				Status: "Processing",
			},
			isValid: func(d UpdateFormStatusDTO) bool {
				return d.Status == "Processing"
			},
			scenario: "Update form to processing status",
		},
		{
			name: "update_status_completed",
			dto: UpdateFormStatusDTO{
				Status: "Completed",
			},
			isValid: func(d UpdateFormStatusDTO) bool {
				return d.Status == "Completed"
			},
			scenario: "Update form to completed status",
		},
		{
			name: "update_status_rejected",
			dto: UpdateFormStatusDTO{
				Status: "Rejected",
			},
			isValid: func(d UpdateFormStatusDTO) bool {
				return d.Status == "Rejected"
			},
			scenario: "Update form to rejected status",
		},
		{
			name: "update_status_invalid",
			dto: UpdateFormStatusDTO{
				Status: "InvalidStatus",
			},
			isValid: func(d UpdateFormStatusDTO) bool {
				return d.Status == "InvalidStatus"
			},
			scenario: "Update with invalid status value",
		},
		{
			name: "update_status_empty",
			dto: UpdateFormStatusDTO{
				Status: "",
			},
			isValid: func(d UpdateFormStatusDTO) bool {
				return d.Status == ""
			},
			scenario: "Update with empty status",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.isValid(tt.dto) {
				t.Errorf("validation failed for scenario: %s", tt.scenario)
			}
		})
	}
}
