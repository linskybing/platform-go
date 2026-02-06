package utils

import (
	"errors"

	"github.com/gin-gonic/gin"
)

var (
	ErrEmptyParameter = errors.New("empty parameter")
)

func ParseIDParam(c *gin.Context, param string) (string, error) {
	idStr := c.Param(param)
	if idStr == "" {
		return "", errors.New("empty parameter")
	}
	return idStr, nil
}

func ParseQueryIDParam(c *gin.Context, param string) (string, error) {
	valStr := c.Query(param)
	if valStr == "" {
		return "", ErrEmptyParameter
	}
	return valStr, nil
}
