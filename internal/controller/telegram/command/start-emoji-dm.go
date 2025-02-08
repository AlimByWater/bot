package command

import (
	"elysium/internal/controller/telegram/inline_keyboard"
	"log/slog"
	"reflect"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

type StartEmojiDM struct {
	logger  *slog.Logger
	message interface {
		Error(langCode string) (msg string)
		BalanceBtn(langCode string) (msg string)
		BotsListBtn(langCode string) (msg string)
		SupportBtn(langCode string) (msg string)
		BuyTokensBtn(langCode string) (msg string)
		StartDripTech(langCode string) (msg string)
		CreatePackkInfoBtn(lang string) string
		MyPacksBtn(lang string) string
	}
}

func NewStartEmojiDM(
	message interface {
		Error(langCode string) (msg string)
		BalanceBtn(langCode string) (msg string)
		BotsListBtn(langCode string) (msg string)
		SupportBtn(langCode string) (msg string)
		BuyTokensBtn(langCode string) (msg string)
		StartDripTech(langCode string) (msg string)
		CreatePackkInfoBtn(lang string) string
		MyPacksBtn(lang string) string
	}) *StartEmojiDM {
	return &StartEmojiDM{
		message: message,
	}
}

func (h *StartEmojiDM) AddLogger(logger *slog.Logger) {
	h.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(h)).Type().PkgPath()))
}

func (h *StartEmojiDM) Command() string     { return "start" }
func (h *StartEmojiDM) Description() string { return "Start DripTech Bot" }

func (h *StartEmojiDM) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		var lang string
		var chat telego.Chat
		if update.CallbackQuery != nil {
			lang = update.CallbackQuery.From.LanguageCode
			chat = update.CallbackQuery.Message.GetChat()
		} else if update.Message != nil {
			lang = update.Message.From.LanguageCode
			chat = update.Message.Chat
		}
		text := h.message.StartDripTech(lang)

		inlineKeyboard := inline_keyboard.GetEmojiBotStartKeyboard(lang, h.message)

		if update.CallbackQuery != nil {
			_, err := bot.EditMessageText(&telego.EditMessageTextParams{
				ChatID:      chat.ChatID(),
				MessageID:   update.CallbackQuery.Message.GetMessageID(),
				Text:        text,
				ReplyMarkup: inlineKeyboard,
			})
			if err != nil {
				h.logger.Error("send message", slog.String("err", err.Error()))
			}
		} else if update.Message != nil {
			message := telegoutil.Message(
				chat.ChatID(),
				text,
			).WithReplyMarkup(inlineKeyboard)
			_, err := bot.SendMessage(message)
			if err != nil {
				h.logger.Error("send message", slog.String("err", err.Error()))
			}

		}

	}
}

func (h *StartEmojiDM) Predicate() telegohandler.Predicate {
	return telegohandler.CommandEqual(h.Command())
}
