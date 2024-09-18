package tampermonkey_methods

import (
	"elysium/internal/entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

type statistic struct {
	usecase botUC
}

func (s statistic) method() string {
	return http.MethodPost
}

func (s statistic) path() string {
	return "/submit"
}

func (s statistic) submit(c *gin.Context) {
	var info entity.TrackInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if info.ArtistName == "Unknown" && info.TrackTitle == "Unknown" && info.TrackLink == "Unknown" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	//j, _ := json.MarshalIndent(info, "", "  ")
	//fmt.Println(string(j))

	s.usecase.NextSong(info)
	c.Status(http.StatusOK)
}

func NewSubmitMethod(usecase botUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		stat := statistic{usecase: usecase}
		return stat.method(), stat.path(), stat.submit
	}
}
