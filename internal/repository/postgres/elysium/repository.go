package elysium

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
	"arimadj-helper/internal/entity"
)

const defaultQueryTimeout = 30 * time.Second

type (
	dbtx interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
		PrepareContext(context.Context, string) (*sql.Stmt, error)
		QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
		QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	}

	queries struct {
		db dbtx
	}

	Repository struct {
		db *sqlx.DB
	}
)

func NewRepository() *Repository {
	return &Repository{}
}

// AddDb - добавляет в структуру пул соединений с СУБД
func (r *Repository) AddDb(db *sqlx.DB) {
	r.db = db
}

func newQueries(db dbtx) *queries {
	return &queries{db: db}
}

func (r *Repository) execTX(ctx context.Context, fn func(*queries) error) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := newQueries(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %w, rb err: %s", err, rbErr.Error())
		}

		return fmt.Errorf("tx error: %w", err)
	}

	return tx.Commit()
}

func (r *Repository) Close() error {
	return r.db.Close()
}

func (r *Repository) SaveWebAppEvent(ctx context.Context, event entity.WebAppEvent) error {
	if event.UserID == 0 {
		userID, err := r.getUserIDByTelegramUserID(ctx, event.TelegramUserID)
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
}

func (r *Repository) getUserIDByTelegramUserID(ctx context.Context, telegramUserID int64) (int, error) {
	query := `
		SELECT id FROM elysium.users WHERE telegram_user_id = $1
	`
	
	var userID int
	err := r.db.QueryRowContext(ctx, query, telegramUserID).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user not found for telegram_user_id %d", telegramUserID)
		}
		return 0, fmt.Errorf("failed to query user ID: %w", err)
	}
	
	return userID, nil
}

func (r *Repository) GetEventsByTelegramUserID(ctx context.Context, telegramUserID int64) ([]entity.WebAppEvent, error) {
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
