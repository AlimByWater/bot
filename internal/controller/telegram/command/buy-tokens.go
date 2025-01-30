package command

import (
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"log/slog"
	"reflect"
)

type BuyTokens struct {
	logger  *slog.Logger
	message interface {
		Error(langCode string) (msg string)
		BuyTokens(langCode string) (msg string)
		BackBtn(langCode string) (msg string)
	}
}

func NewBuyTokens(
	message interface {
		Error(langCode string) (msg string)
		BuyTokens(langCode string) (msg string)
		BackBtn(langCode string) (msg string)
	}) *BuyTokens {
	return &BuyTokens{
		message: message,
	}
}

func (h *BuyTokens) AddLogger(logger *slog.Logger) {
	h.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(h)).Type().PkgPath()), slog.String("command", h.Command()))
}

func (h *BuyTokens) Command() string     { return "buy_tokens" }
func (h *BuyTokens) Description() string { return "Купить Дрипкоины" }

func (h *BuyTokens) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		if update.CallbackQuery == nil {
			return
		}

		chat := update.CallbackQuery.Message.GetChat()

		langCode := update.CallbackQuery.From.LanguageCode
		h.logger.Info("buy tokens", slog.String("from", chat.Username))
		text := h.message.BuyTokens(langCode)
		inlineKeyboard := telegoutil.InlineKeyboard(
			telegoutil.InlineKeyboardRow(
				telegoutil.InlineKeyboardButton("50").WithCallbackData("buy:50"),
				telegoutil.InlineKeyboardButton("100").WithCallbackData("buy:100"),
				telegoutil.InlineKeyboardButton("500").WithCallbackData("buy:500"),
			),
			telegoutil.InlineKeyboardRow(
				telegoutil.InlineKeyboardButton(h.message.BackBtn(langCode)).WithCallbackData("start"),
			),
		)

		_, err := bot.EditMessageText(&telego.EditMessageTextParams{
			ChatID:      chat.ChatID(),
			MessageID:   update.CallbackQuery.Message.GetMessageID(),
			Text:        text,
			ReplyMarkup: inlineKeyboard,
		})
		if err != nil {
			h.logger.Error("send message", slog.String("err", err.Error()))
		}
	}
}

func (h *BuyTokens) Predicate() telegohandler.Predicate {
	return telegohandler.CallbackDataPrefix(h.Command())
}
