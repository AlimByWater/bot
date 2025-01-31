package middleware

import (
	"context"
	"elysium/internal/entity"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"log/slog"
)

type BotSelf struct {
	botUser *telego.User
	logger  *slog.Logger
}

func NewBotSelf(botUser *telego.User) *BotSelf {
	return &BotSelf{
		botUser: botUser,
	}
}

func (m *BotSelf) AddLogger(logger *slog.Logger) {
	m.logger = logger.With(slog.String("middleware", "bot_self"))
}

func (m *BotSelf) Handler() telegohandler.Middleware {
	return func(bot *telego.Bot, update telego.Update, next telegohandler.Handler) {
		next(bot, update.WithContext(context.WithValue(update.Context(), entity.BotSelfCtxKey, m.botUser)))
	}
}
