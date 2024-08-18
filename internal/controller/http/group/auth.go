package api

import (
	"github.com/gin-gonic/gin"
)

type authUC interface {
	//CheckAccessTokenByUserID(ctx context.Context, token string, userID int) (bool, error)
}

// NewAuthGroup создает группу хендлеров
func NewAuthGroup(tc authUC, h ...func() (method string, path string, handlerFunc gin.HandlerFunc)) Auth {
	return Auth{
		authUC:   tc,
		handlers: h,
	}
}

type Auth struct {
	authUC   authUC
	handlers []func() (method string, path string, handlerFunc gin.HandlerFunc)
}

// Path ...
func (Auth) Path() string {
	return "/auth/v1/provider"
}

// Handlers ...
func (g Auth) Handlers() []func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return g.handlers
}

// Auth миддлвейр авторизаций
func (g Auth) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
	}
}
