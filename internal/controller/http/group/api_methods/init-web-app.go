package api_methods

import (
	"arimadj-helper/internal/entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

type initWebApp struct {
	usecase botUC
}

func (iwa initWebApp) method() string {
	return http.MethodPost
}

func (iwa initWebApp) path() string {
	return "/initwebapp"
}

func (iwa initWebApp) submit(c *gin.Context) {
	var data entity.InitWebApp
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	//iwa.usecase.NextSong(info)
	c.Status(http.StatusOK)
}

func NewInitWebApp(usecase botUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		iwa := initWebApp{usecase: usecase}
		return iwa.method(), iwa.path(), iwa.submit
	}
}
