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
	_, err = m.layout.GenerateAndSaveDefaultLayout(ctx, createdUser.ID)
	if err != nil {
		m.logger.Error("Failed to create default layout for user", slog.Any("error", err), slog.Any("userID", createdUser.ID))
		// We don't return an error here because the user was successfully created
		// and the layout creation is a secondary operation
	}

	return createdUser, nil
}
