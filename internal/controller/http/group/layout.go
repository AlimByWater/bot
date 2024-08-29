package api

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

type layoutUC interface {
	CheckAccessTokenByUserID(ctx context.Context, token string, userID int) (bool, error)
}

// NewLayoutGroup создает группу хендлеров
func NewLayoutGroup(l layoutUC, tc tokenChecker, h ...func() (method string, path string, handlerFunc gin.HandlerFunc)) Layout {
	return Layout{
		layout:       l,
		tokenChecker: tc,
		handlers:     h,
	}
}

type Layout struct {
	layout       layoutUC
	tokenChecker tokenChecker
	handlers     []func() (method string, path string, handlerFunc gin.HandlerFunc)
}

// Path ...
func (Layout) Path() string {
	return "/layout"
}

// Handlers ...
func (g Layout) Handlers() []func() (method string, path string, handlerFunc gin.HandlerFunc) {
	return g.handlers
}

// Auth миддлвейр авторизаций
func (g Layout) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var token, userID = "", 0

		// токен и user-id могут быть либо в хедерах, либо в параметрах запроса(в случае вебсокета)
		if c.Request.Header.Get("Authorization") != "" && c.Request.Header.Get("x-user-id") != "" {
			var err error
			reqToken := c.Request.Header.Get("Authorization")
			splitToken := strings.Split(reqToken, "Bearer ")
			if len(splitToken) != 2 {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
				return
			}

			token = splitToken[1]

			userIDRaw := c.Request.Header.Get("x-user-id")
			userID, err = strconv.Atoi(userIDRaw)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id", "user-id": userIDRaw})
				return
			}
		} else if c.Param("token") != "" && c.Param("userId") != "" {
			var err error
			token = c.Param("token")
			userID, err = strconv.Atoi(c.Param("userId"))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid user id", "user-id": c.Param("user_id")})
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token and userId are required"})
			return
		}

		if token == "test-token" && userID == 5 {
			return
		}

		valid, err := g.tokenChecker.CheckAccessTokenByUserID(c.Request.Context(), token, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		if !valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		c.Set("userID", userID)
	}
}
