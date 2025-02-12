package payments

import (
	"context"
	"elysium/internal/entity"
	"elysium/internal/usecase/use_message"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"github.com/mymmrac/telego/telegoutil"
)

const defaultProvider = "yoomoney"

type BuyTokens struct {
	logger *slog.Logger
	cache  interface {
		Store(key string, value any)
	}
	userUC interface {
		CreateBulkPendingDeposits(ctx context.Context, botID int64, telegramUserID int64, amounts []int, provider string) ([]entity.UserTransaction, error)
	}
}

func NewBuyTokens(
	userUC interface {
		CreateBulkPendingDeposits(ctx context.Context, botID int64, telegramUserID int64, amounts []int, provider string) ([]entity.UserTransaction, error)
	},
	cache interface {
		Store(key string, value any)
	},
) *BuyTokens {
	return &BuyTokens{
		userUC: userUC,
		cache:  cache,
	}
}

func (h *BuyTokens) AddLogger(logger *slog.Logger) {
	h.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(h)).Type().PkgPath()))
}

func (h *BuyTokens) Command() string     { return "buy_tokens" }
func (h *BuyTokens) Description() string { return "Купить Дрипкоины" }

func (h *BuyTokens) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		var chat telego.Chat
		var langCode string
		var messageID int
		var telegramUserID int64

		if update.CallbackQuery != nil {
			chat = update.CallbackQuery.Message.GetChat()
			langCode = update.CallbackQuery.From.LanguageCode
			messageID = update.CallbackQuery.Message.GetMessageID()
			telegramUserID = update.CallbackQuery.From.ID
		} else if update.Message != nil {
			chat = update.Message.GetChat()
			langCode = update.Message.From.LanguageCode
			telegramUserID = update.Message.From.ID
		}

		text := use_message.GL.BuyTokens(langCode)

		// Определяем доступные цены
		labeldPrices := []telego.LabeledPrice{
			// {"50", 5000},
			{"100", 10000},
			{"500", 50000},
		}

		// Создаем слайс сумм для создания транзакций
		amounts := make([]int, 0, len(labeldPrices))
		for _, price := range labeldPrices {
			amounts = append(amounts, price.Amount)
		}

		botUser, ok := update.Context().Value(entity.BotSelfCtxKey).(*telego.User)
		if !ok || botUser == nil {
			h.logger.Error("bot info not found in context")
			return
		}

		// Создаем транзакции
		txns, err := h.userUC.CreateBulkPendingDeposits(context.Background(), botUser.ID, telegramUserID, amounts, defaultProvider)
		if err != nil {
			h.logger.Error("create bulk pending transactions", slog.String("err", err.Error()))
			return
		}

		// Создаем мапу для соответствия суммы и транзакции
		amountToTxn := make(map[int]entity.UserTransaction)
		for i, txn := range txns {
			amountToTxn[amounts[i]] = txn
		}

		invoiceLinks := make([]string, 0, len(labeldPrices))
		buttons := make([]telego.InlineKeyboardButton, 0, len(labeldPrices))

		for _, price := range labeldPrices {
			txn := amountToTxn[price.Amount]
			payload := entity.InvoicePayload(txn.ID)

			invoiceLink, err := bot.CreateInvoiceLink(&telego.CreateInvoiceLinkParams{
				Title:         "title",
				Payload:       payload,
				Description:   text,
				PhotoURL:      "https://4572d2e7-566d-4343-8fb5-9614a527af3d.selstorage.ru/checkout.jpeg",
				ProviderToken: "381764678:TEST:111587",
				Currency:      "RUB",
				Prices:        []telego.LabeledPrice{price},
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
			// telegoutil.InlineKeyboardRow(buttons[2]),
			telegoutil.InlineKeyboardRow(
				telegoutil.InlineKeyboardButton(use_message.GL.BackBtn(langCode)).WithCallbackData("start"),
			),
		)

		//inv := telegoutil.Invoice(chat.ChatID(), "title", text, "payload", "", "XTR")
		//inv.WithPrices(labeldPrices[0])
		//inv.WithReplyMarkup(inlineKeyboard)

		//m, err := bot.SendInvoice(inv)

		var message *telego.Message

		if messageID != 0 {
			message, err = bot.EditMessageText(&telego.EditMessageTextParams{
				ChatID:      chat.ChatID(),
				MessageID:   update.CallbackQuery.Message.GetMessageID(),
				Text:        text,
				ReplyMarkup: inlineKeyboard,
			})
			if err != nil {
				h.logger.Error("edit message", slog.String("err", err.Error()))
			}
		} else {
			message, err = bot.SendMessage(&telego.SendMessageParams{
				ChatID:      chat.ChatID(),
				Text:        text,
				ReplyMarkup: inlineKeyboard,
			})
			if err != nil {
				h.logger.Error("send message", slog.String("err", err.Error()))
			}
		}

		if message != nil {
			key := fmt.Sprintf("remove_after_success_payment:%d", chat.ID) // todo если пользоваться где то кроме личных чатов, то нужно пересмотреть логику формирования ключа
			h.cache.Store(key, h.removeMessageWithInvoicesAfterSuccessPayment(message.MessageID))
		}

	}
}

func (h *BuyTokens) removeMessageWithInvoicesAfterSuccessPayment(msgID int) telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		err := bot.DeleteMessage(&telego.DeleteMessageParams{
			ChatID:    update.Message.Chat.ChatID(),
			MessageID: msgID,
		})
		if err != nil {
			h.logger.Error("delete message with invoice", slog.String("err", err.Error()), slog.Int64("telegram_id", update.Message.From.ID))
		}
	}
}

func (h *BuyTokens) Predicate() telegohandler.Predicate {
	return telegohandler.Or(telegohandler.CallbackDataPrefix(h.Command()), telegohandler.CommandEqual(h.Command()))
}
