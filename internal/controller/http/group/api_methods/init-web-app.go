package api_methods

import (
	"arimadj-helper/internal/entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

type webAppEvent struct {
	usecase botUC
}

func (wae webAppEvent) method() string {
	return http.MethodPost
}

func (wae webAppEvent) path() string {
	return "/web-app-event"
}

func (wae webAppEvent) submit(c *gin.Context) {
	var event entity.WebAppEvent
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Process the received event
	if err := wae.usecase.ProcessWebAppEvent(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process web app event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Web app event processed successfully"})
}

func NewWebAppEvent(usecase botUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		wae := webAppEvent{usecase: usecase}
		return wae.method(), wae.path(), wae.submit
	}
}
