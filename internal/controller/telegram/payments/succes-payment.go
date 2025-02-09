package payments

import (
	"context"
	"elysium/internal/entity"
	"elysium/internal/usecase/use_message"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
	"log/slog"
	"reflect"
)

type SuccessPayment struct {
	logger *slog.Logger
	userUC interface {
		UserByTelegramID(ctx context.Context, userID int64) (entity.User, error)
		CompleteTransaction(ctx context.Context, txnID string) error
	}
}

func NewSuccessPayment(userUC interface {
	UserByTelegramID(ctx context.Context, userID int64) (entity.User, error)
	CompleteTransaction(ctx context.Context, txnID string) error
}) *SuccessPayment {
	return &SuccessPayment{
		userUC: userUC,
	}
}

func (h *SuccessPayment) AddLogger(logger *slog.Logger) {
	h.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(h)).Type().PkgPath()))
}

func (h *SuccessPayment) Command() string     { return "" }
func (h *SuccessPayment) Description() string { return "Success payment" }

func (h *SuccessPayment) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		msg := update.Message
		payment := msg.SuccessfulPayment
		langCode := msg.From.LanguageCode

		// Получаем пользователя по Telegram ID
		// user, err := h.userUC.UserByTelegramID(context.Background(), msg.From.ID)
		// if err != nil {
		// 	h.logger.Error("SuccessPayment.Handler: UserByTelegramID", slog.String("error", err.Error()), slog.Int64("telegram_id", msg.From.ID))
		// 	_, err = bot.SendMessage(&telego.SendMessageParams{
		// 		ChatID: telego.ChatID{ID: msg.Chat.ID},
		// 		Text:   use_message.GL.PaymentsUserNotFound(langCode),
		// 	})
		// 	if err != nil {
		// 		h.logger.Error("SuccessPayment.Handler: SendMessage", slog.String("error", err.Error()), slog.Int64("telegram_id", msg.From.ID))
		// 	}
		// 	return
		// }

		// Получаем ID транзакции из InvoicePayload
		txnID := entity.GetTransactionIDFromInvoicePayload(payment.InvoicePayload)
		if txnID == "" {
			h.logger.Error("SuccessPayment.Handler: GetTransactionIDFromInvoicePayload", slog.String("error", "txnID is 0"), slog.Int64("telegram_id", msg.From.ID))
			_, err := bot.SendMessage(&telego.SendMessageParams{
				ChatID: telego.ChatID{ID: msg.Chat.ID},
				Text:   use_message.GL.PaymentsTransactionError(langCode),
			})
			if err != nil {
				h.logger.Error("SuccessPayment.Handler: SendMessage", slog.String("error", err.Error()), slog.Int64("telegram_id", msg.From.ID))
			}
		}

		// Обновляем статус транзакции
		err := h.userUC.CompleteTransaction(context.Background(), txnID)
		if err != nil {
			h.logger.Error("SuccessPayment.Handler: CompleteTransaction", slog.String("error", err.Error()), slog.Int64("telegram_id", msg.From.ID))
			_, err = bot.SendMessage(&telego.SendMessageParams{
				ChatID: telego.ChatID{ID: msg.Chat.ID},
				Text:   use_message.GL.PaymentsTransactionError(langCode),
			})
			if err != nil {
				h.logger.Error("SuccessPayment.Handler: SendMessage", slog.String("error", err.Error()), slog.Int64("telegram_id", msg.From.ID))
			}
			return
		}

		// Отправляем сообщение об успешном пополнении
		_, err = bot.SendMessage(&telego.SendMessageParams{
			ChatID: telego.ChatID{ID: msg.Chat.ID},
			Text:   use_message.GL.PaymentsSuccess(langCode),
		})
		if err != nil {
			h.logger.Error("SuccessPayment.Handler: SendMessage", slog.String("error", err.Error()), slog.Int64("telegram_id", msg.From.ID))
		}
	}
}

func (h *SuccessPayment) Predicate() telegohandler.Predicate {
	return telegohandler.SuccessPayment()
}
