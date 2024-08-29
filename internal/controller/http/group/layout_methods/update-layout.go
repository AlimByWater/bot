package layout_methods

import (
	"arimadj-helper/internal/entity"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type updateLayout struct {
	layout layoutUC
}

func (ul updateLayout) method() string {
	return http.MethodPut
}

func (ul updateLayout) path() string {
	return "/layout/:id"
}

// sendEvent обрабатывает запрос на обновление макета
func (ul updateLayout) sendEvent(c *gin.Context) {
	initiatorUserID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	layoutID := c.Param("id")
	var updatedLayout entity.UserLayout
	if err := c.ShouldBindJSON(&updatedLayout); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ul.layout.UpdateLayoutFull(c.Request.Context(), layoutID, initiatorUserID, updatedLayout)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNoPermission):
			c.JSON(http.StatusForbidden, gin.H{"error": "У вас нет прав на редактирование этого макета"})
		case errors.Is(err, entity.ErrLayoutNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Макет не найден"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить макет"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Макет успешно обновлен"})
}

func NewUpdateLayout(usecase layoutUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		ul := updateLayout{layout: usecase}
		return ul.method(), ul.path(), ul.sendEvent
	}
}
