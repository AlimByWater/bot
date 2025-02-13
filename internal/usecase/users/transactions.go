package users

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"log/slog"
	"strconv"
)

func (m *Module) GetTransactionByID(ctx context.Context, txnID string) (entity.UserTransaction, error) {
	return m.repo.GetTransactionByID(ctx, txnID)
}

func (m *Module) CreateTransaction(ctx context.Context, txn entity.UserTransaction) (entity.UserTransaction, error) {
	return m.repo.CreateTransaction(ctx, txn)
}

func (m *Module) GetUserByID(ctx context.Context, userID int) (entity.User, error) {
	return m.repo.GetUserByID(ctx, userID)
}

func (m *Module) CreateBulkPendingDeposits(ctx context.Context, botID int64, telegramUserID int64, amounts []int, provider string) ([]entity.UserTransaction, error) {
	user, err := m.UserByTelegramID(ctx, telegramUserID)
	if err != nil {
		return nil, err
	}

	if botID > 0 {
		botID, _ = strconv.ParseInt(fmt.Sprintf("-100%d", botID), 10, 64)
	}

	var result []entity.UserTransaction
	for _, amount := range amounts {
		txn := entity.UserTransaction{
			UserID:   user.ID,
			Type:     entity.TransactionTypeDeposit,
			BotID:    botID,
			Amount:   amount,
			Provider: provider,
			Status:   entity.TransactionStatusPending,
		}
		created, err := m.repo.CreateTransaction(ctx, txn)
		if err != nil {
			return nil, err
		}
		result = append(result, created)
	}
	return result, nil
}

// computeBalanceChange вычисляет изменение баланса в зависимости от типа транзакции
func (m *Module) computeBalanceChange(txnType string, amount int) int {
	switch txnType {
	case entity.TransactionTypeDeposit, entity.TransactionTypeRefund,
		entity.TransactionTypePromoRedeem, entity.TransactionTypeReferralBonus,
		entity.TransactionTypePromoTransfer:
		return amount
	case entity.TransactionTypeWithdrawal:
		return -amount
	default:
		return 0
	}
}

// validateTransactionStatus проверяет, что статус транзакции допустимый
func (m *Module) validateTransactionStatus(status string) error {
	switch status {
	case entity.TransactionStatusPending,
		entity.TransactionStatusCompleted,
		entity.TransactionStatusFailed,
		entity.TransactionStatusExpired:
		return nil
	default:
		return fmt.Errorf("invalid transaction status: %s", status)
	}
}

// ProcessTransaction обрабатывает транзакцию, обновляя её статус и баланс пользователя в одной транзакции
func (m *Module) ProcessTransaction(ctx context.Context, txnID string, newStatus string) error {
	// Валидируем статус
	if err := m.validateTransactionStatus(newStatus); err != nil {
		return err
	}

	// Получаем транзакцию
	txn, err := m.repo.GetTransactionByID(ctx, txnID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	var balanceChange int
	if txn.Type == entity.TransactionTypeWithdrawal {
		if newStatus == entity.TransactionStatusCompleted {
			// Средства уже зарезервированы – оставляем без изменений.
			balanceChange = 0
		} else if newStatus == entity.TransactionStatusFailed || newStatus == entity.TransactionStatusExpired {
			// Возвращаем зарезервированные средства.
			balanceChange = txn.Amount
		} else {
			balanceChange = 0
		}
	} else {
		// Кредитные транзакции (Deposit, Refund, и т.п.)
		if newStatus == entity.TransactionStatusCompleted {
			balanceChange = m.computeBalanceChange(txn.Type, txn.Amount)
		} else {
			balanceChange = 0
		}
	}

	// Обрабатываем транзакцию
	return m.repo.ProcessTransaction(ctx, txnID, txn.UserID, balanceChange, newStatus)
}

// CompleteDepositTransaction завершает транзакцию депозита, устанавливая статус completed
func (m *Module) CompleteDepositTransaction(ctx context.Context, txnID string, externalID string) error {
	// Получаем транзакцию для проверки типа
	txn, err := m.repo.GetTransactionByID(ctx, txnID)
	if err != nil {
		return fmt.Errorf("failed to get transaction: %w", err)
	}

	// Проверяем, что это транзакция депозита
	if txn.Type != entity.TransactionTypeDeposit {
		return fmt.Errorf("invalid transaction type: expected deposit, got %s", txn.Type)
	}

	// Обрабатываем транзакцию
	err = m.ProcessTransaction(ctx, txnID, entity.TransactionStatusCompleted)
	if err != nil {
		return fmt.Errorf("failed to process transaction: %w", err)
	}

	go func() {
		err := m.repo.UpdateTransactionExternalID(context.Background(), txnID, externalID)
		if err != nil {
			m.logger.Error("Failed to update transaction external ID",
				slog.String("error", err.Error()),
				slog.String("txn_id", txnID),
				slog.String("external_id", externalID),
				slog.Int("user_id", txn.UserID),
			)
		}
	}()
	return nil
}
