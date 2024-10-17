package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func getTokenAndUserID(c *gin.Context) (string, int, error) {
	// токен и user-id могут быть либо в хедерах, либо в параметрах запроса(в случае вебсокета)
	var err error
	token, userID := "", 0

	if c.Request.Header.Get("Authorization") != "" && c.Request.Header.Get("x-user-id") != "" {
		reqToken := c.Request.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return "", 0, fmt.Errorf("invalid authorization heder format")
		}

		token = splitToken[1]

		userIDRaw := c.Request.Header.Get("x-user-id")
		userID, err = strconv.Atoi(userIDRaw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id", "user-id": userIDRaw})
			return "", 0, fmt.Errorf("invalud user id %s: %w", userIDRaw, err)
		}
	} else if c.Query("token") != "" && c.Query("userId") != "" {
		token = c.Query("token")
		userID, err = strconv.Atoi(c.Query("userId"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id", "user-id": c.Param("user_id")})
			return "", 0, fmt.Errorf("invalid user id %s: %w", c.Query("user_id"), err)
		}
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token and userId are required"})
		return "", 0, fmt.Errorf("token and userId are required")
	}

	return token, userID, nil
}
