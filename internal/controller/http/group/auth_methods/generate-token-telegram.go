package auth_methods

import (
	"elysium/internal/entity"
	"github.com/gin-gonic/gin"
	"net/http"
)

type generateTokenTelegram struct {
	usecase authUC
}

func (s generateTokenTelegram) method() string {
	return http.MethodPost
}

func (s generateTokenTelegram) path() string {
	return "/telegram"
}

func (s generateTokenTelegram) generate(c *gin.Context) {
	var info entity.TelegramLoginInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := s.usecase.GenerateTokenForTelegram(c.Request.Context(), info)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, token)
}

func NewGenerateMethod(usecase authUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		stat := generateTokenTelegram{usecase: usecase}
		return stat.method(), stat.path(), stat.generate
	}
}
