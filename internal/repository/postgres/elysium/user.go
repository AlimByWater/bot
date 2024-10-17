package elysium

import (
	"context"
	"database/sql"
	"elysium/internal/entity"
	"errors"
	"fmt"
	"strings"
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

func (r *Repository) GetUserByID(ctx context.Context, userID int) (entity.User, error) {
	var user entity.User
	err := r.execTX(ctx, func(q *queries) error {
		query := `
		SELECT telegram_id, telegram_username, firstname, date_create
		FROM elysium.users
		WHERE id = $1
		`

		var telegramID sql.NullInt64
		var telegramUsername sql.NullString
		var firstname sql.NullString

		err := q.db.QueryRowContext(ctx, query, userID).Scan(
			&telegramID,
			&telegramUsername,
			&firstname,
			&user.DateCreate,
		)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		user.ID = userID
		if telegramID.Valid {
			user.TelegramID = telegramID.Int64
		}
		if telegramUsername.Valid {
			user.TelegramUsername = telegramUsername.String
		}
		if firstname.Valid {
			user.Firstname = firstname.String
		}

		return nil
	})
	if err != nil {
		return entity.User{}, fmt.Errorf("exec tx: %w", err)
	}

	user.ID = userID

	return user, nil
}

// func get full user by telegram user_id in queries
func (q *queries) getUserByTelegramUserID(ctx context.Context, telegramUserID int64) (entity.User, error) {
	query := `
		SELECT id, telegram_id, telegram_username, firstname, date_create
		FROM elysium.users
		WHERE telegram_id = $1
	`

	var telegramUsername sql.NullString
	var firstname sql.NullString

	var user entity.User
	err := q.db.QueryRowContext(ctx, query, telegramUserID).Scan(
		&user.ID,
		&user.TelegramID,
		&telegramUsername,
		&firstname,
		&user.DateCreate,
	)
	if err != nil {
		return entity.User{}, fmt.Errorf("failed to query user: %w; telegram_id: %d", err, telegramUserID)
	}

	if telegramUsername.Valid {
		user.TelegramUsername = telegramUsername.String
	}
	if firstname.Valid {
		user.Firstname = firstname.String
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

func (r *Repository) GetUsersByTelegramID(ctx context.Context, telegramIDs []int64) ([]entity.User, error) {
	query := `
		SELECT id, telegram_id, telegram_username, firstname, date_create
		FROM elysium.users
		WHERE telegram_id IN (%s)
	`

	ids := ""
	args := make([]interface{}, len(telegramIDs))
	placeholders := make([]string, len(telegramIDs))
	for i, id := range telegramIDs {
		args[i] = id
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		ids += placeholders[i] + ", "
	}
	ids = strings.TrimSuffix(ids, ", ")

	query = fmt.Sprintf(query, ids)

	//sql: converting argument $1 type: unsupported type []int64, a slice of int66
	var users []entity.User
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}

	for rows.Next() {
		var user entity.User
		err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&user.TelegramUsername,
			&user.Firstname,
			&user.DateCreate,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}
