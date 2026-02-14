package types

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestClaimsRegisteredClaimsValidation tests JWT registered claims
func TestClaimsRegisteredClaimsValidation(t *testing.T) {
	tests := []struct {
		name       string
		setupClaim func() Claims
		verify     func(Claims) bool
		scenario   string
	}{
		{
			name: "claims_with_issued_at",
			setupClaim: func() Claims {
				return Claims{
					UserID:   "1",
					Username: "user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						IssuedAt: jwt.NewNumericDate(time.Now()),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.IssuedAt != nil
			},
			scenario: "Claims with issued at time",
		},
		{
			name: "claims_with_expiry",
			setupClaim: func() Claims {
				return Claims{
					UserID:   "2",
					Username: "user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.ExpiresAt != nil
			},
			scenario: "Claims with expiry time",
		},
		{
			name: "claims_with_subject",
			setupClaim: func() Claims {
				return Claims{
					UserID:   "3",
					Username: "user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "user_123",
					},
				}
			},
			verify: func(c Claims) bool {
				return c.Subject == "user_123"
			},
			scenario: "Claims with subject",
		},
		{
			name: "claims_with_issuer",
			setupClaim: func() Claims {
				return Claims{
					UserID:   "4",
					Username: "user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer: "auth-service",
					},
				}
			},
			verify: func(c Claims) bool {
				return c.Issuer == "auth-service"
			},
			scenario: "Claims with issuer",
		},
		{
			name: "claims_with_audience",
			setupClaim: func() Claims {
				return Claims{
					UserID:   "5",
					Username: "user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						Audience: jwt.ClaimStrings{"api", "web"},
					},
				}
			},
			verify: func(c Claims) bool {
				return len(c.Audience) > 0
			},
			scenario: "Claims with audience",
		},
		{
			name: "claims_with_not_before",
			setupClaim: func() Claims {
				return Claims{
					UserID:   "6",
					Username: "user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						NotBefore: jwt.NewNumericDate(time.Now()),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.NotBefore != nil
			},
			scenario: "Claims with not before",
		},
		{
			name: "claims_with_all_registered_fields",
			setupClaim: func() Claims {
				now := time.Now()
				return Claims{
					UserID:   "7",
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
				return c.UserID == "7" && c.IsAdmin && c.IssuedAt != nil &&
					c.ExpiresAt != nil && len(c.Audience) > 0
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
