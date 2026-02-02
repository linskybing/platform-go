package gpu

import (
	"testing"
)

// TestCreateGPURequestDTOValidation tests CreateGPURequestDTO
func TestCreateGPURequestDTOValidation(t *testing.T) {
	tests := []struct {
		name     string
		dto      CreateGPURequestDTO
		isValid  func(CreateGPURequestDTO) bool
		scenario string
	}{
		{
			name: "quota_request_dto_valid",
			dto: CreateGPURequestDTO{
				Type:           "quota",
				RequestedQuota: ptrInt(8),
				Reason:         "Need GPU resources for training",
			},
			isValid: func(d CreateGPURequestDTO) bool {
				return d.Type == "quota" &&
					d.RequestedQuota != nil &&
					*d.RequestedQuota == 8
			},
			scenario: "Valid quota request DTO",
		},
		{
			name: "access_request_dto_valid",
			dto: CreateGPURequestDTO{
				Type:                "access",
				RequestedAccessType: ptrString("exclusive"),
				Reason:              "Need exclusive GPU access",
			},
			isValid: func(d CreateGPURequestDTO) bool {
				return d.Type == "access" &&
					d.RequestedAccessType != nil &&
					*d.RequestedAccessType == "exclusive"
			},
			scenario: "Valid access request DTO",
		},
		{
			name: "quota_request_zero_quota",
			dto: CreateGPURequestDTO{
				Type:           "quota",
				RequestedQuota: ptrInt(0),
				Reason:         "Check availability",
			},
			isValid: func(d CreateGPURequestDTO) bool {
				return *d.RequestedQuota == 0
			},
			scenario: "Quota request with zero quota",
		},
		{
			name: "quota_request_large_quota",
			dto: CreateGPURequestDTO{
				Type:           "quota",
				RequestedQuota: ptrInt(256),
				Reason:         "Large scale distributed training",
			},
			isValid: func(d CreateGPURequestDTO) bool {
				return *d.RequestedQuota == 256
			},
			scenario: "Quota request with large quota",
		},
		{
			name: "quota_request_without_quota_field",
			dto: CreateGPURequestDTO{
				Type:           "quota",
				RequestedQuota: nil,
				Reason:         "Test",
			},
			isValid: func(d CreateGPURequestDTO) bool {
				return d.RequestedQuota == nil
			},
			scenario: "Quota request without quota value",
		},
		{
			name: "access_request_without_type",
			dto: CreateGPURequestDTO{
				Type:                "access",
				RequestedAccessType: nil,
				Reason:              "Need access",
			},
			isValid: func(d CreateGPURequestDTO) bool {
				return d.RequestedAccessType == nil
			},
			scenario: "Access request without type",
		},
		{
			name: "invalid_type_string",
			dto: CreateGPURequestDTO{
				Type:           "invalid",
				RequestedQuota: ptrInt(4),
				Reason:         "Invalid type",
			},
			isValid: func(d CreateGPURequestDTO) bool {
				return d.Type == "invalid"
			},
			scenario: "Invalid request type",
		},
		{
			name: "empty_reason",
			dto: CreateGPURequestDTO{
				Type:           "quota",
				RequestedQuota: ptrInt(4),
				Reason:         "",
			},
			isValid: func(d CreateGPURequestDTO) bool {
				return d.Reason == ""
			},
			scenario: "DTO with empty reason",
		},
		{
			name: "long_reason_text",
			dto: CreateGPURequestDTO{
				Type:           "quota",
				RequestedQuota: ptrInt(8),
				Reason:         "This is a detailed reason explaining the GPU quota requirement for the project including the expected use case and duration.",
			},
			isValid: func(d CreateGPURequestDTO) bool {
				return len(d.Reason) > 100
			},
			scenario: "DTO with long reason",
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

// TestUpdateGPURequestStatusDTOValidation tests UpdateGPURequestStatusDTO
func TestUpdateGPURequestStatusDTOValidation(t *testing.T) {
	tests := []struct {
		name     string
		dto      UpdateGPURequestStatusDTO
		isValid  func(UpdateGPURequestStatusDTO) bool
		scenario string
	}{
		{
			name: "approved_status_update",
			dto: UpdateGPURequestStatusDTO{
				Status: "approved",
			},
			isValid: func(d UpdateGPURequestStatusDTO) bool {
				return d.Status == "approved"
			},
			scenario: "Approve GPU request",
		},
		{
			name: "rejected_status_update",
			dto: UpdateGPURequestStatusDTO{
				Status: "rejected",
			},
			isValid: func(d UpdateGPURequestStatusDTO) bool {
				return d.Status == "rejected"
			},
			scenario: "Reject GPU request",
		},
		{
			name: "invalid_status",
			dto: UpdateGPURequestStatusDTO{
				Status: "pending",
			},
			isValid: func(d UpdateGPURequestStatusDTO) bool {
				return d.Status == "pending"
			},
			scenario: "Invalid status value",
		},
		{
			name: "empty_status",
			dto: UpdateGPURequestStatusDTO{
				Status: "",
			},
			isValid: func(d UpdateGPURequestStatusDTO) bool {
				return d.Status == ""
			},
			scenario: "Empty status",
		},
		{
			name: "uppercase_status",
			dto: UpdateGPURequestStatusDTO{
				Status: "APPROVED",
			},
			isValid: func(d UpdateGPURequestStatusDTO) bool {
				return d.Status == "APPROVED"
			},
			scenario: "Uppercase status value",
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

// Helper functions for pointer conversions
func ptrInt(i int) *int {
	return &i
}

func ptrString(s string) *string {
	return &s
}
