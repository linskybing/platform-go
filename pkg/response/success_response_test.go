package response

import "testing"

// TestSuccessResponseStructure tests SuccessResponse struct
func TestSuccessResponseStructure(t *testing.T) {
	tests := []struct {
		name     string
		response SuccessResponse
		verify   func(SuccessResponse) bool
		scenario string
	}{
		{
			name: "success_with_data",
			response: SuccessResponse{
				Code:    200,
				Message: "OK",
				Data:    map[string]string{"key": "value"},
			},
			verify: func(r SuccessResponse) bool {
				return r.Code == 200 && r.Message == "OK" && r.Data != nil
			},
			scenario: "Success response with data",
		},
		{
			name: "success_with_nil_data",
			response: SuccessResponse{
				Code:    201,
				Message: "Created",
				Data:    nil,
			},
			verify: func(r SuccessResponse) bool {
				return r.Code == 201 && r.Data == nil
			},
			scenario: "Success response with nil data",
		},
		{
			name: "success_with_array_data",
			response: SuccessResponse{
				Code:    200,
				Message: "Retrieved list",
				Data:    []interface{}{"item1", "item2", "item3"},
			},
			verify: func(r SuccessResponse) bool {
				return r.Code == 200 && len(r.Message) > 0
			},
			scenario: "Success response with array data",
		},
		{
			name: "success_with_complex_data",
			response: SuccessResponse{
				Code:    200,
				Message: "User data retrieved",
				Data: map[string]interface{}{
					"id":    1,
					"name":  "John",
					"email": "john@example.com",
					"roles": []string{"user", "admin"},
				},
			},
			verify: func(r SuccessResponse) bool {
				return r.Code == 200 && r.Data != nil
			},
			scenario: "Success response with complex nested data",
		},
		{
			name: "success_no_content",
			response: SuccessResponse{
				Code:    204,
				Message: "No Content",
				Data:    nil,
			},
			verify: func(r SuccessResponse) bool {
				return r.Code == 204
			},
			scenario: "Success response with no content",
		},
		{
			name: "success_with_string_data",
			response: SuccessResponse{
				Code:    200,
				Message: "String data",
				Data:    "Simple string response",
			},
			verify: func(r SuccessResponse) bool {
				return r.Code == 200 && r.Data != nil
			},
			scenario: "Success response with string data",
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
