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

	_, _ = initiatorUserID, wantedUserID

	//c.JSON(http.StatusOK, layout)
}

func NewUpdateLayout(usecase layoutUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		ul := updateLayout{layout: usecase}
		return ul.method(), ul.path(), ul.sendEvent
	}
}
