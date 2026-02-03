package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/linskybing/platform-go/pkg/response"
	"github.com/linskybing/platform-go/pkg/utils"
)

// AuthStatusHandler returns the status of the user's token (valid/expired)
func AuthStatusHandler(c *gin.Context) {
	uid, err := utils.GetUserIDFromContext(c)
	if err != nil {
		response.Unauthorized(c, "token expired")
		return
	}
	response.Success(c, map[string]interface{}{
		"status":  "valid",
		"user_id": uid,
	}, "Token is valid")
}
