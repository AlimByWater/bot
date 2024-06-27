package http

import (
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"
)

func ginLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		end := time.Now()
		latency := end.Sub(start)
		status := c.Writer.Status()

		attributes := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", c.Request.Method),
			slog.String("route", c.FullPath()),
			slog.String("ip", c.ClientIP()),
			slog.String("latency", latency.String()),
			slog.String("user-agent", c.Request.UserAgent()),
		}

		level := slog.LevelInfo
		msg := "Incoming request"
		if status >= http.StatusBadRequest && status < http.StatusInternalServerError {
			level = slog.LevelWarn
			msg = c.Errors.String()
		} else if status >= http.StatusInternalServerError {
			level = slog.LevelError
			msg = c.Errors.String()
		}

		logger.WithGroup("request").LogAttrs(c.Request.Context(), level, msg, attributes...)
	}
}

func ginRecovery(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				attributes := []slog.Attr{
					slog.String("method", c.Request.Method),
					slog.String("route", c.FullPath()),
					slog.String("ip", c.ClientIP()),
					slog.String("user-agent", c.Request.UserAgent()),
					slog.String("stack", string(debug.Stack())),
					slog.Any("err", err),
				}

				logger.WithGroup("recovery").LogAttrs(c.Request.Context(), slog.LevelError, "Recovery", attributes...)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}

// Middleware для проверки заголовка x-api-key
func apiKeyAuthMiddleware(logger *slog.Logger, apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("x-api-key") != apiKey {
			logger.Error("😈unauthorized access", slog.String("from IP", c.ClientIP()), slog.String("remote IP", c.RemoteIP()))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func cors(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, x-api-key, baggage")

		// Обработка preflight-запросов (OPTIONS)
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusOK)
			return
		}

		c.Next()
	}
}
