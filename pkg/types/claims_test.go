package types

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestClaimsStructure verifies Claims struct initialization
func TestClaimsStructure(t *testing.T) {
	tests := []struct {
		name     string
		setupClaim func() Claims
		verify   func(Claims) bool
		scenario string
	}{
		{
			name: "admin_claims",
			setupClaim: func() Claims {
				return Claims{
					UserID:   1,
					Username: "admin",
					IsAdmin:  true,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.UserID == 1 && c.IsAdmin && c.Username == "admin"
			},
			scenario: "Admin user claims",
		},
		{
			name: "regular_user_claims",
			setupClaim: func() Claims {
				return Claims{
					UserID:   100,
					Username: "john_doe",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.UserID == 100 && !c.IsAdmin && c.Username == "john_doe"
			},
			scenario: "Regular user claims",
		},
		{
			name: "service_account_claims",
			setupClaim: func() Claims {
				return Claims{
					UserID:   9999,
					Username: "service-account",
					IsAdmin:  true,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.IsAdmin && c.Username == "service-account"
			},
			scenario: "Service account claims",
		},
		{
			name: "user_zero_id",
			setupClaim: func() Claims {
				return Claims{
					UserID:   0,
					Username: "anonymous",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.UserID == 0
			},
			scenario: "User with zero ID",
		},
		{
			name: "expired_claims",
			setupClaim: func() Claims {
				return Claims{
					UserID:   200,
					Username: "expired_user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-25 * time.Hour)),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.UserID == 200 && c.RegisteredClaims.ExpiresAt != nil
			},
			scenario: "Expired claims",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim := tt.setupClaim()
			if !tt.verify(claim) {
				t.Errorf("verification failed for scenario: %s", tt.scenario)
			}
		})
	}
}

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
				UserID:   12345,
				Username: "user12345",
				IsAdmin:  false,
			},
			verify: func(c Claims) bool {
				return c.UserID > 0
			},
			scenario: "Valid positive user ID",
		},
		{
			name: "empty_username",
			claims: Claims{
				UserID:   1,
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
				UserID:   1,
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
				UserID:   2,
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
				UserID:   1,
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
				UserID:   100,
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

// TestClaimsRegisteredClaimsValidation tests JWT registered claims
func TestClaimsRegisteredClaimsValidation(t *testing.T) {
	tests := []struct {
		name     string
		setupClaim func() Claims
		verify   func(Claims) bool
		scenario string
	}{
		{
			name: "claims_with_issued_at",
			setupClaim: func() Claims {
				return Claims{
					UserID:   1,
					Username: "user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						IssuedAt: jwt.NewNumericDate(time.Now()),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.RegisteredClaims.IssuedAt != nil
			},
			scenario: "Claims with issued at time",
		},
		{
			name: "claims_with_expiry",
			setupClaim: func() Claims {
				return Claims{
					UserID:   2,
					Username: "user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.RegisteredClaims.ExpiresAt != nil
			},
			scenario: "Claims with expiry time",
		},
		{
			name: "claims_with_subject",
			setupClaim: func() Claims {
				return Claims{
					UserID:   3,
					Username: "user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "user_123",
					},
				}
			},
			verify: func(c Claims) bool {
				return c.RegisteredClaims.Subject == "user_123"
			},
			scenario: "Claims with subject",
		},
		{
			name: "claims_with_issuer",
			setupClaim: func() Claims {
				return Claims{
					UserID:   4,
					Username: "user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer: "auth-service",
					},
				}
			},
			verify: func(c Claims) bool {
				return c.RegisteredClaims.Issuer == "auth-service"
			},
			scenario: "Claims with issuer",
		},
		{
			name: "claims_with_audience",
			setupClaim: func() Claims {
				return Claims{
					UserID:   5,
					Username: "user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						Audience: jwt.ClaimStrings{"api", "web"},
					},
				}
			},
			verify: func(c Claims) bool {
				return len(c.RegisteredClaims.Audience) > 0
			},
			scenario: "Claims with audience",
		},
		{
			name: "claims_with_not_before",
			setupClaim: func() Claims {
				return Claims{
					UserID:   6,
					Username: "user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						NotBefore: jwt.NewNumericDate(time.Now()),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.RegisteredClaims.NotBefore != nil
			},
			scenario: "Claims with not before",
		},
		{
			name: "claims_with_all_registered_fields",
			setupClaim: func() Claims {
				now := time.Now()
				return Claims{
					UserID:   7,
					Username: "admin",
					IsAdmin:  true,
					RegisteredClaims: jwt.RegisteredClaims{
						Subject:   "user_admin",
						Issuer:    "auth-service",
						Audience:  jwt.ClaimStrings{"api", "admin"},
						IssuedAt:  jwt.NewNumericDate(now),
						NotBefore: jwt.NewNumericDate(now),
						ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.UserID == 7 && c.IsAdmin && c.RegisteredClaims.IssuedAt != nil &&
					c.RegisteredClaims.ExpiresAt != nil && len(c.RegisteredClaims.Audience) > 0
			},
			scenario: "Claims with all registered fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim := tt.setupClaim()
			if !tt.verify(claim) {
				t.Errorf("registered claims validation failed for scenario: %s", tt.scenario)
			}
		})
	}
}

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
				UserID:   4294967295, // Max uint32
				Username: "maxuser",
				IsAdmin:  false,
			},
			verify: func(c Claims) bool {
				return c.UserID == 4294967295
			},
			scenario: "Maximum uint user ID",
		},
		{
			name: "min_user_id",
			claims: Claims{
				UserID:   0,
				Username: "minuser",
				IsAdmin:  false,
			},
			verify: func(c Claims) bool {
				return c.UserID == 0
			},
			scenario: "Minimum user ID (zero)",
		},
		{
			name: "single_char_username",
			claims: Claims{
				UserID:   1,
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
				UserID:   1,
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
