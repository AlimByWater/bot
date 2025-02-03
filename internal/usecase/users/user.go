package users

import (
	"context"
	"elysium/internal/entity"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
)

func (m *Module) CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error) {
	cachedUser, err := m.cache.GetUserByTelegramIDCache(ctx, user.TelegramID)
	if err != nil && !errors.Is(err, entity.ErrUserNotFound) {
		m.logger.Error("Failed to get user from cache", slog.Any("error", err))
		// Continue without returning error since this shouldn't interrupt execution
	}

	// Check if user data has changed
	if err == nil && usersEqual(cachedUser, user) {
		// Data hasn't changed, return cached user
		return cachedUser, nil
	}

	// If data changed or user isn't in cache, update in repository
	updatedUser, err := m.repo.CreateOrUpdateUser(ctx, user)
	if err != nil {
		m.logger.Error("Failed to create user", slog.Any("error", err), slog.Any("user", user))
		return entity.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	// Update user in cache
	err = m.cache.SaveOrUpdateUserCache(ctx, updatedUser)
	if err != nil {
		m.logger.Error("Failed to save user to cache", slog.Any("error", err))
		// Don't interrupt execution due to cache error
	}

	return updatedUser, nil
}

func usersEqual(u1, u2 entity.User) bool {
	return u1.TelegramUsername == u2.TelegramUsername &&
		u1.Firstname == u2.Firstname &&
		reflect.DeepEqual(u1.Permissions, u2.Permissions) &&
		u1.Balance == u2.Balance
}

func (m *Module) UserByTelegramID(ctx context.Context, telegramID int64) (entity.User, error) {
	user, err := m.repo.GetUserByTelegramID(ctx, telegramID)
	if err != nil {
		m.logger.Error("Failed to get user by Telegram ID",
			slog.Int64("telegram_id", telegramID),
			slog.String("error", err.Error()))
		return entity.User{}, fmt.Errorf("failed to get user by Telegram ID: %w", err)
	}
	return user, nil
}

func (m *Module) SetUserToBotActive(ctx context.Context, userID int, botID int64) error {
	err := m.repo.SetUserToBotActive(ctx, userID, botID)
	if err != nil {
		m.logger.Error("Failed to set user to bot active",
			slog.Int("user_id", userID),
			slog.Int64("bot_id", botID),
			slog.String("error", err.Error()))
		return fmt.Errorf("failed to set user to bot active: %w", err)
	}
	return nil
}
