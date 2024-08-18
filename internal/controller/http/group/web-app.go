package api

import (
	"context"
	"github.com/gin-gonic/gin"
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
		//
		//if c.Request.Header.Get("Authorization") == "" || c.Request.Header.Get("x-user-id") == "" {
		//	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization and x-user-id headers are required"})
		//	return
		//}
		//
		//reqToken := c.Request.Header.Get("Authorization")
		//splitToken := strings.Split(reqToken, "Bearer ")
		//if len(splitToken) != 2 {
		//	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
		//	return
		//}
		//
		//token := splitToken[1]
		//
		//userIDRaw := c.Request.Header.Get("x-user-id")
		//userID, err := strconv.Atoi(userIDRaw)
		//if err != nil {
		//	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id", "user-id": userIDRaw})
		//	return
		//}
		//
		//valid, err := g.tokenChecker.CheckAccessTokenByUserID(c.Request.Context(), token, userID)
		//if err != nil {
		//	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		//	return
		//}
		//
		//if !valid {
		//	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		//	return
		//}
	}
}
