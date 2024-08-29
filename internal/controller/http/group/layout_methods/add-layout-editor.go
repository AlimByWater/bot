package layout_methods

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type addLayoutEditor struct {
	layout layoutUC
}

func (ale addLayoutEditor) method() string {
	return http.MethodPost
}

func (ale addLayoutEditor) path() string {
	return "/layout/:id/editor"
}

// addLayoutEditor обрабатывает запрос на добавление редактора макета
func (ale addLayoutEditor) addLayoutEditor(c *gin.Context) {
	initiatorUserID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	layoutID := c.Param("id")

	var request struct {
		EditorID int `json:"editorId" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ale.layout.AddLayoutEditor(c.Request.Context(), layoutID, initiatorUserID, request.EditorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

// NewAddLayoutEditor создает новый обработчик для добавления редактора макета
func NewAddLayoutEditor(usecase layoutUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		ale := addLayoutEditor{layout: usecase}
		return ale.method(), ale.path(), ale.addLayoutEditor
	}
}
