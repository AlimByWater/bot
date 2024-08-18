package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type apiKey interface {
	GetApiKey() string
}

// NewGroup создает группу хендлеров
func NewGroup(ak apiKey, h ...func() (method string, path string, handlerFunc gin.HandlerFunc)) Api {
	return Api{
		apiKey:   ak,
		handlers: h,
	}
}

// Api ...
type Api struct {
	apiKey   apiKey
	handlers []func() (method string, path string, handlerFunc gin.HandlerFunc)
}

// Path ...
func (Api) Path() string {
	return "/api"
}

// Handlers ...
func (g Api) Handlers() []func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return g.handlers
}

// Auth миддлвейр авторизаций
func (g Api) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("x-api-key") != g.apiKey.GetApiKey() {
			//logger.Error("😈unauthorized access", slog.String("from IP", c.ClientIP()), slog.String("remote IP", c.RemoteIP()))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "😈 unauthorized"})
			c.Abort()
			return
		}
	}
}
