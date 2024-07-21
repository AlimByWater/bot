package elysium

import (
	"arimadj-helper/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (r *Repository) CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error) {
	err := r.execTX(ctx, func(q *queries) error {
		currentUser, err := q.getUserByTelegramUserID(ctx, user.TelegramID)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("failed to get user: %w", err)
			}
		}

		if currentUser.ID != 0 {
			if currentUser.TelegramUsername != user.TelegramUsername || currentUser.Firstname != user.Firstname {
				query := `
				UPDATE elysium.users
				SET firstname = $1, telegram_username = $2
				WHERE id = $3
				`
				_, err := q.db.ExecContext(ctx, query, user.Firstname, user.TelegramUsername, currentUser.ID)
				if err != nil {
					return fmt.Errorf("failed to update user: %w", err)
				}

			}
			return nil
		}

		query := `
		INSERT INTO elysium.users 
		(telegram_id, telegram_username, firstname, date_create)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

		if user.DateCreate.IsZero() {
			user.DateCreate = time.Now()
		}

		err = q.db.QueryRowContext(ctx, query,
			user.TelegramID,
			user.TelegramUsername,
			user.Firstname,
			user.DateCreate,
		).Scan(&user.ID)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return entity.User{}, fmt.Errorf("exec tx: %w", err)
	}

	return user, nil
}

func (r *Repository) GetUserByTelegramID(ctx context.Context, telegramID int64) (entity.User, error) {
	var user entity.User
	err := r.execTX(ctx, func(q *queries) error {
		var err error
		user, err = q.getUserByTelegramUserID(ctx, telegramID)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		return nil

	})
	if err != nil {
		return entity.User{}, fmt.Errorf("exec tx: %w", err)

	}

	return user, nil
}

// func get full user by telegram user_id in queries
func (q *queries) getUserByTelegramUserID(ctx context.Context, telegramUserID int64) (entity.User, error) {
	query := `
		SELECT id, telegram_id, telegram_username, firstname, date_create
		FROM elysium.users
		WHERE telegram_id = $1
	`

	var user entity.User
	err := q.db.QueryRowContext(ctx, query, telegramUserID).Scan(
		&user.ID,
		&user.TelegramID,
		&user.TelegramUsername,
		&user.Firstname,
		&user.DateCreate,
	)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to query user: %w", err)
	}

	return user, nil
}

func (q *queries) getUserIDByTelegramUserID(ctx context.Context, telegramUserID int64) (int, error) {
	query := `
		SELECT id FROM elysium.users WHERE telegram_user_id = $1
	`

	var userID int
	err := q.db.QueryRowContext(ctx, query, telegramUserID).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to query user ID: %w", err)
	}

	return userID, nil
}

func (r *Repository) DeleteUser(ctx context.Context, userID int) error {
	query := `
		DELETE FROM elysium.users WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
