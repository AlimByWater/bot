package layout_methods

import (
	http2 "arimadj-helper/internal/controller/http"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type layoutByID struct {
	layout layoutUC
}

func (gul layoutByID) method() string {
	return http.MethodGet
}

func (gul layoutByID) path() string {
	return "/:id"
}

// layoutByID обрабатывает запрос на получение макета пользователя
func (gul layoutByID) layoutByID(c *gin.Context) {
	initiatorUserID, err := http2.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	layoutIDStr := c.Param("id")
	layoutID, err := strconv.Atoi(layoutIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid layout ID"})
		return
	}

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
