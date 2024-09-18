package song_methods

import (
	http2 "elysium/internal/controller/http"
	"github.com/gin-gonic/gin"
	"net/http"
)

type getSongByTrackLink struct {
	songDownloader songDownloadUC
}

func (gsdl getSongByTrackLink) method() string {
	return http.MethodGet
}

func (gsdl getSongByTrackLink) path() string {
	return "/by-link"
}

func (gsdl getSongByTrackLink) downloadSongByTrackLink(c *gin.Context) {
	userID, err := http2.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	trackLink := c.Query("link")

	// validate link
	if trackLink == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "link is required"})
		return
	}

	err = gsdl.songDownloader.SendSongByTrackLink(c.Request.Context(), userID, trackLink)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func NewSongByURL(usecase songDownloadUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		gsdl := getSongByTrackLink{songDownloader: usecase}
		return gsdl.method(), gsdl.path(), gsdl.downloadSongByTrackLink
	}
}
