package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// NewSongGroup создает группу хендлеров
func NewSongGroup(tokenChecker tokenChecker, h ...func() (method string, path string, handlerFunc gin.HandlerFunc)) Song {
	return Song{
		tokenChecker: tokenChecker,
		handlers:     h,
	}
}

type Song struct {
	tokenChecker tokenChecker
	handlers     []func() (method string, path string, handlerFunc gin.HandlerFunc)
}

// Path ...
func (Song) Path() string {
	return "/song"
}

// Handlers ...
func (s Song) Handlers() []func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return s.handlers
}

// Auth миддат авторизаций
func (s Song) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, userID, err := getTokenAndUserID(c)

		valid, err := s.tokenChecker.CheckAccessTokenByUserID(c.Request.Context(), token, userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("userID", userID)
	}
}
