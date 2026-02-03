package response

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/internal/domain/group"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type TokenResponse struct {
	Token    string `json:"token"`
	UID      uint   `json:"user_id"`
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_super_admin"`
}

type GroupResponse struct {
	Message string      `json:"message"`
	Group   group.Group `json:"group"`
}

type SuccessResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// Success returns a success response with data
func Success(c *gin.Context, data interface{}, message string) {
	c.JSON(200, SuccessResponse{
		Code:    200,
		Message: message,
		Data:    data,
	})
}

// Error returns an error response
func Error(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, ErrorResponse{
		Error: message,
	})
}

// Unauthorized returns a 401 unauthorized response
func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, message)
}
