package api_methods

import (
	"arimadj-helper/internal/entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

type botUC interface {
	NextSong(track entity.TrackInfo)
}

type statistic struct {
	usecase botUC
}

func (s statistic) method() string {
	return http.MethodPost
}

func (s statistic) path() string {
	return "/submit"
}

// getExample    godoc
// @Summary      Sum a and b
// @Description  Get sun
// @Tags         Sum
// @Accept       json
// @Produce      json
// @Param        a header int true "a"
// @Param        b header int true "b"
// @Success      200 {integer} integer "sum"
// @Router       /api/example [get]
func (s statistic) submit(c *gin.Context) {
	var info entity.TrackInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	s.usecase.NextSong(info)
	c.Status(http.StatusOK)
}

func NewSubmitMethod(usecase botUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		stat := statistic{usecase: usecase}
		return stat.method(), stat.path(), stat.submit
	}
}
