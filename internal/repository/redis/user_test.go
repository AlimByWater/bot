package redis_test

import (
	"context"
	"elysium/internal/entity"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSaveOrUpdateUserCache(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	user := entity.User{
		ID:               73456789,
		TelegramID:       123456789,
		TelegramUsername: "testuser",
		Firstname:        "Test",
	}

	err := redisModule.SaveOrUpdateUserCache(context.Background(), user)
	require.NoError(t, err)

	t.Run("Save new user to cache", func(t *testing.T) {
		cachedUser, err := redisModule.GetUserByTelegramIDCache(context.Background(), user.TelegramID)
		require.NoError(t, err)
		require.Equal(t, user.TelegramID, cachedUser.TelegramID)
		require.Equal(t, user.TelegramUsername, cachedUser.TelegramUsername)
	})
}

func TestGetUserByTelegramIDCache(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	telegramID := int64(223456789)
	_, err := redisModule.GetUserByTelegramIDCache(context.Background(), telegramID)
	require.Error(t, err)
	require.Equal(t, entity.ErrUserNotFound, err)

	user := entity.User{
		ID:               83456789,
		TelegramID:       telegramID,
		TelegramUsername: "testuser",
		Firstname:        "Test",
	}
	err = redisModule.SaveOrUpdateUserCache(context.Background(), user)
	require.NoError(t, err)

	t.Run("Get user from cache", func(t *testing.T) {
		cachedUser, err := redisModule.GetUserByTelegramIDCache(context.Background(), telegramID)
		require.NoError(t, err)
		require.Equal(t, user.TelegramID, cachedUser.TelegramID)
		require.Equal(t, user.TelegramUsername, cachedUser.TelegramUsername)
	})
}

func TestRemoveUserCache(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	user := entity.User{
		ID:               758,
		TelegramID:       123456789,
		TelegramUsername: "testuser",
	}
	err := redisModule.SaveOrUpdateUserCache(context.Background(), user)
	require.NoError(t, err)

	err = redisModule.RemoveUserCache(context.Background(), user.ID)
	require.NoError(t, err)

	t.Run("User should be removed from cache", func(t *testing.T) {
		_, err := redisModule.GetUserByTelegramIDCache(context.Background(), user.TelegramID)
		require.Error(t, err)
		require.Equal(t, entity.ErrUserNotFound, err)
	})
}
