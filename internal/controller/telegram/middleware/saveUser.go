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
	userUC interface {
		CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)
		SetUserToBotActive(ctx context.Context, userID int, botID int64) error
	}
}

func NewSaveUser(
	userUC interface {
		CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)
		SetUserToBotActive(ctx context.Context, userID int, botID int64) error
	},
) *SaveUser {
	return &SaveUser{
		userUC: userUC,
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

		// Проверяем валидность пользователя
		if user == nil || user.IsBot || user.ID <= 0 {
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
			b, err := bot.GetMe()
			if err != nil {
				m.logger.Error("Failed to get bot info",
					slog.String("error", err.Error()))
				return
			} else {
				botIDstr := fmt.Sprintf("-100%d", b.ID)
				botID, err := strconv.ParseInt(botIDstr, 10, 64)
				if err != nil {
					m.logger.Error("Failed to parse bot ID",
						slog.String("error", err.Error()),
						slog.String("bot_id", botIDstr))
					return
				}

				err = m.userUC.SetUserToBotActive(context.Background(), savedUser.ID, botID)
				if err != nil {
					m.logger.Error("Failed to activate bot for user",
						slog.Int("user_id", savedUser.ID),
						slog.Int64("bot_id", botID),
						slog.String("error", err.Error()))
				}
				m.logger.Debug("Activated bot for user",
					slog.Int("user_id", savedUser.ID),
					slog.Int64("bot_id", botID))
			}
		}

		next(bot, update)
	}
}
