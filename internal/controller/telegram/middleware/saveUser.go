package middleware

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"log/slog"
	"reflect"
	"strconv"
	"strings"
)

type SaveUser struct {
	logger *slog.Logger
	cache  interface {
		LoadAndDelete(key string) (value any, loaded bool)
	}
	userUC interface {
		CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)
		SetUserToBotActive(ctx context.Context, userID int, botID int64) error
	}
}

func NewSaveUser(
	cache interface {
		LoadAndDelete(key string) (value any, loaded bool)
	},
	userUC interface {
		CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)
		SetUserToBotActive(ctx context.Context, userID int, botID int64) error
	},
) *SaveUser {
	return &SaveUser{
		userUC: userUC,
		cache:  cache,
	}
}

func (m *SaveUser) AddLogger(logger *slog.Logger) {
	m.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(m)).Type().PkgPath()))
}

func (m *SaveUser) Handler() telegohandler.Middleware {
	return func(bot *telego.Bot, update telego.Update, next telegohandler.Handler) {
		var user *telego.User
		//var langCode string
		switch {
		case update.Message != nil:
			user = update.Message.From
		case update.CallbackQuery != nil:
			user = &update.CallbackQuery.From
		default:
			next(bot, update)
			return
		}

		if user == nil {
			m.logger.Debug("Skipping update without user",
				slog.Any("update", update))
			next(bot, update)
			return
		}

		// Проверяем валидность пользователя
		if user.IsBot || user.ID <= 0 {
			m.logger.Debug("Skipping invalid user",
				slog.Any("user", user),
				slog.Bool("is_bot", user != nil && user.IsBot),
				slog.Int64("user_id", user.ID))
			next(bot, update)
			return
		}

		savedUser, err := m.userUC.CreateOrUpdateUser(context.Background(), entity.User{
			TelegramID:       user.ID,
			TelegramUsername: user.Username,
			Firstname:        user.FirstName,
		})

		if err != nil {
			m.logger.Error("Failed to save user",
				slog.Int64("user_id", user.ID),
				slog.String("error", err.Error()))
		}

		if update.Message != nil && strings.HasPrefix(update.Message.Text, "/start") {
			handler, ok := m.cache.LoadAndDelete(fmt.Sprintf("%s:%d", entity.CacheKeyInitMessageToDelete, user.ID))
			if ok {
				switch handler.(type) {
				case telegohandler.Handler:
					handler.(telegohandler.Handler)(bot, update)
				default:
				}
			}

			botUser, ok := update.Context().Value(entity.BotSelfCtxKey).(*telego.User)
			if !ok || botUser == nil {
				m.logger.Error("bot info not found in context")
				next(bot, update)
				return
			}

			botID, err := strconv.ParseInt(fmt.Sprintf("-100%d", botUser.ID), 10, 64)
			if err != nil {
				m.logger.Error("Failed to parse bot ID",
					slog.Int64("bot_id", botUser.ID),
					slog.String("error", err.Error()))
				next(bot, update)
				return
			}

			err = m.userUC.SetUserToBotActive(context.Background(), savedUser.ID, botID)
			if err != nil {
				m.logger.Error("Failed to activate bot for user",
					slog.Int("user_id", savedUser.ID),
					slog.Int64("bot_id", botID),
					slog.String("error", err.Error()))
			}
		}

		next(bot, update)
	}
}
