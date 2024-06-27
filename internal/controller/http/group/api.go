package api

import (
	"github.com/gin-gonic/gin"
)

type password interface {
	GetPassword(string) (string, error)
}

// NewGroup создает группу хендлеров
func NewGroup(password password, h ...func() (method string, path string, handlerFunc gin.HandlerFunc)) Api {
	return Api{
		password: password,
		handlers: h,
	}
}

// Api ...
type Api struct {
	password password
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
		//username, pwd, ok := c.Request.BasicAuth()
		//if !ok {
		//	c.Header("WWW-Authenticate", "Basic realm=Restricted")
		//	c.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		//storagePwd, err := g.password.GetPassword(username)
		//if err != nil {
		//	_ = c.AbortWithError(http.StatusInternalServerError, err)
		//	return
		//}
		//if storagePwd != hex.EncodeToString(sha256.New().Sum([]byte(pwd))) {
		//	c.Header("WWW-Authenticate", "Basic realm=Restricted")
		//	c.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
	}
}
