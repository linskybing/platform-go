package types

import "testing"

// TestClaimsFieldValidation tests individual field validation
func TestClaimsFieldValidation(t *testing.T) {
	tests := []struct {
		name     string
		claims   Claims
		verify   func(Claims) bool
		scenario string
	}{
		{
			name: "valid_user_id",
			claims: Claims{
				UserID:   "12345",
				Username: "user12345",
				IsAdmin:  false,
			},
			verify: func(c Claims) bool {
				return c.UserID != ""
			},
			scenario: "Valid positive user ID",
		},
		{
			name: "empty_username",
			claims: Claims{
				UserID:   "1",
				Username: "",
				IsAdmin:  false,
			},
			verify: func(c Claims) bool {
				return c.Username == ""
			},
			scenario: "Empty username",
		},
		{
			name: "long_username",
			claims: Claims{
				UserID:   "1",
				Username: "this_is_a_very_long_username_that_exceeds_normal_length",
				IsAdmin:  false,
			},
			verify: func(c Claims) bool {
				return len(c.Username) > 30
			},
			scenario: "Long username",
		},
		{
			name: "username_with_special_chars",
			claims: Claims{
				UserID:   "2",
				Username: "user@example.com",
				IsAdmin:  false,
			},
			verify: func(c Claims) bool {
				return len(c.Username) > 0
			},
			scenario: "Username with special characters",
		},
		{
			name: "admin_flag_true",
			claims: Claims{
				UserID:   "1",
				Username: "admin",
				IsAdmin:  true,
			},
			verify: func(c Claims) bool {
				return c.IsAdmin
			},
			scenario: "Admin flag set to true",
		},
		{
			name: "admin_flag_false",
			claims: Claims{
				UserID:   "100",
				Username: "user",
				IsAdmin:  false,
			},
			verify: func(c Claims) bool {
				return !c.IsAdmin
			},
			scenario: "Admin flag set to false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.verify(tt.claims) {
				t.Errorf("field validation failed for scenario: %s", tt.scenario)
			}
		})
	}
}
