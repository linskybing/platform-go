package response

import (
	"testing"

	"github.com/linskybing/platform-go/internal/domain/group"
)

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

// TestTokenResponseStructure tests TokenResponse struct
func TestTokenResponseStructure(t *testing.T) {
	tests := []struct {
		name     string
		response TokenResponse
		verify   func(TokenResponse) bool
		scenario string
	}{
		{
			name: "admin_token",
			response: TokenResponse{
				Token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
				UID:      "1",
				Username: "admin",
				IsAdmin:  true,
			},
			verify: func(r TokenResponse) bool {
				return r.IsAdmin && r.UID == "1"
			},
			scenario: "Admin user token",
		},
		{
			name: "regular_user_token",
			response: TokenResponse{
				Token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
				UID:      "100",
				Username: "john_doe",
				IsAdmin:  false,
			},
			verify: func(r TokenResponse) bool {
				return !r.IsAdmin && r.UID == "100"
			},
			scenario: "Regular user token",
		},
		{
			name: "token_with_long_jwt",
			response: TokenResponse{
				Token:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
				UID:      "200",
				Username: "user200",
				IsAdmin:  false,
			},
			verify: func(r TokenResponse) bool {
				return len(r.Token) > 50
			},
			scenario: "Token with long JWT string",
		},
		{
			name: "service_account_token",
			response: TokenResponse{
				Token:    "sa_token_12345...",
				UID:      "999",
				Username: "service-account",
				IsAdmin:  true,
			},
			verify: func(r TokenResponse) bool {
				return r.UID == "999" && r.IsAdmin
			},
			scenario: "Service account token",
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

// TestGroupResponseStructure tests GroupResponse struct
func TestGroupResponseStructure(t *testing.T) {
	tests := []struct {
		name     string
		response GroupResponse
		verify   func(GroupResponse) bool
		scenario string
	}{
		{
			name: "group_response_with_data",
			response: GroupResponse{
				Message: "Group retrieved successfully",
				Group: group.Group{
					GroupName: "TestGroup",
				},
			},
			verify: func(r GroupResponse) bool {
				return r.Message != "" && r.Group.GroupName == "TestGroup"
			},
			scenario: "Group response with group data",
		},
		{
			name: "group_response_empty_name",
			response: GroupResponse{
				Message: "Group created",
				Group: group.Group{
					GroupName: "",
				},
			},
			verify: func(r GroupResponse) bool {
				return r.Message != ""
			},
			scenario: "Group response with empty name",
		},
		{
			name: "group_response_long_message",
			response: GroupResponse{
				Message: "Group operation completed successfully with all validations passed",
				Group: group.Group{
					GroupName: "LongMessageGroup",
				},
			},
			verify: func(r GroupResponse) bool {
				return len(r.Message) > 30
			},
			scenario: "Group response with long message",
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

// TestResponseCodeConstants tests HTTP status codes
func TestResponseCodeConstants(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		verify   func(int) bool
		scenario string
	}{
		{
			name:     "ok_response",
			code:     200,
			verify:   func(c int) bool { return c == 200 },
			scenario: "HTTP 200 OK",
		},
		{
			name:     "created_response",
			code:     201,
			verify:   func(c int) bool { return c == 201 },
			scenario: "HTTP 201 Created",
		},
		{
			name:     "no_content_response",
			code:     204,
			verify:   func(c int) bool { return c == 204 },
			scenario: "HTTP 204 No Content",
		},
		{
			name:     "bad_request",
			code:     400,
			verify:   func(c int) bool { return c == 400 },
			scenario: "HTTP 400 Bad Request",
		},
		{
			name:     "unauthorized",
			code:     401,
			verify:   func(c int) bool { return c == 401 },
			scenario: "HTTP 401 Unauthorized",
		},
		{
			name:     "forbidden",
			code:     403,
			verify:   func(c int) bool { return c == 403 },
			scenario: "HTTP 403 Forbidden",
		},
		{
			name:     "not_found",
			code:     404,
			verify:   func(c int) bool { return c == 404 },
			scenario: "HTTP 404 Not Found",
		},
		{
			name:     "internal_error",
			code:     500,
			verify:   func(c int) bool { return c == 500 },
			scenario: "HTTP 500 Internal Server Error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.verify(tt.code) {
				t.Errorf("code verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}
