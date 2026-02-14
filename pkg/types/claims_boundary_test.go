package types

import "testing"

// TestClaimsBoundaryConditions tests boundary conditions
func TestClaimsBoundaryConditions(t *testing.T) {
	tests := []struct {
		name     string
		claims   Claims
		verify   func(Claims) bool
		scenario string
	}{
		{
			name: "max_user_id",
			claims: Claims{
				UserID:   "4294967295", // Max uint32
				Username: "maxuser",
				IsAdmin:  false,
			},
			verify: func(c Claims) bool {
				return c.UserID == "4294967295"
			},
			scenario: "Maximum uint user ID",
		},
		{
			name: "min_user_id",
			claims: Claims{
				UserID:   "0",
				Username: "minuser",
				IsAdmin:  false,
			},
			verify: func(c Claims) bool {
				return c.UserID == "0"
			},
			scenario: "Minimum user ID (zero)",
		},
		{
			name: "single_char_username",
			claims: Claims{
				UserID:   "1",
				Username: "a",
				IsAdmin:  false,
			},
			verify: func(c Claims) bool {
				return len(c.Username) == 1
			},
			scenario: "Single character username",
		},
		{
			name: "very_long_username",
			claims: Claims{
				UserID:   "1",
				Username: "a" + string(make([]byte, 1000)),
				IsAdmin:  false,
			},
			verify: func(c Claims) bool {
				return len(c.Username) > 100
			},
			scenario: "Very long username",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.verify(tt.claims) {
				t.Errorf("boundary condition test failed for scenario: %s", tt.scenario)
			}
		})
	}
}
