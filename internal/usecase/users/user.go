package users

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"log/slog"
)

func (m *Module) CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error) {
	createdUser, err := m.repo.CreateOrUpdateUser(ctx, user)
	if err != nil {
		m.logger.Error("Failed to create user", slog.Any("error", err), slog.Any("user", user))
		return entity.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
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
