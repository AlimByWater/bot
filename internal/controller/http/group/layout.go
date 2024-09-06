package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// NewLayoutGroup создает группу хендлеров
func NewLayoutGroup(tc tokenChecker, h ...func() (method string, path string, handlerFunc gin.HandlerFunc)) Layout {
	return Layout{
		tokenChecker: tc,
		handlers:     h,
	}
}

type Layout struct {
	tokenChecker tokenChecker
	handlers     []func() (method string, path string, handlerFunc gin.HandlerFunc)
}

// Path ...
func (Layout) Path() string {
	return "/layout"
}

// Handlers ...
func (g Layout) Handlers() []func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return g.handlers
}

// Auth миддлвейр авторизаций
func (g Layout) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, userID, err := getTokenAndUserID(c)

		valid, err := g.tokenChecker.CheckAccessTokenByUserID(c.Request.Context(), token, userID)
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
