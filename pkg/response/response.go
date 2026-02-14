package response

import (
	"github.com/gin-gonic/gin"
)

// StandardResponse defines the unified structure for all API responses.
type StandardResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ErrorResponse keeps legacy error payload compatibility.
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse keeps legacy success payload compatibility.
type SuccessResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// MessageResponse is a lightweight message-only response.
type MessageResponse struct {
	Message string `json:"message"`
}

type TokenData struct {
	Token    string `json:"token"`
	UID      string `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_super_admin"`
}

// TokenResponse keeps legacy auth response compatibility.
type TokenResponse struct {
	Token    string `json:"token"`
	UID      string `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_super_admin"`
}

// Success returns a success response with data
func Success(c *gin.Context, data interface{}, message string) {
	SuccessWithStatus(c, 200, data, message)
}

// SuccessWithStatus returns a success response with a specific status code
func SuccessWithStatus(c *gin.Context, statusCode int, data interface{}, message string) {
	c.JSON(statusCode, StandardResponse{
		Code:    statusCode,
		Message: message,
		Data:    data,
	})
}

// Error returns an error response with a standardized structure
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, StandardResponse{
		Code:    statusCode,
		Message: message,
		Data:    nil,
	})
}

// Unauthorized returns a 401 unauthorized response
func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, message)
}
