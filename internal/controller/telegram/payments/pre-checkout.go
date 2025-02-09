package payments

import (
	"context"
	"elysium/internal/entity"
	"elysium/internal/usecase/use_message"
	"log/slog"
	"reflect"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

// Updated PreCheckout struct with additional method GetTransactionByID
type PreCheckout struct {
	logger *slog.Logger
	userUC interface {
		GetTransactionByID(ctx context.Context, txnID string) (entity.UserTransaction, error)
	}
}

func NewPreCheckout(userUC interface {
	GetTransactionByID(ctx context.Context, txnID string) (entity.UserTransaction, error)
}) *PreCheckout {
	return &PreCheckout{
		userUC: userUC,
	}
}

func (h *PreCheckout) AddLogger(logger *slog.Logger) {
	h.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(h)).Type().PkgPath()))
}

func (h *PreCheckout) Command() string { return "" }

func (h *PreCheckout) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		query := update.PreCheckoutQuery
		langCode := query.From.LanguageCode
		var txID string
		if query.InvoicePayload != "" {
			txID = entity.GetTransactionIDFromInvoicePayload(query.InvoicePayload)
			if txID == "" {
				h.logger.Error("PreCheckout.Handler: txn_id not found in payload", slog.String("payload", query.InvoicePayload))
				_ = bot.AnswerPreCheckoutQuery(&telego.AnswerPreCheckoutQueryParams{
					PreCheckoutQueryID: query.ID,
					Ok:                 false,
					ErrorMessage:       use_message.GL.PaymentsTransactionError(langCode),
				})
				return
			}
		}

		// Если в payload присутствует txn_id, проверяем статус транзакции
		if txID != "" {
			txn, err := h.userUC.GetTransactionByID(context.Background(), txID)
			if err != nil {
				h.logger.Error("PreCheckout.Handler: GetTransactionByID", slog.String("error", err.Error()))
				_ = bot.AnswerPreCheckoutQuery(&telego.AnswerPreCheckoutQueryParams{
					PreCheckoutQueryID: query.ID,
					Ok:                 false,
					ErrorMessage:       use_message.GL.PaymentsTransactionError(langCode),
				})
				return
			}
			if txn.Status == "failed" || txn.Status == "expired" {
				h.logger.Error("PreCheckout.Handler: transaction status invalid", slog.String("status", txn.Status))
				_ = bot.AnswerPreCheckoutQuery(&telego.AnswerPreCheckoutQueryParams{
					PreCheckoutQueryID: query.ID,
					Ok:                 false,
					ErrorMessage:       use_message.GL.PaymentsTransactionError(langCode),
				})
				return
			}
			// Если статус транзакции корректен, отвечаем успехом
			_ = bot.AnswerPreCheckoutQuery(&telego.AnswerPreCheckoutQueryParams{
				PreCheckoutQueryID: query.ID,
				Ok:                 true,
			})
			return
		}
	}
}

func (h *PreCheckout) Predicate() telegohandler.Predicate {
	return telegohandler.AnyPreCheckoutQuery()
}
