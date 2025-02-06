package inline_keyboard

import (
	"log/slog"
	"strconv"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

// Определяем минимальный интерфейс progressManager, с которым будем работать.
type ProgressManager interface {
	Cancel(cancelKey string)
}

type CancelEmojiPackGeneration struct {
	logger          *slog.Logger
	progressManager ProgressManager
}

// Конструктор хэндлера.
func NewCancelEmojiPackGeneration(pm ProgressManager) *CancelEmojiPackGeneration {
	return &CancelEmojiPackGeneration{
		progressManager: pm,
	}
}

func (h *CancelEmojiPackGeneration) AddLogger(logger *slog.Logger) {
	h.logger = logger.With(slog.String("handler", "cancel_emoji_pack_generation"))
}

func (h *CancelEmojiPackGeneration) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		if update.CallbackQuery == nil {
			return
		}
		data := update.CallbackQuery.Data

		cancelKey := strings.TrimPrefix(data, "cancel_")
		parts := strings.Split(cancelKey, "_")
		if len(parts) < 3 {
			h.logger.Error("Неверные данные callback для отмены", "data", data)
			_ = bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "Неверные данные для отмены генерации",
				ShowAlert:       true,
			})
			return
		}

		// chatID, err := strconv.ParseInt(parts[0], 10, 64)
		// if err != nil {
		// 	h.logger.Error("Неверный chatID в callback", "data", data, "error", err)
		// 	return
		// }
		// msgID, err := strconv.Atoi(parts[1])
		// if err != nil {
		// 	h.logger.Error("Неверный msgID в callback", "data", data, "error", err)
		// 	return
		// }
		initiatorID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			h.logger.Error("Неверный initiatorID в callback", "data", data, "error", err)
			return
		}

		// Проверяем, что вызывающий callback – тот же, кто инициировал генерацию
		if update.CallbackQuery.From.ID != initiatorID {
			_ = bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
				CallbackQueryID: update.CallbackQuery.ID,
				Text:            "🖕",
				ShowAlert:       true,
			})
			return
		}

		// Выполняем отмену через progressManager
		h.progressManager.Cancel(cancelKey)

		_ = bot.AnswerCallbackQuery(&telego.AnswerCallbackQueryParams{
			CallbackQueryID: update.CallbackQuery.ID,
			Text:            "Генерация отменена",
		})
	}
}

func (h *CancelEmojiPackGeneration) Predicate() telegohandler.Predicate {
	return telegohandler.CallbackDataPrefix("cancel_")
}
