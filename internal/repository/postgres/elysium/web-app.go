package elysium

import (
	"arimadj-helper/internal/entity"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

func (r *Repository) SaveWebAppEvent(ctx context.Context, event entity.WebAppEvent) error {
	err := r.execTX(ctx, func(q *queries) error {
		query := `
INSERT INTO elysium.web_app_events
(event_type, telegram_id, payload, session_id, timestamp)
VALUES ($1, $2, $3, $4, $5)
`

		_, err := r.db.ExecContext(ctx, query,
			event.EventType,
			event.TelegramID,
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

func (r *Repository) SaveWebAppEvents(ctx context.Context, events []entity.WebAppEvent) error {
	err := r.execTX(ctx, func(q *queries) error {

		// placeholderCount
		pC := 0
		valueStrings := make([]string, 0, len(events))
		valueArgs := make([]interface{}, 0, len(events)*5)
		for _, e := range events {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", pC+1, pC+2, pC+3, pC+4, pC+5))
			valueArgs = append(valueArgs, e.EventType, e.TelegramID, e.Payload, e.SessionID, e.Timestamp)
			pC += 5
		}

		query := fmt.Sprintf(`
INSERT INTO elysium.web_app_events (event_type, telegram_id, payload, session_id, timestamp)
VALUES %s`, strings.Join(valueStrings, ","))

		_, err := r.db.ExecContext(ctx, query, valueArgs...)
		if err != nil {
			return fmt.Errorf("failed to save web app events: %w", err)
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
		SELECT event_type, telegram_id, payload, session_id, timestamp
		FROM elysium.web_app_events
		WHERE telegram_id = $1
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
			&event.TelegramID,
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
