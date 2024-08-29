package layout_methods

import (
	"arimadj-helper/internal/entity"
	"errors"
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

	var updatedLayout entity.UserLayout
	if err := c.ShouldBindJSON(&updatedLayout); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ul.layout.UpdateLayoutFull(c.Request.Context(), wantedUserID, initiatorUserID, updatedLayout)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNoPermission):
			c.JSON(http.StatusForbidden, gin.H{"error": "You don't have permission to edit this layout"})
		case errors.Is(err, entity.ErrLayoutNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Layout not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update layout"})
		}
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
