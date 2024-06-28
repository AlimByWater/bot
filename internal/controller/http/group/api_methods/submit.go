package api_methods

import (
	"arimadj-helper/internal/entity"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
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

	fmt.Println(len(info.Artwork))
	fmt.Println(info.ArtworkUrl)

	res, err := http.Get(info.ArtworkUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	img, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("read img body:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save artwork"})
		return
	}
	//// Сохранение обложки на файловую систему
	fileName := filepath.Join("./", time.Now().Format("20060102_150405")+".jpg")
	err = os.WriteFile(fileName, img, 0644)
	if err != nil {
		log.Println("Error saving artwork:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save artwork"})
		return
	}

	//s.usecase.NextSong(info)
	c.Status(http.StatusOK)
}

func NewSubmitMethod(usecase botUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		stat := statistic{usecase: usecase}
		return stat.method(), stat.path(), stat.submit
	}
}
