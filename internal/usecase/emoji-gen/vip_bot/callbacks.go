package vip_bot

import (
	"context"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"strconv"
	"strings"
)

func (d *DBot) HandleCancelCallback(ctx context.Context, update *models.Update) {
	cancelKey := strings.TrimPrefix(update.CallbackQuery.Data, "cancel_")

	parts := strings.Split(cancelKey, "_")
	chatID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		d.logger.Error("cancel order: parse chatID", slog.String("cancelKey", cancelKey), slog.String("error", err.Error()))
	}
	msgID, err := strconv.Atoi(parts[1])
	if err != nil {
		d.logger.Error("cancel order: parse msgID", slog.String("cancelKey", cancelKey), slog.String("error", err.Error()))
	}

	initiatorID, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return
	}

	// Проверяем, что отменяет тот же пользователь, который начал генерацию
	if update.CallbackQuery.From.ID != initiatorID {
		// Отвечаем на callback с сообщением об ошибке
		_, _ = d.b.BotApi.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "🖕",
			ShowAlert:       true,
		})
		return
	}

	if chatID != 0 || msgID != 0 {
		d.progressManager.Cancel(cancelKey)
		// Отвечаем на callback
		_, err = d.b.BotApi.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Генерация отменена",
		})
		if err != nil {
			d.logger.Error("cancel order: answer callback", slog.String("cancelKey", cancelKey), slog.String("error", err.Error()))
		}
	} else {
		_, _ = d.b.BotApi.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Не удалось отменить генерацию",
		})
	}

	return

}
