package response

import "testing"

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
