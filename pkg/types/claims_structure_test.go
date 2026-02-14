package types

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TestClaimsStructure verifies Claims struct initialization
func TestClaimsStructure(t *testing.T) {
	tests := []struct {
		name       string
		setupClaim func() Claims
		verify     func(Claims) bool
		scenario   string
	}{
		{
			name: "admin_claims",
			setupClaim: func() Claims {
				return Claims{
					UserID:   "1",
					Username: "admin",
					IsAdmin:  true,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.UserID == "1" && c.IsAdmin && c.Username == "admin"
			},
			scenario: "Admin user claims",
		},
		{
			name: "regular_user_claims",
			setupClaim: func() Claims {
				return Claims{
					UserID:   "100",
					Username: "john_doe",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(12 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.UserID == "100" && !c.IsAdmin && c.Username == "john_doe"
			},
			scenario: "Regular user claims",
		},
		{
			name: "service_account_claims",
			setupClaim: func() Claims {
				return Claims{
					UserID:   "9999",
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
					UserID:   "0",
					Username: "anonymous",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now()),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.UserID == "0"
			},
			scenario: "User with zero ID",
		},
		{
			name: "expired_claims",
			setupClaim: func() Claims {
				return Claims{
					UserID:   "200",
					Username: "expired_user",
					IsAdmin:  false,
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
						IssuedAt:  jwt.NewNumericDate(time.Now().Add(-25 * time.Hour)),
					},
				}
			},
			verify: func(c Claims) bool {
				return c.UserID == "200" && c.ExpiresAt != nil
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
