package users

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"log/slog"
)

func (m *Module) CreateUser(ctx context.Context, user entity.User) (entity.User, error) {
	createdUser, err := m.repo.CreateOrUpdateUser(ctx, user)
	if err != nil {
		m.logger.Error("Failed to create user", slog.Any("error", err), slog.Any("user", user))
		return entity.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	// Create default layout for the new user
	_, err = m.layout.GenerateAndSaveDefaultLayout(ctx, createdUser.ID, user.TelegramUsername)
	if err != nil {
		m.logger.Error("Failed to create default layout for user", slog.Any("error", err), slog.Any("userID", createdUser.ID))
	}

	return createdUser, nil
}
