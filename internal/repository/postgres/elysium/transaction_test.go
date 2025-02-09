package elysium_test

import (
	"context"
	"elysium/internal/entity"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCreateTransaction(t *testing.T) {
	//t.Skip("Skipping test")
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
			name: "Successful refund",
			txn: entity.UserTransaction{
				UserID:       createdUser.ID,
				Type:         entity.TransactionTypeRefund,
				Amount:       150,
				Status:       entity.TransactionStatusCompleted,
				Provider:     "stripe",
				ExternalID:   "ref_123",
				Description:  "Refund for order #123",
				BalanceAfter: 1450,
			},
			checkBalance: 1450,
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
		{
			name: "Zero amount transaction",
			txn: entity.UserTransaction{
				UserID: createdUser.ID,
				Type:   entity.TransactionTypeDeposit,
				Amount: 0,
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

func TestCreatePendingTransactions(t *testing.T) {
	// Создаем тестового пользователя
	user := entity.User{
		TelegramID:       9000000013,
		TelegramUsername: "txn_test_user3",
		Firstname:        "Transaction3",
		Balance:          3000,
	}
	createdUser, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user)
	require.NoError(t, err)
	defer func() {
		err := elysiumRepo.DeleteUser(context.Background(), createdUser.ID)
		require.NoError(t, err)
	}()

	testCases := []struct {
		name          string
		transactions  []entity.PendingTransactionInput
		expectedLen   int
		expectedError bool
	}{
		{
			name: "Multiple pending deposits",
			transactions: []entity.PendingTransactionInput{
				{
					Amount:   100,
					Type:     entity.TransactionTypeDeposit,
					Provider: "telegram",
				},
				{
					Amount:   200,
					Type:     entity.TransactionTypeDeposit,
					Provider: "telegram",
				},
			},
			expectedLen: 2,
		},
		{
			name:         "Empty transactions slice",
			transactions: []entity.PendingTransactionInput{},
			expectedLen:  0,
		},
		{
			name: "Single pending withdrawal",
			transactions: []entity.PendingTransactionInput{
				{
					Amount:   300,
					Type:     entity.TransactionTypeWithdrawal,
					Provider: "stripe",
				},
			},
			expectedLen: 1,
		},
		{
			name: "Invalid transaction type",
			transactions: []entity.PendingTransactionInput{
				{
					Amount:   100,
					Type:     "invalid_type",
					Provider: "telegram",
				},
			},
			expectedError: true,
		},
		{
			name: "Zero amount transaction",
			transactions: []entity.PendingTransactionInput{
				{
					Amount:   0,
					Type:     entity.TransactionTypeDeposit,
					Provider: "telegram",
				},
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := elysiumRepo.CreatePendingTransactions(context.Background(), createdUser.ID, tc.transactions)

			if tc.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, result, tc.expectedLen)

			if tc.expectedLen > 0 {
				for i, txn := range result {
					require.NotEmpty(t, txn.ID)
					require.Equal(t, createdUser.ID, txn.UserID)
					require.Equal(t, entity.TransactionStatusPending, txn.Status)
					require.Equal(t, tc.transactions[i].Amount, txn.Amount)
					require.False(t, txn.CreatedAt.IsZero())
					require.False(t, txn.UpdatedAt.IsZero())
				}

				// Проверяем, что баланс не изменился (так как транзакции в статусе pending)
				updatedUser, err := elysiumRepo.GetUserByID(context.Background(), createdUser.ID)
				require.NoError(t, err)
				require.Equal(t, user.Balance, updatedUser.Balance)
			}
		})
	}
}

func TestUpdateTransactionStatus(t *testing.T) {
	// Создаем тестового пользователя
	user := entity.User{
		TelegramID:       9000000014,
		TelegramUsername: "txn_test_user4",
		Firstname:        "Transaction4",
		Balance:          4000,
	}
	createdUser, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user)
	require.NoError(t, err)
	defer func() {
		err := elysiumRepo.DeleteUser(context.Background(), createdUser.ID)
		require.NoError(t, err)
	}()

	testCases := []struct {
		name             string
		initialTxn       entity.UserTransaction
		newStatus        string
		expectedError    bool
		checkBalance     int
		skipBalanceCheck bool
	}{
		{
			name: "Complete pending deposit",
			initialTxn: entity.UserTransaction{
				UserID:   createdUser.ID,
				Type:     entity.TransactionTypeDeposit,
				Amount:   500,
				Status:   entity.TransactionStatusPending,
				Provider: "telegram",
			},
			newStatus:    entity.TransactionStatusCompleted,
			checkBalance: 4500,
		},
		{
			name: "Fail pending withdrawal",
			initialTxn: entity.UserTransaction{
				UserID:   createdUser.ID,
				Type:     entity.TransactionTypeWithdrawal,
				Amount:   1000,
				Status:   entity.TransactionStatusPending,
				Provider: "stripe",
			},
			newStatus:    entity.TransactionStatusFailed,
			checkBalance: 4500, // Баланс не должен измениться
		},
		{
			name: "Expire pending deposit",
			initialTxn: entity.UserTransaction{
				UserID:   createdUser.ID,
				Type:     entity.TransactionTypeDeposit,
				Amount:   300,
				Status:   entity.TransactionStatusPending,
				Provider: "telegram",
			},
			newStatus:    entity.TransactionStatusExpired,
			checkBalance: 4500, // Баланс не должен измениться
		},
		{
			name: "Complete pending refund",
			initialTxn: entity.UserTransaction{
				UserID:      createdUser.ID,
				Type:        entity.TransactionTypeRefund,
				Amount:      200,
				Status:      entity.TransactionStatusPending,
				Provider:    "stripe",
				ExternalID:  "ref_456",
				Description: "Refund for failed service",
			},
			newStatus:    entity.TransactionStatusCompleted,
			checkBalance: 4700, // Баланс должен увеличиться
		},
		{
			name: "Invalid transaction ID",
			initialTxn: entity.UserTransaction{
				ID:     "non-existent-id",
				UserID: createdUser.ID,
				Type:   entity.TransactionTypeDeposit,
				Amount: 100,
				Status: entity.TransactionStatusPending,
			},
			newStatus:        entity.TransactionStatusCompleted,
			expectedError:    true,
			skipBalanceCheck: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Создаем начальную транзакцию
			if tc.initialTxn.ID != "non-existent-id" {
				createdTxn, err := elysiumRepo.CreateTransaction(context.Background(), tc.initialTxn)
				require.NoError(t, err)
				tc.initialTxn.ID = createdTxn.ID
			}

			// Обновляем статус
			err := elysiumRepo.UpdateTransactionStatus(context.Background(), tc.initialTxn.ID, tc.newStatus)

			if tc.expectedError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if !tc.skipBalanceCheck {
				// Проверяем обновленный баланс
				updatedUser, err := elysiumRepo.GetUserByID(context.Background(), createdUser.ID)
				require.NoError(t, err)
				require.Equal(t, tc.checkBalance, updatedUser.Balance)
			}

			// Проверяем обновленную транзакцию
			txns, err := elysiumRepo.GetTransactionsByUserID(context.Background(), createdUser.ID)
			require.NoError(t, err)
			require.NotEmpty(t, txns)

			var updatedTxn *entity.UserTransaction
			for i := range txns {
				if txns[i].ID == tc.initialTxn.ID {
					updatedTxn = &txns[i]
					break
				}
			}

			require.NotNil(t, updatedTxn)
			require.Equal(t, tc.newStatus, updatedTxn.Status)
			require.True(t, updatedTxn.UpdatedAt.After(updatedTxn.CreatedAt))
		})
	}
}
