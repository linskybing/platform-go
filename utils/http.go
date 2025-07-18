package utils

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParseIDParam(c *gin.Context, param string) (uint, error) {
	idStr := c.Param(param)
	idUint64, err := strconv.ParseUint(idStr, 10, 64)
	return uint(idUint64), err
}

func ParseQueryUintParam(c *gin.Context, param string) (uint, error) {
	valStr := c.Query(param)
	if valStr == "" {
		return 0, errors.New("empty parameter")
	}
	valUint64, err := strconv.ParseUint(valStr, 10, 64)
	return uint(valUint64), err
}
