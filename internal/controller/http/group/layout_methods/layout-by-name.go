package layout_methods

import (
	http2 "arimadj-helper/internal/controller/http"
	"github.com/gin-gonic/gin"
	"net/http"
)

type layoutByName struct {
	layout layoutUC
}

func (gul layoutByName) method() string {
	return http.MethodGet
}

func (gul layoutByName) path() string {
	return "/name/:name"
}

// layoutByName обрабатывает запрос на получение макета по имени
func (gul layoutByName) layoutByName(c *gin.Context) {
	initiatorUserID, err := http2.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	layoutName := c.Param("name")

	layout, err := gul.layout.GetLayoutByName(c.Request.Context(), layoutName, initiatorUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, layout)
}

// NewLayoutByName создает новый обработчик для получения макета по имени
func NewLayoutByName(usecase layoutUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		gul := layoutByName{layout: usecase}
		return gul.method(), gul.path(), gul.layoutByName
	}
}
