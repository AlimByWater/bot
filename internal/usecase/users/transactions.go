package users

import (
	"context"
	"elysium/internal/entity"
)

func (m *Module) CreatePendingTransaction(ctx context.Context, userID int, amount int, provider string) (entity.UserTransaction, error) {
	txn := entity.UserTransaction{
		UserID:   userID,
		Type:     entity.TransactionTypeDeposit,
		Amount:   amount,
		Status:   entity.TransactionStatusPending,
		Provider: provider,
	}

	return m.repo.CreateTransaction(ctx, txn)
}

func (m *Module) CompleteTransaction(ctx context.Context, txnID string) error {
	return m.repo.UpdateTransactionStatus(ctx, txnID, entity.TransactionStatusCompleted)
}

func (m *Module) GetTransactionByID(ctx context.Context, txnID string) (entity.UserTransaction, error) {
	return m.repo.GetTransactionByID(ctx, txnID)
}

func (m *Module) CreateBulkPendingTransactions(ctx context.Context, telegramUserID int64, amounts []int, provider string) ([]entity.UserTransaction, error) {
	user, err := m.UserByTelegramID(ctx, telegramUserID)
	if err != nil {
		return nil, err
	}

	transactions := make([]entity.PendingTransactionInput, len(amounts))
	for i, amount := range amounts {
		transactions[i] = entity.PendingTransactionInput{
			Amount:   amount,
			Type:     entity.TransactionTypeDeposit,
			Provider: provider,
		}
	}

	return m.repo.CreatePendingTransactions(ctx, user.ID, transactions)
}
