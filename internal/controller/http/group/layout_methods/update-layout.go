package layout_methods

import (
	http2 "elysium/internal/controller/http"
	"elysium/internal/entity"
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
	return "/:id"
}

// updateLayout обрабатывает запрос на обновление макета
func (ul updateLayout) updateLayout(c *gin.Context) {
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

	// {
	//	"name": "New Name",
	//	"background": {
	//		"type": "color",
	//		"value": "#FFFFFF"
	//	},
	//	"elements": [
	//		{
	//			"id": 145,
	//			"root_element_id": 1,
	//			"type": "clickable_navigable",
	//			"position": {
	//				"x": 1,
	//				"y": 1,
	//				"z_index": 1,
	//				"width": 1,
	//				"height": 1
	//			},
	//			"properties": {
	//				"icon": "path/to/icon1.png",
	//				"title": "Navigation Item",
	//				"navigationUrl": "/some-page"
	//			},
	//			"is_public": true,
	//			"is_removable": true
	//		},
	//		{
	//			"id": 145,
	//			"root_element_id": 2,
	//			"type": "clickable_navigable",
	//			"position": {
	//				"x": 1,
	//				"y": 1,
	//				"z_index": 1,
	//				"width": 1,
	//				"height": 1
	//			},
	//			"properties": {
	//				"icon": "path/to/icon1.png",
	//				"title": "Navigation Item",
	//				"navigationUrl": "/some-page"
	//			},
	//			"is_public": true,
	//			"is_removable": true
	//		},
	//	],
	//	"creator": 123,
	//	"editors": [123, 456]
	// }

	var updatedLayout entity.UserLayout
	if err := c.ShouldBindJSON(&updatedLayout); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = ul.layout.UpdateLayoutFull(c.Request.Context(), layoutID, initiatorUserID, updatedLayout)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrNoPermissionToEditLayout):
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
		return ul.method(), ul.path(), ul.updateLayout
	}
}
