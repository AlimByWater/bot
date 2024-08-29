package layout_methods

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type layoutByID struct {
	layout layoutUC
}

func (gul layoutByID) method() string {
	return http.MethodGet
}

func (gul layoutByID) path() string {
	return "/layout/:id"
}

// layoutByID обрабатывает запрос на получение макета пользователя
func (gul layoutByID) layoutByID(c *gin.Context) {
	initiatorUserID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	layoutID := c.Param("id")

	layout, err := gul.layout.GetLayout(c.Request.Context(), layoutID, initiatorUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, layout)
}

// NewLayoutByID создает новый обработчик для получения макета пользователя
func NewLayoutByID(usecase layoutUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		gul := layoutByID{layout: usecase}
		return gul.method(), gul.path(), gul.layoutByID
	}
}
