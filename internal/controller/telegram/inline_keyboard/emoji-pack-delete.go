package inline_keyboard

import (
	"context"
	"log/slog"
	"strings"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

type EmojiPackDelete struct {
	logger  *slog.Logger
	message interface {
		Error(langCode string) string
		BuyTokensBtn(lang string) string
		PackDeletedSuccess(langCode string) string
		CreatePackkInfoBtn(lang string) string
		MyPacksBtn(lang string) string
	}
	repo interface {
		SetEmojiPackDeleted(ctx context.Context, packName string) error
	}
}

func NewEmojiPackDelete(
	message interface {
		Error(langCode string) string
		BuyTokensBtn(lang string) string
		PackDeletedSuccess(langCode string) string
		CreatePackkInfoBtn(lang string) string
		MyPacksBtn(lang string) string
	},
	repo interface {
		SetEmojiPackDeleted(ctx context.Context, packName string) error
	},
) *EmojiPackDelete {
	return &EmojiPackDelete{
		repo:    repo,
		message: message,
	}
}

func (h *EmojiPackDelete) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		if update.CallbackQuery == nil {
			return
		}

		user := update.CallbackQuery.From

		packName := strings.Split(update.CallbackQuery.Data, ":")[1]

		err := h.repo.SetEmojiPackDeleted(update.Context(), packName)
		if err != nil {
			h.logger.Error("Failed to delete emoji pack", slog.String("err", err.Error()), slog.Int64("user_id", user.ID), slog.String("data", update.CallbackQuery.Data), slog.String("username", user.Username))

			_, err = bot.SendMessage(telegoutil.Message(
				telegoutil.ID(user.ID),
				h.message.Error(user.LanguageCode),
			))
			if err != nil {
				h.logger.Error("Failed to send error message", slog.String("err", err.Error()), slog.Int64("user_id", user.ID), slog.String("username", user.Username))
				return
			}

			return
		}

		message := &telego.EditMessageTextParams{
			ChatID:      telegoutil.ID(user.ID),
			MessageID:   update.CallbackQuery.Message.GetMessageID(),
			Text:        h.message.PackDeletedSuccess(user.LanguageCode),
			ReplyMarkup: GetEmojiBotStartKeyboard(user.LanguageCode, h.message),
		}

		_, err = bot.EditMessageText(message)
		if err != nil {
			h.logger.Error("Failed to send message", slog.String("err", err.Error()), slog.Int64("user_id", user.ID), slog.String("username", user.Username))
			return
		}
	}
}

func (h *EmojiPackDelete) AddLogger(logger *slog.Logger) {
	h.logger = logger
}

func (h *EmojiPackDelete) Command() string {
	return "pack_delete:"
}

func (h *EmojiPackDelete) Predicate() telegohandler.Predicate {
	return telegohandler.Or(
		telegohandler.CallbackDataPrefix(h.Command()),
	)
}
