package command

import (
	"elysium/internal/usecase/use_message"
	"log/slog"
	"reflect"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

type BuyTokens struct {
	logger *slog.Logger
}

func NewBuyTokens() *BuyTokens {
	return &BuyTokens{}
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
		text := use_message.GL.BuyTokens(langCode)

		labeldPrices := []telego.LabeledPrice{
			{"⭐️50", 50},
			{"⭐️⭐️100", 100},
			{"⭐️⭐️⭐️500", 500},
		}

		invoiceLinks := make([]string, 0, len(labeldPrices))
		buttons := make([]telego.InlineKeyboardButton, 0, len(labeldPrices))

		for _, price := range labeldPrices {
			invoiceLink, err := bot.CreateInvoiceLink(&telego.CreateInvoiceLinkParams{
				Title:       "title",
				Payload:     "payload",
				Description: text,
				//ProviderToken: "provider_token",
				Currency: "XTR",
				Prices:   []telego.LabeledPrice{price},
			})
			if err != nil {
				h.logger.Error("create invoice link", slog.String("err", err.Error()))
				continue
			}
			invoiceLinks = append(invoiceLinks, *invoiceLink)
			buttons = append(buttons, telegoutil.InlineKeyboardButton(price.Label).WithURL(*invoiceLink).WithPay())
		}

		inlineKeyboard := telegoutil.InlineKeyboard(
			telegoutil.InlineKeyboardRow(buttons[0]),
			telegoutil.InlineKeyboardRow(buttons[1]),
			telegoutil.InlineKeyboardRow(buttons[2]),
			telegoutil.InlineKeyboardRow(
				telegoutil.InlineKeyboardButton(use_message.GL.BackBtn(langCode)).WithCallbackData("start"),
			),
		)

		//inv := telegoutil.Invoice(chat.ChatID(), "title", text, "payload", "", "XTR")
		//inv.WithPrices(labeldPrices[0])
		//inv.WithReplyMarkup(inlineKeyboard)

		//m, err := bot.SendInvoice(inv)

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
	return telegohandler.Or(telegohandler.CallbackDataPrefix(h.Command()), telegohandler.CommandEqual(h.Command()))
}
