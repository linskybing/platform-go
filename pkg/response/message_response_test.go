package response

import "testing"

// TestMessageResponseStructure tests MessageResponse struct
func TestMessageResponseStructure(t *testing.T) {
	tests := []struct {
		name     string
		response MessageResponse
		verify   func(MessageResponse) bool
		scenario string
	}{
		{
			name: "success_message",
			response: MessageResponse{
				Message: "Operation completed successfully",
			},
			verify: func(r MessageResponse) bool {
				return r.Message == "Operation completed successfully"
			},
			scenario: "Success message",
		},
		{
			name: "info_message",
			response: MessageResponse{
				Message: "Resource created",
			},
			verify: func(r MessageResponse) bool {
				return r.Message != ""
			},
			scenario: "Info message",
		},
		{
			name: "warning_message",
			response: MessageResponse{
				Message: "Warning: action may have side effects",
			},
			verify: func(r MessageResponse) bool {
				return len(r.Message) > 0
			},
			scenario: "Warning message",
		},
		{
			name: "empty_message",
			response: MessageResponse{
				Message: "",
			},
			verify: func(r MessageResponse) bool {
				return r.Message == ""
			},
			scenario: "Empty message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.verify(tt.response) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}
