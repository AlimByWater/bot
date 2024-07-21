package elysium

import (
	"arimadj-helper/internal/entity"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

func (r *Repository) SaveWebAppEvent(ctx context.Context, event entity.WebAppEvent) error {
	err := r.execTX(ctx, func(q *queries) error {
		if event.UserID == 0 {
			userID, err := q.getUserIDByTelegramUserID(ctx, event.TelegramUserID)
			if err != nil {
				return fmt.Errorf("failed to get user ID: %w", err)
			}
			event.UserID = userID
		}

		query := `
		INSERT INTO elysium.web_app_events 
		(event_type, user_id, telegram_user_id, payload, session_id, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

		_, err := r.db.ExecContext(ctx, query,
			event.EventType,
			event.UserID,
			event.TelegramUserID,
			event.Payload,
			event.SessionID,
			event.Timestamp,
		)

		if err != nil {
			return fmt.Errorf("failed to save web app event: %w", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("exec tx: %w", err)
	}

	return nil
}

func (r *Repository) GetEventsByTelegramUserID(ctx context.Context, telegramUserID int64, since time.Time) ([]entity.WebAppEvent, error) {
	query := `
		SELECT event_type, user_id, telegram_user_id, payload, session_id, timestamp
		FROM elysium.web_app_events
		WHERE telegram_user_id = $1
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, telegramUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to query web app events: %w", err)
	}
	defer rows.Close()

	var events []entity.WebAppEvent
	for rows.Next() {
		var event entity.WebAppEvent
		var payloadJSON []byte

		err := rows.Scan(
			&event.EventType,
			&event.UserID,
			&event.TelegramUserID,
			&payloadJSON,
			&event.SessionID,
			&event.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan web app event: %w", err)
		}

		err = json.Unmarshal(payloadJSON, &event.Payload)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return events, nil
}
