package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func getTokenAndUserID(c *gin.Context) (token string, userID int, err error) {
	// токен и user-id могут быть либо в хедерах, либо в параметрах запроса(в случае вебсокета)
	if c.Request.Header.Get("Authorization") != "" && c.Request.Header.Get("x-user-id") != "" {
		var err error
		reqToken := c.Request.Header.Get("Authorization")
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			return
		}

		token = splitToken[1]

		userIDRaw := c.Request.Header.Get("x-user-id")
		userID, err = strconv.Atoi(userIDRaw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id", "user-id": userIDRaw})
			return
		}
	} else if c.Param("token") != "" && c.Param("userId") != "" {
		var err error
		token = c.Param("token")
		userID, err = strconv.Atoi(c.Param("userId"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id", "user-id": c.Param("user_id")})
			return
		}
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token and userId are required"})
		return
	}

	if token == "test-token" && userID == 5 {
		return
	}

	return
}
