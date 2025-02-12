package entity

import (
	"fmt"
	"github.com/valyala/fastjson"
	"time"
)

type UserTransaction struct {
	ID          string     `db:"id" json:"id"`
	UserID      int        `db:"user_id" json:"user_id"`
	Type        string     `db:"type" json:"type"`
	Amount      int        `db:"amount" json:"amount"`
	Status      string     `db:"status" json:"status"`
	Provider    string     `db:"provider" json:"provider"`
	ExternalID  string     `db:"external_id" json:"external_id"`
	ServiceID   *int       `db:"service_id" json:"service_id,omitempty"`
	BotID       int64      `db:"bot_id" json:"bot_id,omitempty"`
	Description string     `db:"description" json:"description"`
	PromoCode   *PromoCode `db:"promo_code" json:"promo_code,omitempty"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
}

type PromoCode struct {
	Code          string     `db:"code" json:"code"`
	Type          string     `db:"type" json:"type"`
	UserID        *int       `db:"user_id" json:"user_id,omitempty"`
	BonusRedeemer int        `db:"bonus_redeemer" json:"bonus_redeemer"`
	BonusReferrer int        `db:"bonus_referrer" json:"bonus_referrer"`
	UsageLimit    int        `db:"usage_limit" json:"usage_limit"`
	UsageCount    int        `db:"usage_count" json:"usage_count"`
	ValidFrom     time.Time  `db:"valid_from" json:"valid_from"`
	ValidTo       *time.Time `db:"valid_to" json:"valid_to,omitempty"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}

const (
	TransactionTypeDeposit       = "deposit"
	TransactionTypeWithdrawal    = "withdrawal"
	TransactionTypeRefund        = "refund"
	TransactionTypePromoRedeem   = "promo_redeem"
	TransactionTypeReferralBonus = "referral_bonus"
	TransactionTypePromoTransfer = "promo_transfer"
)

const (
	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
	TransactionStatusExpired   = "expired"
)

type PendingTransactionInput struct {
	Amount   int    `json:"amount"`
	Type     string `json:"type"`
	Provider string `json:"provider"`
}

func InvoicePayload(txnID string) string {
	return fmt.Sprintf(`{"txn_id": "%s"}`, txnID)
}

func GetTransactionIDFromInvoicePayload(payload string) string {
	return fastjson.GetString([]byte(payload), "txn_id")
}
