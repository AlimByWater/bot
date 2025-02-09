package entity

import (
	"fmt"
	"time"

	"github.com/valyala/fastjson"
)

const (
	TransactionTypeDeposit    = "deposit"
	TransactionTypeWithdrawal = "withdrawal"
	TransactionTypeRefund     = "refund"

	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
	TransactionStatusExpired   = "expired"
)

// easyjson:json
type Service struct {
	ID          int       `json:"id" db:"id"`
	BotID       int64     `json:"bot_id" db:"bot_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       int       `json:"price" db:"price"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// easyjson:json
type UserTransaction struct {
	ID           string    `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	Type         string    `json:"type" db:"type"`
	Amount       int       `json:"amount" db:"amount"`
	Status       string    `json:"status" db:"status"`
	Provider     string    `json:"provider" db:"provider"`
	ExternalID   string    `json:"external_id" db:"external_id"`
	PromoCodeID  *string   `json:"promo_code_id,omitempty" db:"promo_code_id"`
	ServiceID    *int      `json:"service_id,omitempty" db:"service_id"`
	BotID        *int64    `json:"bot_id,omitempty" db:"bot_id"`
	BalanceAfter int       `json:"balance_after" db:"balance_after"`
	Description  string    `json:"description,omitempty" db:"description"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type PendingTransactionInput struct {
	Amount   int    `json:"amount"`
	Type     string `json:"type"`
	Provider string `json:"provider"`
}

type PromoCode struct {
	Code          string     `json:"code" db:"code"`
	Type          string     `json:"type" db:"type"`
	UserID        *int       `json:"user_id,omitempty" db:"user_id"`
	BonusRedeemer int        `json:"bonus_redeemer" db:"bonus_redeemer"`
	BonusReferrer int        `json:"bonus_referrer" db:"bonus_referrer"`
	UsageLimit    int        `json:"usage_limit" db:"usage_limit"`
	UsageCount    int        `json:"usage_count" db:"usage_count"`
	ValidFrom     time.Time  `json:"valid_from" db:"valid_from"`
	ValidTo       *time.Time `json:"valid_to,omitempty" db:"valid_to"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

type PromoCodeUsage struct {
	ID          int       `json:"id" db:"id"`
	PromoCodeID string    `json:"promo_code_id" db:"promo_code_id"`
	UserID      int       `json:"user_id" db:"user_id"`
	UsedAt      time.Time `json:"used_at" db:"used_at"`
}

func InvoicePayload(txnID string) string {
	return fmt.Sprintf(`{"txn_id": "%s"}`, txnID)
}

func GetTransactionIDFromInvoicePayload(payload string) string {
	return fastjson.GetString([]byte(payload), "txn_id")
}
