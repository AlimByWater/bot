package layout_methods

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type getUserLayout struct {
	layout layoutUC
}

func (gul getUserLayout) method() string {
	return http.MethodGet
}

func (gul getUserLayout) path() string {
	return "/layout/:userID"
}

func (gul getUserLayout) sendEvent(c *gin.Context) {
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

	layout, err := gul.layout.GetUserLayout(c.Request.Context(), wantedUserID, initiatorUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, layout)
}

func NewGetUserLayout(usecase layoutUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		gul := getUserLayout{layout: usecase}
		return gul.method(), gul.path(), gul.sendEvent
	}
}
