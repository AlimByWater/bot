package entity

import (
	"encoding/json"
	"time"
)

// TransactionAuditEvent описывает формат сохраняемого JSON события.
type TransactionAuditEvent struct {
	Version int                     `json:"version"` // версия схемы JSON
	Data    TransactionAuditPayload `json:"data"`
}

// TransactionAuditPayload содержит детали транзакции.
type TransactionAuditPayload struct {
	TransactionID string    `json:"transaction_id"`
	UserID        int       `json:"user_id"`
	Type          string    `json:"type"`
	Amount        int       `json:"amount"`
	Status        string    `json:"status"`
	Provider      string    `json:"provider"`
	ExternalID    string    `json:"external_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TransactionAudit – структура для передачи в репозиторий аудита.
// Поле Payload хранит уже замаршированный JSON, а Version дублирует версию.
type TransactionAudit struct {
	TransactionID string          `json:"transaction_id"`
	Payload       json.RawMessage `json:"payload"`
	Version       int             `json:"version"`
	EventTime     time.Time       `json:"event_time"`
}

// DefaultAuditEventVersion – версия схемы JSON для событий аудита.
const DefaultAuditEventVersion = 1

// NewTransactionAuditEvent создаёт событие аудита, используя явные значения параметров.
// Версия события выставляется автоматически из константы DefaultAuditEventVersion.
func NewTransactionAuditEvent(
	txnID string,
	userID int,
	txnType string,
	amount int,
	status, provider, externalID string,
	createdAt, updatedAt time.Time,
) TransactionAuditEvent {
	return TransactionAuditEvent{
		Version: DefaultAuditEventVersion,
		Data: TransactionAuditPayload{
			TransactionID: txnID,
			UserID:        userID,
			Type:          txnType,
			Amount:        amount,
			Status:        status,
			Provider:      provider,
			ExternalID:    externalID,
			CreatedAt:     createdAt,
			UpdatedAt:     updatedAt,
		},
	}
}

// NewTransactionAuditEventFromTransaction создаёт событие аудита, используя данные из переданной транзакции.
// Версия также устанавливается автоматически.
func NewTransactionAuditEventFromTransaction(txn UserTransaction) TransactionAuditEvent {
	return TransactionAuditEvent{
		Version: DefaultAuditEventVersion,
		Data: TransactionAuditPayload{
			TransactionID: txn.ID,
			UserID:        txn.UserID,
			Type:          txn.Type,
			Amount:        txn.Amount,
			Status:        txn.Status,
			Provider:      txn.Provider,
			ExternalID:    txn.ExternalID,
			CreatedAt:     txn.CreatedAt,
			UpdatedAt:     txn.UpdatedAt,
		},
	}
}
