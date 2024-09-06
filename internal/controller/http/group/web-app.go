package api

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type tokenChecker interface {
	CheckAccessTokenByUserID(ctx context.Context, token string, userID int) (bool, error)
}

// NewWebAppGroup создает группу хендлеров
func NewWebAppGroup(tc tokenChecker, h ...func() (method string, path string, handlerFunc gin.HandlerFunc)) WebApp {
	return WebApp{
		tokenChecker: tc,
		handlers:     h,
	}
}

type WebApp struct {
	tokenChecker tokenChecker
	handlers     []func() (method string, path string, handlerFunc gin.HandlerFunc)
}

// Path ...
func (WebApp) Path() string {
	return "/web-app"
}

// Handlers ...
func (g WebApp) Handlers() []func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return g.handlers
}

// Auth миддлвейр авторизаций
func (g WebApp) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, userID, err := getTokenAndUserID(c)

		fmt.Printf("TOKEN AND USERID %v %v\n", token, userID)
		valid, err := g.tokenChecker.CheckAccessTokenByUserID(c.Request.Context(), token, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
	}
}
