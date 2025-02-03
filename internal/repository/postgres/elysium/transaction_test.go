package elysium_test

import (
	"context"
	"elysium/internal/entity"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCreateTransaction(t *testing.T) {
	t.Skip("Skipping test")
	// Создаем тестового пользователя
	user := entity.User{
		TelegramID:       9000000011,
		TelegramUsername: "txn_test_user",
		Firstname:        "Transaction",
		Balance:          1000,
	}
	createdUser, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user)
	require.NoError(t, err)
	defer func() {
		err := elysiumRepo.DeleteUser(context.Background(), createdUser.ID)
		require.NoError(t, err)
	}()

	testCases := []struct {
		name          string
		txn           entity.UserTransaction
		expectedError bool
		checkBalance  int
	}{
		{
			name: "Successful deposit",
			txn: entity.UserTransaction{
				UserID:       createdUser.ID,
				Type:         entity.TransactionTypeDeposit,
				Amount:       500,
				Status:       entity.TransactionStatusCompleted,
				BalanceAfter: 1500,
			},
			checkBalance: 1500,
		},
		{
			name: "Withdrawal with provider",
			txn: entity.UserTransaction{
				UserID:       createdUser.ID,
				Type:         entity.TransactionTypeWithdrawal,
				Amount:       200,
				Status:       entity.TransactionStatusCompleted,
				Provider:     "paypal",
				ExternalID:   "PAYID-123",
				BalanceAfter: 1300,
			},
			checkBalance: 1300,
		},
		{
			name: "Invalid transaction type",
			txn: entity.UserTransaction{
				UserID: createdUser.ID,
				Type:   "invalid_type",
				Amount: 100,
				Status: entity.TransactionStatusCompleted,
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultTxn, err := elysiumRepo.CreateTransaction(context.Background(), tc.txn)

			if tc.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotZero(t, resultTxn.ID)
			require.False(t, resultTxn.CreatedAt.IsZero())

			// Проверяем обновленный баланс
			updatedUser, err := elysiumRepo.GetUserByID(context.Background(), createdUser.ID)
			require.NoError(t, err)
			require.Equal(t, tc.checkBalance, updatedUser.Balance)
		})
	}
}

func TestGetTransactionsByUserID(t *testing.T) {
	// Создаем тестового пользователя
	user := entity.User{
		TelegramID:       9000000012,
		TelegramUsername: "txn_test_user2",
		Firstname:        "Transaction2",
		Balance:          2000,
	}
	createdUser, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user)
	require.NoError(t, err)
	defer func() {
		err := elysiumRepo.DeleteUser(context.Background(), createdUser.ID)
		require.NoError(t, err)
	}()

	// Создаем тестовые транзакции
	txns := []entity.UserTransaction{
		{
			UserID: createdUser.ID,
			Type:   entity.TransactionTypeDeposit,
			Amount: 1000,
			Status: entity.TransactionStatusCompleted,
		},
		{
			UserID:     createdUser.ID,
			Type:       entity.TransactionTypeWithdrawal,
			Amount:     500,
			Status:     entity.TransactionStatusCompleted,
			Provider:   "stripe",
			ExternalID: "ch_123",
		},
	}

	for i := range txns {
		_, err := elysiumRepo.CreateTransaction(context.Background(), txns[i])
		require.NoError(t, err)
		if i == 0 {
			time.Sleep(1 * time.Millisecond)
		}
	}

	// Получаем транзакции
	retrievedTxns, err := elysiumRepo.GetTransactionsByUserID(context.Background(), createdUser.ID)
	require.NoError(t, err)
	require.Len(t, retrievedTxns, 2)

	// Проверяем данные первой транзакции
	require.Equal(t, entity.TransactionTypeWithdrawal, retrievedTxns[0].Type)
	require.Equal(t, 500, retrievedTxns[0].Amount)
	require.Equal(t, entity.TransactionStatusCompleted, retrievedTxns[0].Status)
	require.Equal(t, "stripe", retrievedTxns[0].Provider)
	require.Equal(t, "ch_123", retrievedTxns[0].ExternalID)

	// Проверяем данные второй транзакции
	require.Equal(t, entity.TransactionTypeDeposit, retrievedTxns[1].Type)
	require.Equal(t, 1000, retrievedTxns[1].Amount)
	require.Equal(t, entity.TransactionStatusCompleted, retrievedTxns[1].Status)
}
