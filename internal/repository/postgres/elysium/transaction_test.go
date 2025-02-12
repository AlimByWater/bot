package elysium_test

import (
	"context"
	"elysium/internal/entity"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_Transactions_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	user := createTestUser(t, ctx, 900000003030303)

	t.Cleanup(func() {
		err := elysiumRepo.DeleteUser(ctx, user.ID)
		require.NoError(t, err)
	})

	// Тестируем создание транзакции со статусом pending — баланс меняться не должен.
	pendingTxn := entity.UserTransaction{
		UserID:      user.ID,
		Type:        entity.TransactionTypeDeposit,
		Amount:      1000,
		Status:      entity.TransactionStatusPending,
		Provider:    "test_provider",
		ExternalID:  "ext123",
		Description: "Pending deposit",
	}
	createdTxn, err := elysiumRepo.CreateTransaction(ctx, pendingTxn)
	require.NoError(t, err)
	require.NotEmpty(t, createdTxn.ID)

	// Баланс не должен измениться, так как статус pending.
	balance, err := elysiumRepo.GetUserBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, balance)

	// Обрабатываем транзакцию: обновляем статус и баланс пользователя в одной атомарной транзакции.
	err = elysiumRepo.ProcessTransaction(ctx, createdTxn.ID, user.ID, 1000, entity.TransactionStatusCompleted)
	require.NoError(t, err)

	balance, err = elysiumRepo.GetUserBalance(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 1000, balance)

	// Получаем транзакцию по её ID.
	txnByID, err := elysiumRepo.GetTransactionByID(ctx, createdTxn.ID)
	require.NoError(t, err)
	assert.Equal(t, createdTxn.ID, txnByID.ID)
	assert.Equal(t, entity.TransactionStatusCompleted, txnByID.Status)

	// Получаем список транзакций для пользователя.
	txns, err := elysiumRepo.GetTransactionsByUserID(ctx, user.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(txns))
}

func TestRepository_Transactions_PromoCode(t *testing.T) {
	ctx := context.Background()
	user := createTestUser(t, ctx, 900000003030305)

	// Создаем промокод с помощью нового метода
	promoCode := "PROMO123"
	now := time.Now()
	promo := entity.PromoCode{
		Code:          promoCode,
		Type:          "free_use",
		BonusRedeemer: 10,
		BonusReferrer: 5,
		UsageLimit:    1,
		UsageCount:    0,
		ValidFrom:     now,
	}

	_, err := elysiumRepo.CreatePromoCode(ctx, promo)
	require.NoError(t, err)

	// Создаем транзакцию с использованием промокода.
	txn := entity.UserTransaction{
		UserID:      user.ID,
		Type:        entity.TransactionTypePromoRedeem,
		Amount:      300,
		Status:      entity.TransactionStatusCompleted,
		Provider:    "test_provider",
		ExternalID:  "ext126",
		Description: "Promo redeem transaction",
		PromoCode: &entity.PromoCode{
			Code: promoCode,
		},
	}
	createdTxn, err := elysiumRepo.CreateTransaction(ctx, txn)
	require.NoError(t, err)
	require.NotEmpty(t, createdTxn.ID)

	// Получаем транзакцию и проверяем, что поле PromoCode заполнено.
	txnByID, err := elysiumRepo.GetTransactionByID(ctx, createdTxn.ID)
	require.NoError(t, err)
	require.NotNil(t, txnByID.PromoCode)
	assert.Equal(t, promoCode, txnByID.PromoCode.Code)
}
