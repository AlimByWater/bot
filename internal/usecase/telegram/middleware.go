package telegram

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"strings"
)

// saveUserMiddleware сохраняет или обновляет информацию о пользователе при любом взаимодействии
// и активирует бота при команде /start.
func (m *Manager) saveUserMiddleware(b *entity.Bot) func(bot.HandlerFunc) bot.HandlerFunc {
	return func(next bot.HandlerFunc) bot.HandlerFunc {
		return func(ctx context.Context, bot *bot.Bot, update *models.Update) {
			var user *models.User
			var isStartCommand bool

			m.logger.Debug("saveUserMiddleware", slog.String("method", "saveUserMiddleware"), slog.String("msg", fmt.Sprintf("%s", update.Message.Text)))

			if update.CallbackQuery != nil && update.CallbackQuery.Message.Message != nil {
				user = update.CallbackQuery.Message.Message.From
			} else if update.Message != nil {
				user = update.Message.From
				// Проверяем команду /start только для личных сообщений
				isStartCommand = update.Message.Chat.Type == models.ChatTypePrivate &&
					strings.HasPrefix(update.Message.Text, "/start")
			}

			if user != nil {
				entityUser := entity.User{
					TelegramID:       user.ID,
					TelegramUsername: user.Username,
					Firstname:        user.FirstName,
				}

				savedUser, err := m.userUC.CreateOrUpdateUser(ctx, entityUser)
				if err != nil {
					m.logger.Error("Failed to save user basic info",
						slog.Int64("user_id", user.ID),
						slog.String("error", err.Error()))
					next(ctx, bot, update)
					return
				}

				// Если это команда /start, активируем бота для пользователя
				if isStartCommand {
					err = m.userUC.SetUserToBotActive(ctx, savedUser.ID, b.ID)
					if err != nil {
						m.logger.Error("Failed to set user bot active",
							slog.Int("user_id", savedUser.ID),
							slog.Int64("bot_id", b.ID),
							slog.String("error", err.Error()))
					}
				}
			}

			next(ctx, bot, update)
		}
	}
}
