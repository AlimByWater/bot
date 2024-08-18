package web_app_methods

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type webSocket struct {
	users usersUC
}

func (ws webSocket) method() string {
	return http.MethodGet
}

func (ws webSocket) path() string {
	return "/ws"
}

func (ws webSocket) ws(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for {
		info, err := ws.users.WebsocketInfo()
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			break
		}

		err = conn.WriteJSON(info)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			break
		}

		time.Sleep(5 * time.Second)
	}
}

func NewWebsocketEvent(usecase usersUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		wae := webSocket{users: usecase}
		return wae.method(), wae.path(), wae.ws
	}
}
