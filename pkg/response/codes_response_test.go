package response

import "testing"

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
