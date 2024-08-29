package layout_methods

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type updateLayout struct {
	layout layoutUC
}

func (ul updateLayout) method() string {
	return http.MethodPut
}

func (ul updateLayout) path() string {
	return "/layout/:userID"
}

func (ul updateLayout) sendEvent(c *gin.Context) {
	initiatorUserID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wantedUserID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем текущий макет
	currentLayout, err := ul.layout.GetUserLayout(c.Request.Context(), wantedUserID, initiatorUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, является ли инициатор создателем или редактором
	isCreator := currentLayout.Creator == strconv.Itoa(initiatorUserID)
	isEditor, err := ul.layout.IsEditor(c.Request.Context(), currentLayout.LayoutID, strconv.Itoa(initiatorUserID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !isCreator && !isEditor {
		c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to edit this layout"})
		return
	}

	// Обновляем макет
	var updatedLayout entity.UserLayout
	if err := c.ShouldBindJSON(&updatedLayout); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Сохраняем создателя и редакторов
	updatedLayout.Creator = currentLayout.Creator
	updatedLayout.Editors = currentLayout.Editors

	if err := ul.layout.UpdateLayoutFull(c.Request.Context(), wantedUserID, updatedLayout); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Layout updated successfully"})
}

func NewUpdateLayout(usecase layoutUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		ul := updateLayout{layout: usecase}
		return ul.method(), ul.path(), ul.sendEvent
	}
}
