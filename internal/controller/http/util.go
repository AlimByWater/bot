package http

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func GetUserID(c *gin.Context) (int, error) {
	u, exist := c.Get("userID")
	if !exist {
		return 0, fmt.Errorf("user id not found in token")
	}
	userID, ok := u.(int)
	if !ok {
		return 0, fmt.Errorf("user id is not a number")
	}

	return userID, nil
}
