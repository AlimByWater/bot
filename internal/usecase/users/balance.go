package users

import (
	"context"
	"elysium/internal/entity"
)

// GetUserBalance возвращает текущий баланс пользователя
func (m *Module) GetUserBalance(ctx context.Context, userID int) (int, error) {
	return m.repo.GetUserBalance(ctx, userID)
}

// GetUserAccount возвращает информацию об аккаунте пользователя
func (m *Module) GetUserAccount(ctx context.Context, userID int) (entity.UserAccount, error) {
	return m.repo.GetUserAccount(ctx, userID)
}
