package layout_methods

import (
	http2 "elysium/internal/controller/http"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type addLayoutEditor struct {
	layout layoutUC
}

func (ale addLayoutEditor) method() string {
	return http.MethodPost
}

func (ale addLayoutEditor) path() string {
	return "/:id/editor"
}

// addLayoutEditor обрабатывает запрос на добавление редактора макета
func (ale addLayoutEditor) addLayoutEditor(c *gin.Context) {
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
