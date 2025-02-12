package elysium

import (
	"context"
	"database/sql"
	"elysium/internal/entity"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
)

// CreateOrUpdateUser создает пользователя. Если пользователь уже существует, то обновляется его данные из телеграма (TelegramID, TelegramUsername, Firstname)
func (r *Repository) CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error) {
	err := r.execTX(ctx, func(q *queries) error {
		// Основной запрос для upsert пользователя
		userQuery := `
            INSERT INTO users AS u 
                (telegram_id, telegram_username, firstname)
            VALUES ($1, $2, $3)
            ON CONFLICT (telegram_id) DO UPDATE
                SET 
                    telegram_username = COALESCE(NULLIF($2, ''), u.telegram_username),
                    firstname = COALESCE(NULLIF($3, ''), u.firstname)
            RETURNING id
        `
		// Получаем ID пользователя после вставки/обновления
		err := q.db.QueryRowContext(ctx, userQuery,
			user.TelegramID,
			user.TelegramUsername,
			user.Firstname,
		).Scan(&user.ID)
		if err != nil {
			return fmt.Errorf("failed to upsert user: %w", err)
		}

		// Добавляем права ТОЛЬКО для новых пользователей
		permQuery := `
            INSERT INTO permissions 
                (user_id, private_generation, use_by_channel_name, vip)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (user_id) DO NOTHING
        `
		_, err = q.db.ExecContext(ctx, permQuery,
			user.ID,
			user.Permissions.PrivateGeneration,
			user.Permissions.UseByChannelName,
			user.Permissions.Vip,
		)
		if err != nil {
			return fmt.Errorf("failed to insert permissions: %w", err)
		}

		return nil
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
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

		bots, err := q.getUserActiveBots(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("get user active bots: %w", err)
		}

		user.BotsActivated = bots

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
		SELECT 
			u.telegram_id, 
			u.telegram_username, 
			u.firstname, 
			u.date_create,
			p.private_generation,
			p.use_by_channel_name,
			p.vip
		FROM users u
		LEFT JOIN permissions p ON u.id = p.user_id
		WHERE u.id = $1
		`
		var telegramID sql.NullInt64
		var telegramUsername sql.NullString
		var firstname sql.NullString
		var privateGeneration sql.NullBool
		var useByChannelName sql.NullBool
		var vip sql.NullBool

		err := q.db.QueryRowContext(ctx, query, userID).Scan(
			&telegramID,
			&telegramUsername,
			&firstname,
			&user.DateCreate,
			&privateGeneration,
			&useByChannelName,
			&vip,
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

		// Handle possible NULL permissions
		user.Permissions.PrivateGeneration = privateGeneration.Valid && privateGeneration.Bool
		user.Permissions.UseByChannelName = useByChannelName.Valid && useByChannelName.Bool
		user.Permissions.Vip = vip.Valid && vip.Bool

		bots, err := q.getUserActiveBots(ctx, user.ID)
		if err != nil {
			return fmt.Errorf("get user active bots: %w", err)
		}

		user.BotsActivated = bots

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
		SELECT 
			u.id, 
			u.telegram_id, 
			u.telegram_username, 
			u.firstname, 
			u.date_create,
			p.private_generation,
			p.use_by_channel_name,
			p.vip
		FROM users u
		LEFT JOIN permissions p ON u.id = p.user_id
		WHERE u.telegram_id = $1
	`

	var telegramUsername sql.NullString
	var firstname sql.NullString

	var (
		user              entity.User
		privateGeneration sql.NullBool
		useByChannelName  sql.NullBool
		vip               sql.NullBool
	)

	err := q.db.QueryRowContext(ctx, query, telegramUserID).Scan(
		&user.ID,
		&user.TelegramID,
		&telegramUsername,
		&firstname,
		&user.DateCreate,
		&privateGeneration,
		&useByChannelName,
		&vip,
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

	// Handle possible NULL permissions
	user.Permissions.PrivateGeneration = privateGeneration.Valid && privateGeneration.Bool
	user.Permissions.UseByChannelName = useByChannelName.Valid && useByChannelName.Bool
	user.Permissions.Vip = vip.Valid && vip.Bool

	return user, nil
}

func (q *queries) getUserIDByTelegramUserID(ctx context.Context, telegramUserID int64) (int, error) {
	query := `
		SELECT id FROM users WHERE telegram_user_id = $1
	`

	var userID int
	err := q.db.QueryRowContext(ctx, query, telegramUserID).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("failed to query user ID: %w", err)
	}

	return userID, nil
}

func (r *Repository) UpdatePermissions(ctx context.Context, userID int, perms entity.Permissions) error {
	query := `
        UPDATE permissions
        SET 
            private_generation = $1,
            use_by_channel_name = $2,
            vip = $3,
            updated_at = NOW()
        WHERE user_id = $4
    `
	_, err := r.db.ExecContext(ctx, query,
		perms.PrivateGeneration,
		perms.UseByChannelName,
		perms.Vip,
		userID,
	)
	return err
}

func (r *Repository) DeleteUser(ctx context.Context, userID int) error {
	query := `
		DELETE FROM users WHERE id = $1
	`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

func (r *Repository) GetUsersByTelegramID(ctx context.Context, telegramIDs []int64) ([]entity.User, error) {
	query, args, err := sqlx.In(`                                                                                                                                                                                 
         SELECT                                                                                                                                                                                                    
             u.id,                                                                                                                                                                                                 
             u.telegram_id,                                                                                                                                                                                        
             u.telegram_username,                                                                                                                                                                                  
             u.firstname,                                                                                                                                                                                          
             u.date_create,                                                                                                                                                                                                                                                                                                                                                                                 
             p.private_generation,                                                                                                                                                                                 
             p.use_by_channel_name,                                                                                                                                                                                
             p.vip,                                                                                                                                                                                                
             COALESCE(json_agg(                                                                                                                                                                                    
                 json_build_object(                                                                                                                                                                                
                     'id', b.id,                                                                                                                                                                                   
                     'name', b.name,                                                                                                                                                                               
                     'token', b.token,                                                                                                                                                                             
                     'purpose', b.purpose,                                                                                                                                                                         
                     'test', b.test,                                                                                                                                                                               
                     'enabled', b.enabled                                                                                                                                                                          
                 )                                                                                                                                                                                                 
             ) FILTER (WHERE b.id IS NOT NULL), '[]') AS bots_activated                                                                                                                                            
         FROM users u                                                                                                                                                                                      
         LEFT JOIN permissions p ON u.id = p.user_id                                                                                                                                                               
         LEFT JOIN user_to_bots ub ON u.id = ub.user_id AND ub.active = true                                                                                                                                       
         LEFT JOIN bots b ON ub.bot_id = b.id AND b.enabled = true                                                                                                                                                 
         WHERE u.telegram_id IN (?)                                                                                                                                                                                
         GROUP BY u.id, u.telegram_id, u.telegram_username, u.firstname, u.date_create, p.private_generation, p.use_by_channel_name, p.vip                                                              
     `, telegramIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	query = r.db.Rebind(query)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []entity.User

	for rows.Next() {
		var user entity.User
		var telegramUsername sql.NullString
		var firstname sql.NullString
		var privateGeneration sql.NullBool
		var useByChannelName sql.NullBool
		var vip sql.NullBool
		var botsJSON []byte

		err := rows.Scan(
			&user.ID,
			&user.TelegramID,
			&telegramUsername,
			&firstname,
			&user.DateCreate,
			&privateGeneration,
			&useByChannelName,
			&vip,
			&botsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		if telegramUsername.Valid {
			user.TelegramUsername = telegramUsername.String
		}
		if firstname.Valid {
			user.Firstname = firstname.String
		}

		user.Permissions = entity.Permissions{
			PrivateGeneration: privateGeneration.Valid && privateGeneration.Bool,
			UseByChannelName:  useByChannelName.Valid && useByChannelName.Bool,
			Vip:               vip.Valid && vip.Bool,
		}

		// Парсим JSON с ботами
		var bots []*entity.Bot
		err = json.Unmarshal(botsJSON, &bots)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal bots: %w", err)
		}
		user.BotsActivated = bots

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}
