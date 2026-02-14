package response

import "testing"

// TestErrorResponseStructure tests ErrorResponse struct
func TestErrorResponseStructure(t *testing.T) {
	tests := []struct {
		name     string
		response ErrorResponse
		verify   func(ErrorResponse) bool
		scenario string
	}{
		{
			name: "simple_error",
			response: ErrorResponse{
				Error: "Invalid request",
			},
			verify: func(r ErrorResponse) bool {
				return r.Error == "Invalid request"
			},
			scenario: "Simple error response",
		},
		{
			name: "empty_error",
			response: ErrorResponse{
				Error: "",
			},
			verify: func(r ErrorResponse) bool {
				return r.Error == ""
			},
			scenario: "Empty error message",
		},
		{
			name: "long_error_message",
			response: ErrorResponse{
				Error: "The request failed due to multiple validation errors: invalid email format, missing required field 'name', and invalid project ID reference",
			},
			verify: func(r ErrorResponse) bool {
				return len(r.Error) > 50
			},
			scenario: "Long error message",
		},
		{
			name: "error_with_special_chars",
			response: ErrorResponse{
				Error: "Error: {field} = 'value' is invalid!",
			},
			verify: func(r ErrorResponse) bool {
				return len(r.Error) > 0
			},
			scenario: "Error with special characters",
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
