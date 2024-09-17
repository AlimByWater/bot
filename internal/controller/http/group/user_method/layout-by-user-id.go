package user_method

import (
	http2 "arimadj-helper/internal/controller/http"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type getUserLayout struct {
	layout layoutUC
}

func (gul getUserLayout) method() string {
	return http.MethodGet
}

func (gul getUserLayout) path() string {
	return "/layout/:userID"
}

// layoutByUserID обрабатывает запрос на получение макета пользователя
func (gul getUserLayout) layoutByUserID(c *gin.Context) {
	initiatorUserID, err := http2.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paramUserID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return

	}

	layout, err := gul.layout.GetUserLayout(c.Request.Context(), paramUserID, initiatorUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, layout)
}

// NewGetUserLayout создает новый обработчик для получения макета пользователя
func NewGetUserLayout(usecase layoutUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		gul := getUserLayout{layout: usecase}
		return gul.method(), gul.path(), gul.layoutByUserID
	}
}
