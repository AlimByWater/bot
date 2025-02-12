package elysium

import (
	"context"
	"database/sql"
	"elysium/internal/entity"
	"fmt"
)

// GetUserAccount получает информацию о балансе пользователя
func (r *Repository) GetUserAccount(ctx context.Context, userID int) (entity.UserAccount, error) {
	query := `
		SELECT user_id, balance, created_at, updated_at
		FROM user_accounts
		WHERE user_id = $1
	`

	var account entity.UserAccount
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&account.UserID,
		&account.Balance,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если аккаунт не найден, создаем новый с нулевым балансом
			account = entity.UserAccount{
				UserID:  userID,
				Balance: 0,
			}
			err = r.CreateUserAccount(ctx, account)
			if err != nil {
				return entity.UserAccount{}, fmt.Errorf("failed to create user account: %w", err)
			}
			return account, nil
		}
		return entity.UserAccount{}, fmt.Errorf("failed to get user account: %w", err)
	}

	return account, nil
}

// CreateUserAccount создает новый аккаунт пользователя
func (r *Repository) CreateUserAccount(ctx context.Context, account entity.UserAccount) error {
	query := `
		INSERT INTO user_accounts (user_id, balance)
		VALUES ($1, $2)
	`

	_, err := r.db.ExecContext(ctx, query, account.UserID, account.Balance)
	if err != nil {
		return fmt.Errorf("failed to create user account: %w", err)
	}

	return nil
}

// UpdateUserBalance обновляет баланс пользователя
func (r *Repository) UpdateUserBalance(ctx context.Context, userID int, newBalance int) error {
	err := r.execTX(ctx, func(q *queries) error {
		return q.updateUserBalance(ctx, userID, newBalance)
	})
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	return nil
}

// GetUserBalance получает текущий баланс пользователя
func (r *Repository) GetUserBalance(ctx context.Context, userID int) (int, error) {
	query := `
		SELECT balance
		FROM user_accounts
		WHERE user_id = $1
	`

	var balance int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если аккаунт не найден, создаем новый с нулевым балансом
			account := entity.UserAccount{
				UserID:  userID,
				Balance: 0,
			}
			err = r.CreateUserAccount(ctx, account)
			if err != nil {
				return 0, fmt.Errorf("failed to create user account: %w", err)
			}
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance, nil
}
