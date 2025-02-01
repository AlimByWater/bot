package entity

import "time"

const (
	TransactionTypeDeposit    = "deposit"
	TransactionTypeWithdrawal = "withdrawal"
	TransactionTypeRefund     = "refund"

	TransactionStatusPending   = "pending"
	TransactionStatusCompleted = "completed"
	TransactionStatusFailed    = "failed"
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
	ID           int       `json:"id" db:"id"`
	UserID       int       `json:"user_id" db:"user_id"`
	Type         string    `json:"type" db:"type"`
	Amount       int       `json:"amount" db:"amount"`
	Status       string    `json:"status" db:"status"`
	Provider     string    `json:"provider" db:"provider"`
	ExternalID   string    `json:"external_id" db:"external_id"`
	ServiceID    *int      `json:"service_id,omitempty" db:"service_id"`
	BotID        *int64    `json:"bot_id,omitempty" db:"bot_id"`
	BalanceAfter int       `json:"balance_after" db:"balance_after"`
	Description  string    `json:"description,omitempty" db:"description"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
