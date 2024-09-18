package auth_methods

import (
	"elysium/internal/entity"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type refreshTokenTelegram struct {
	usecase authUC
}

func (s refreshTokenTelegram) method() string {
	return http.MethodPost
}

func (s refreshTokenTelegram) path() string {
	return "/telegram/refresh"
}

func (s refreshTokenTelegram) refresh(c *gin.Context) {
	var info entity.TelegramRefreshTokenInfo
	if err := c.ShouldBindJSON(&info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := s.usecase.RefreshToken(c.Request.Context(), info.RefreshToken)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidToken) || errors.Is(err, entity.ErrExpiredRefreshToken) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, token)
}

func NewRefreshMethod(usecase authUC) func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return func() (method string, path string, handlerFunc gin.HandlerFunc) {
		stat := refreshTokenTelegram{usecase: usecase}
		return stat.method(), stat.path(), stat.refresh

	}
}
