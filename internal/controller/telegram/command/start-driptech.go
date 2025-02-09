package command

import (
	"elysium/internal/usecase/use_message"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
	"log/slog"
	"reflect"
)

type Start struct {
	logger *slog.Logger
}

func NewStart() *Start {
	return &Start{}
}

func (h *Start) AddLogger(logger *slog.Logger) {
	h.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(h)).Type().PkgPath()))
}

func (h *Start) Command() string     { return "start" }
func (h *Start) Description() string { return "Start DripTech Bot" }

func (h *Start) Handler() telegohandler.Handler {
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
		text := use_message.GL.StartDripTech(lang)

		inlineKeyboard := telegoutil.InlineKeyboard(
			telegoutil.InlineKeyboardRow(
				telegoutil.InlineKeyboardButton(use_message.GL.BalanceBtn(lang)).WithCallbackData("balance"),
				//telegoutil.InlineKeyboardButton(use_message.GL.BotsListBtn(lang)).WithCallbackData("bots_list"),
				telegoutil.InlineKeyboardButton(use_message.GL.SupportBtn(lang)).WithCallbackData("support"),
			),
			telegoutil.InlineKeyboardRow(
				telegoutil.InlineKeyboardButton(use_message.GL.BuyTokensBtn(lang)).WithCallbackData("buy_tokens"),
			),
		)

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

func (h *Start) Predicate() telegohandler.Predicate {
	return telegohandler.Or(telegohandler.CommandEqual(h.Command()), telegohandler.CallbackDataEqual(h.Command()))

}
