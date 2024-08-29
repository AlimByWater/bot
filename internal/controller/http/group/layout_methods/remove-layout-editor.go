package layout_methods

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type removeLayoutEditor struct {
	layout layoutUC
}

func (rle removeLayoutEditor) method() string {
	return http.MethodDelete
}

func (rle removeLayoutEditor) path() string {
	return "/layout/:id/editor/:editorId"
}

// removeLayoutEditor обрабатывает запрос на удаление редактора макета
func (rle removeLayoutEditor) removeLayoutEditor(c *gin.Context) {
	initiatorUserID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	layoutID := c.Param("id")
	editorIDStr := c.Param("editorId")

	editorID, err := strconv.Atoi(editorIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid editor ID"})
		return
	}

	err = rle.layout.RemoveLayoutEditor(c.Request.Context(), layoutID, initiatorUserID, editorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// NewRemoveLayoutEditor создает новый обработчик для удаления редактора макета
func NewRemoveLayoutEditor(usecase layoutUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		rle := removeLayoutEditor{layout: usecase}
		return rle.method(), rle.path(), rle.removeLayoutEditor
	}
}
