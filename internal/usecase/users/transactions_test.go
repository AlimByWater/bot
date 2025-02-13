package users_test

import (
	"context"
	"elysium/internal/entity"
	"github.com/stretchr/testify/require"
	"testing"
)

// Создаем тестового пользователя и регистрируем очистку после теста.
func createTestUser(t *testing.T) entity.User {
	testUser := entity.User{
		TelegramID:       123456789,
		TelegramUsername: "test_user",
		Firstname:        "Test",
	}
	createdUser, err := module.CreateOrUpdateUser(context.Background(), testUser)
	require.NoError(t, err, "Не удалось создать тестового пользователя")
	require.NotEmpty(t, createdUser.ID, "Не удалось создать тестового пользователя")
	t.Logf("Created user with ID %d", createdUser.ID)

	return createdUser
}

func TestWithdrawalTransactionCompleted(t *testing.T) {
	// Создаём контекст и тестовое окружение.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем тестового пользователя
	testUser := createTestUser(t)

	// Шаг 1. Пополняем баланс пользователя депозитной транзакцией Completed.
	depositTxn := entity.UserTransaction{
		UserID:   testUser.ID,
		Type:     entity.TransactionTypeDeposit,
		Amount:   200,
		Status:   entity.TransactionStatusCompleted,
		Provider: "test_provider",
	}
	dep, err := module.CreateTransaction(ctx, depositTxn)
	t.Logf("Created deposit transaction with ID %s", dep.ID)
	require.NoError(t, err, "Не удалось создать депозитную транзакцию")

	// Шаг 2. Создаем транзакцию списания (withdrawal) на сумму 100 в статусе Pending (резервирование).
	withdrawalTxn := entity.UserTransaction{
		UserID:   dep.UserID,
		Type:     entity.TransactionTypeWithdrawal,
		Amount:   100,
		Status:   entity.TransactionStatusPending,
		Provider: "test_provider",
		BotID:    dep.BotID,
	}
	txn, err := module.CreateTransaction(ctx, withdrawalTxn)
	require.NoError(t, err, "Не удалось создать транзакцию списания")

	// Шаг 3. Обрабатываем транзакцию, переводя ее в статус Completed.
	err = module.ProcessTransaction(ctx, txn.ID, entity.TransactionStatusCompleted)
	require.NoError(t, err, "Не удалось завершить транзакцию списания")

	// Шаг 4. Проверяем, что статус транзакции обновлен на Completed.
	updatedTxn, err := module.GetTransactionByID(ctx, txn.ID)
	require.NoError(t, err, "Не удалось получить транзакцию после обработки")
	require.Equal(t, entity.TransactionStatusCompleted, updatedTxn.Status, "Статус транзакции должен быть Completed")
}

