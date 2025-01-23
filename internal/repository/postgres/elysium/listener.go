package elysium

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"strings"
)

func (r *Repository) SaveUserSessionDuration(ctx context.Context, sessionDuration entity.UserSessionDuration) error {
	err := r.execTX(ctx, func(q *queries) error {
		query := `
INSERT INTO user_session_durations
(telegram_id, start_time, end_time, duration_in_seconds)
VALUES ($1, $2, $3, $4)
`
		_, err := r.db.ExecContext(ctx, query, sessionDuration.TelegramID, sessionDuration.StartTime, sessionDuration.EndTime, sessionDuration.DurationInSeconds)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("exec tx: %w", err)
	}

	return nil
}

func (r *Repository) BatchAddSongToUserSongHistory(ctx context.Context, histories []entity.UserToSongHistory) error {
	if len(histories) == 0 {
		return nil
	}

	err := r.execTX(ctx, func(q *queries) error {
		pC := 0
		valueStrings := make([]string, 0, len(histories))
		valueArgs := make([]interface{}, 0, len(histories)*4)
		for _, h := range histories {
			valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)", pC+1, pC+2, pC+3, pC+4, pC+5))
			valueArgs = append(valueArgs, h.TelegramID, h.SongID, h.SongPlayID, h.StreamSlug, h.Timestamp)

			pC += 5
		}

		query := fmt.Sprintf(`
INSERT INTO user_to_song_history
(telegram_id, song_id, song_plays_id, stream, timestamp)
VALUES %s`, strings.Join(valueStrings, ","))

		_, err := r.db.ExecContext(ctx, query, valueArgs...)
		if err != nil {
			return fmt.Errorf("failed to batch add song to user song history: %w", err)

		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("exec tx: %w", err)
	}

	return nil
}
