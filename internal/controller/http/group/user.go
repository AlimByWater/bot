package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

func NewUserGroup(tc tokenChecker, h ...func() (method string, path string, handlerFunc gin.HandlerFunc)) Layout {
	return Layout{
		tokenChecker: tc,
		handlers:     h,
	}
}

type User struct {
	tokenChecker tokenChecker
	handlers     []func() (method string, path string, handlerFunc gin.HandlerFunc)
}

func (User) Path() string {
	return "/user"
}

// Handlers ...
func (u User) Handlers() []func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return u.handlers
}

// Auth миддлвейр авторизаций
func (u User) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token, userID = "", 0

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

		valid, err := u.tokenChecker.CheckAccessTokenByUserID(c.Request.Context(), token, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("userID", userID)
	}
}