func TestWithdrawalTransactionFailedAndFundsRestored(t *testing.T) {
	// Создаём контекст и поднимаем тестовое окружение.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем тестового пользователя
	testUser := createTestUser(t)

	// Шаг 1. Пополняем баланс пользователя депозитной транзакцией Completed.
	depositTxn := entity.UserTransaction{
		UserID:   testUser.ID,
		Type:     entity.TransactionTypeDeposit,
		Amount:   150,
		Status:   entity.TransactionStatusCompleted,
		Provider: "test_provider",
	}
	dep, err := module.CreateTransaction(ctx, depositTxn)
	require.NoError(t, err, "Не удалось создать депозитную транзакцию")

	// Шаг 2. Создаем транзакцию списания на сумму 100 (резервирование).
	withdrawalTxn := entity.UserTransaction{
		UserID:   dep.UserID,
		Type:     entity.TransactionTypeWithdrawal,
		Amount:   100,
		Status:   entity.TransactionStatusPending,
		Provider: "test_provider",
		BotID:    dep.BotID,
	}
	txn, err := module.CreateTransaction(ctx, withdrawalTxn)
	require.NoError(t, err, "Не удалось создать транзакцию списания")

	// Шаг 3. Обрабатываем транзакцию, переводя ее в статус Failed — средства должны вернуться.
	err = module.ProcessTransaction(ctx, txn.ID, entity.TransactionStatusFailed)
	require.NoError(t, err, "Не удалось обработать транзакцию со статусом Failed")

	// Шаг 4. Проверяем, что статус транзакции обновлен на Failed.
	updatedTxn, err := module.GetTransactionByID(ctx, txn.ID)
	require.NoError(t, err, "Не удалось получить транзакцию после обработки")
	require.Equal(t, entity.TransactionStatusFailed, updatedTxn.Status, "Статус транзакции должен быть Failed")

	// Шаг 5. После возврата средств создаем новую транзакцию списания на 100.
	newWithdrawalTxn := entity.UserTransaction{
		UserID:   dep.UserID,
		Type:     entity.TransactionTypeWithdrawal,
		Amount:   100,
		Status:   entity.TransactionStatusPending,
		Provider: "test_provider",
		BotID:    dep.BotID,
	}
	txn2, err := module.CreateTransaction(ctx, newWithdrawalTxn)
	require.NoError(t, err, "После возврата средств новая транзакция списания должна пройти успешно")
	t.Logf("Новая транзакция после отмены предыдущей успешно создана: %+v", txn2)
}

func TestMultipleServiceUsageInsufficientFunds(t *testing.T) {
	// Создаем контекст и поднимаем тестовую среду (функции setupTest и testUsersConfig уже реализованы в онлайн-тесте)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем тестового пользователя
	testUser := createTestUser(t)
	t.Cleanup(func() {
		err := module.DeleteUserHard(context.Background(), t, testUser.ID)
		t.Logf("Deleted user with ID %d", testUser.ID)
		require.NoError(t, err, "Не удалось удалить тестового пользователя")
	})

	// Первым делом пополним баланс пользователя (пополняем депозитной транзакцией в статусе Completed)
	depositTxn := entity.UserTransaction{
		UserID:   testUser.ID,
		Type:     entity.TransactionTypeDeposit,
		Amount:   100, // баланс пополняется до 100
		Status:   entity.TransactionStatusCompleted,
		BotID:    -1007894673045,
		Provider: "test_provider",
	}
	depositResult, err := module.CreateTransaction(ctx, depositTxn)
	require.NoError(t, err, "Не удалось создать депозитную транзакцию для пополнения баланса")

	// Теперь пытаемся списать средства три раза подряд с суммой 100 (услуга стоит 100)
	withdrawalTxn := entity.UserTransaction{
		UserID:   depositResult.UserID,
		Type:     entity.TransactionTypeWithdrawal,
		Amount:   100,
		BotID:    -1007894673045,
		Status:   entity.TransactionStatusPending, // на стадии Pending происходит резервирование (вычитание)
		Provider: "test_provider",
	}

	// Первая транзакция должна пройти успешно и зарезервировать 100
	txn1, err := module.CreateTransaction(ctx, withdrawalTxn)
	require.NoError(t, err, "Первая транзакция списания должна пройти успешно")
	t.Logf("Создана транзакция: %+v", txn1)

	// Вторая транзакция должна вернуть ошибку: оставшийся баланс равен 0,
	// а попытка списать ещё 100 приводит к отрицательному балансу.
	_, err = module.CreateTransaction(ctx, withdrawalTxn)
	require.Error(t, err, "Вторая транзакция списания должна провалиться из-за недостатка средств")
	require.Contains(t, err.Error(), "insufficient funds", "Ожидаем ошибку 'insufficient funds'")

	// Третья транзакция также должна вернуть ошибку
	_, err = module.CreateTransaction(ctx, withdrawalTxn)
	require.Error(t, err, "Третья транзакция списания должна провалиться из-за недостатка средств")
	require.Contains(t, err.Error(), "insufficient funds", "Ожидаем ошибку 'insufficient funds'")
}
