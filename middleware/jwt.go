package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/linskybing/platform-go/config"
)

// secret key to sign the JWT token, should come from config in real use
var jwtKey []byte

// Claims defines the custom JWT claims structure
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func Init() {
	jwtKey = []byte(config.JwtSecret)
}

// GenerateToken creates a JWT token for a given username and expiration duration
func GenerateToken(userID uint, username string, isAdmin bool, expireDuration time.Duration) (string, error) {
	claims := &Claims{
		UserID:   userID,
		Username: username,
		IsAdmin:  isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expireDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    config.Issuer,
		},
	}

	// Create a token with claims and sign it using HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Return the signed JWT token string
	return token.SignedString(jwtKey)
}

// ParseToken parses and validates a JWT token string and returns the claims
func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}

	// Parse the token with claims and a key function
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	// Check for parsing error or invalid token
	if err != nil || !token.Valid {
		return nil, err
	}

	// Return the claims if token is valid
	return claims, nil
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenStr := parts[1]

		claims, err := ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: " + err.Error()})
			c.Abort()
			return
		}

		c.Set("claims", claims)
		c.Next()
	}
}
