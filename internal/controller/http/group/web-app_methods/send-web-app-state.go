package web_app_methods

import (
	"elysium/internal/entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

type webAppState struct {
	usecase botUC
}

func (was webAppState) method() string {
	return http.MethodPost
}

func (was webAppState) path() string {
	return "/state"
}

func (was webAppState) sendState(c *gin.Context) {
	var state entity.WebAppState
	if err := c.ShouldBindJSON(&state); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process the received event
	was.usecase.ProcessWebAppState(c.Request.Context(), state)

	c.JSON(http.StatusOK, gin.H{"message": "Web app event processed successfully"})
}

func NewWebAppState(usecase botUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		wae := webAppState{usecase: usecase}
		return wae.method(), wae.path(), wae.sendState
	}
}
