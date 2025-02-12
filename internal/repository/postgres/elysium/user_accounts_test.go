package elysium_test

import (
	"context"
	"elysium/internal/entity"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_UserAccounts(t *testing.T) {
	ctx := context.Background()

	// Создаем тестового пользователя
	user := entity.User{
		TelegramID:       123456789,
		TelegramUsername: "test_user",
	}
	createdUser, err := elysiumRepo.CreateOrUpdateUser(ctx, user)
	require.NoError(t, err)

	t.Run("GetUserAccount - новый аккаунт", func(t *testing.T) {
		account, err := elysiumRepo.GetUserAccount(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.Equal(t, createdUser.ID, account.UserID)
		assert.Equal(t, 0, account.Balance)
		assert.False(t, account.CreatedAt.IsZero())
		assert.False(t, account.UpdatedAt.IsZero())
	})

	t.Run("UpdateUserBalance", func(t *testing.T) {
		// Обновляем баланс
		err := elysiumRepo.UpdateUserBalance(ctx, createdUser.ID, 1000)
		require.NoError(t, err)

		// Проверяем обновленный баланс
		account, err := elysiumRepo.GetUserAccount(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.Equal(t, 1000, account.Balance)
	})

	t.Run("GetUserBalance", func(t *testing.T) {
		balance, err := elysiumRepo.GetUserBalance(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.Equal(t, 1000, balance)
	})

	t.Run("UpdateUserBalance - отрицательный баланс", func(t *testing.T) {
		// Пытаемся установить отрицательный баланс
		err := elysiumRepo.UpdateUserBalance(ctx, createdUser.ID, -100)
		require.Error(t, err) // Должна быть ошибка из-за CHECK constraint

		// Проверяем, что баланс не изменился
		balance, err := elysiumRepo.GetUserBalance(ctx, createdUser.ID)
		require.NoError(t, err)
		assert.Equal(t, 1000, balance)
	})

	t.Run("GetUserAccount - несуществующий пользователь", func(t *testing.T) {
		account, err := elysiumRepo.GetUserAccount(ctx, 999999)
		require.NoError(t, err) // Не должно быть ошибки, т.к. создается новый аккаунт
		assert.Equal(t, 999999, account.UserID)
		assert.Equal(t, 0, account.Balance)
	})

	t.Run("CreateUserAccount - дублирование", func(t *testing.T) {
		account := entity.UserAccount{
			UserID:    createdUser.ID,
			Balance:   500,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := elysiumRepo.CreateUserAccount(ctx, account)
		require.Error(t, err) // Должна быть ошибка из-за unique constraint
	})
}
