package redis

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var (
	ErrUserIDZero = fmt.Errorf("userID or telegramID is zero")
)

const userCachePrefix = "user:"
const userCacheByTelegramIDPrefix = "user_telegram_id:"

func (m *Module) SaveOrUpdateUserCache(ctx context.Context, user entity.User) error {
	if user.TelegramID == 0 || user.ID == 0 {
		return ErrUserIDZero
	}

	data, err := user.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	key := fmt.Sprintf("%s%d", userCachePrefix, user.ID)

	err = m.client.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save user to cache: %w", err)
	}

	key = fmt.Sprintf("%s%d", userCacheByTelegramIDPrefix, user.TelegramID)
	err = m.client.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to save user to cache: %w", err)
	}

	return nil
}

func (m *Module) GetUserByTelegramIDCache(ctx context.Context, telegramID int64) (entity.User, error) {
	if telegramID == 0 {
		return entity.User{}, ErrUserIDZero
	}

	key := fmt.Sprintf("%s%d", userCacheByTelegramIDPrefix, telegramID)
	data, err := m.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return entity.User{}, entity.ErrUserNotFound
	} else if err != nil {
		return entity.User{}, fmt.Errorf("failed to get user from cache: %w", err)
	}

	var user entity.User
	err = user.UnmarshalJSON([]byte(data))
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return user, nil
}

func (m *Module) RemoveUserCache(ctx context.Context, userID int) error {
	if userID == 0 {
		return ErrUserIDZero
	}

	key := fmt.Sprintf("%s%d", userCachePrefix, userID)

	data, err := m.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to get user from cache: %w", err)
	}

	var user entity.User
	err = user.UnmarshalJSON([]byte(data))
	if err != nil {
		return fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	err = m.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to remove user by id from cache: %w", err)
	}

	key = fmt.Sprintf("%s%d", userCacheByTelegramIDPrefix, user.TelegramID)
	err = m.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to remove user by telegram id from cache: %w", err)
	}

	return nil
}
