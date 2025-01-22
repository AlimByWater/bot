package elysium

import (
	"context"
	"database/sql"
	"elysium/internal/entity"
	"fmt"
)

func (r *Repository) GetAllBots(ctx context.Context) ([]*entity.Bot, error) {
	var bots []*entity.Bot
	query := `
        SELECT id, name, token, purpose, test, enabled 
        FROM bots
        WHERE enabled = true
    `

	err := r.db.SelectContext(ctx, &bots, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all bots: %w", err)
	}

	return bots, nil
}

func (r *Repository) GetBotByID(ctx context.Context, botID int64) (*entity.Bot, error) {
	var bot entity.Bot
	query := `
        SELECT id, name, token, purpose, test, enabled 
        FROM bots 
        WHERE id = $1 AND enabled = true
    `

	err := r.db.GetContext(ctx, &bot, query, botID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("bot with id %d not found", botID)
		}
		return nil, fmt.Errorf("failed to get bot by id %d: %w", botID, err)
	}

	return &bot, nil
}

func (r *Repository) CreateBot(ctx context.Context, bot *entity.Bot) error {
	query := `
        INSERT INTO bots (id, name, token, purpose, test, enabled)
        VALUES ($1, $2, $3, $4, $5, $6)
        ON CONFLICT (id) DO UPDATE SET
            name = EXCLUDED.name,
            token = EXCLUDED.token,
            purpose = EXCLUDED.purpose,
            test = EXCLUDED.test,
            enabled = EXCLUDED.enabled
    `

	_, err := r.db.ExecContext(ctx, query,
		bot.ID,
		bot.Name,
		bot.Token,
		bot.Purpose,
		bot.Test,
		bot.Enabled,
	)

	if err != nil {
		return fmt.Errorf("failed to create/update bot: %w", err)
	}

	return nil
}

func (r *Repository) UpdateBot(ctx context.Context, bot *entity.Bot) error {
	query := `
        UPDATE bots 
        SET name = $2, token = $3, purpose = $4, test = $5, enabled = $6
        WHERE id = $1 AND enabled = true
    `

	result, err := r.db.ExecContext(ctx, query,
		bot.ID,
		bot.Name,
		bot.Token,
		bot.Purpose,
		bot.Test,
		bot.Enabled,
	)

	if err != nil {
		return fmt.Errorf("failed to update bot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bot with id %d not found or already disabled", bot.ID)
	}

	return nil
}

func (r *Repository) DeleteBot(ctx context.Context, botID int64) error {
	query := `
        UPDATE bots 
        SET enabled = false 
        WHERE id = $1 AND enabled = true
    `

	result, err := r.db.ExecContext(ctx, query, botID)
	if err != nil {
		return fmt.Errorf("failed to delete bot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bot with id %d not found or already disabled", botID)
	}

	return nil
}

func (r *Repository) DeleteBotHard(ctx context.Context, botID int64) error {
	query := `
        delete from bots where id = $1
    `

	_, err := r.db.ExecContext(ctx, query, botID)
	if err != nil {
		return fmt.Errorf("failed to delete bot: %w", err)
	}

	return nil
}

func (q *queries) saveUserToBotInteraction(ctx context.Context, userID int, botID int64) error {
	if botID == 0 {
		return nil
	}
	query := `
        INSERT INTO user_to_bots (user_id, bot_id)
        VALUES ($1, $2)
        ON CONFLICT (user_id, bot_id) DO NOTHING
    `
	_, err := q.db.ExecContext(ctx, query, userID, botID)
	if err != nil {
		return fmt.Errorf("failed to save user to bot interaction: %w", err)
	}

	return nil
}

// SetUserToBotActive removes user from bot interaction
func (r *Repository) SetUserToBotActive(ctx context.Context, userID int, botID int64) error {
	query := `
        INSERT INTO user_to_bots (user_id, bot_id, active)
        VALUES ($1, $2, true)
        ON CONFLICT (user_id, bot_id) DO UPDATE 
        SET active = true
    `

	_, err := r.db.ExecContext(ctx, query, userID, botID)
	if err != nil {
		return fmt.Errorf("failed to set user bot active: %w", err)
	}

	return nil
}

func (r *Repository) UnsetUserToBot(ctx context.Context, userID int, botID int64) error {
	if botID == 0 {
		return nil
	}
	query := `
        UPDATE user_to_bots
        SET active = false
        WHERE user_id = $1 AND bot_id = $2
    `
	_, err := r.db.ExecContext(ctx, query, userID, botID)
	if err != nil {
		return fmt.Errorf("failed to unset user to bot interaction: %w", err)
	}

	return nil
}

func (r *Repository) GetUserActiveBots(ctx context.Context, userID int) ([]*entity.Bot, error) {
	var bots []*entity.Bot
	err := r.execTX(ctx, func(q *queries) error {
		var err error
		bots, err = q.getUserActiveBots(ctx, userID)
		if err != nil {
			return fmt.Errorf("get user active bots: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("exec tx: %w", err)
	}

	return bots, nil
}

func (q *queries) getUserActiveBots(ctx context.Context, userID int) ([]*entity.Bot, error) {
	var bots []*entity.Bot
	query := `
        SELECT id, name, token, purpose, test, enabled 
        FROM bots 
        WHERE enabled = true
        AND EXISTS(SELECT 1 FROM user_to_bots WHERE user_id = $1 AND bot_id = bots.id)
    `

	rows, err := q.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query context: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var bot entity.Bot
		err := rows.Scan(
			&bot.ID,
			&bot.Name,
			&bot.Token,
			&bot.Purpose,
			&bot.Test,
			&bot.Enabled,
		)
		if err != nil {
			return nil, fmt.Errorf("scan bot: %w", err)
		}
		bots = append(bots, &bot)
	}

	return bots, nil
}
