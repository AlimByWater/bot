package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
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
	return "/users"
}

// Handlers ...
func (u User) Handlers() []func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return u.handlers
}

// Auth миддлвейр авторизаций
func (u User) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, userID, err := getTokenAndUserID(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
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
