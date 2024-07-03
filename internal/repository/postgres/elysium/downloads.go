package elysium

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
)

func (r *Repository) GetDownloads(ctx context.Context, limit, offset int) ([]entity.UserSongDownload, error) {
	query := `SELECT * FROM elysium.user_song_downloads ORDER BY download_date DESC LIMIT $1 OFFSET $2`
	downloads := make([]entity.UserSongDownload, 0)
	err := r.db.SelectContext(ctx, &downloads, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select downloads fail: %w", err)
	}

	return downloads, nil
}

func (r *Repository) GetDownloadsByUserID(ctx context.Context, userID int, limit, offset int) ([]entity.UserSongDownload, error) {
	query := `SELECT * FROM elysium.user_song_downloads WHERE user_id = $1 ORDER BY download_date DESC LIMIT $2 OFFSET $3`
	downloads := make([]entity.UserSongDownload, 0)
	err := r.db.SelectContext(ctx, &downloads, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select downloads by user id fail: %w", err)
	}

	return downloads, nil
}

func (r *Repository) GetDownloadsBySongID(ctx context.Context, songID int, limit, offset int) ([]entity.UserSongDownload, error) {
	query := `SELECT * FROM elysium.user_song_downloads WHERE song_id = $1 ORDER BY download_date DESC LIMIT $2 OFFSET $3`
	downloads := make([]entity.UserSongDownload, 0)
	err := r.db.SelectContext(ctx, &downloads, query, songID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("select downloads by song id fail: %w", err)
	}

	return downloads, nil
}

func (r *Repository) SetDownload(ctx context.Context, songID, userID int) error {
	query := `INSERT INTO elysium.user_song_downloads (user_id, song_id, download_date) VALUES ($1, $2, $3)`
	_, err := r.db.ExecContext(ctx, query, userID, songID)
	if err != nil {
		return fmt.Errorf("insert download fail: %w", err)
	}

	return nil
}

func (q *queries) setDownload(ctx context.Context, songID, userID int) error {
	query := `INSERT INTO elysium.user_song_downloads (user_id, song_id, download_date) VALUES ($1, $2, $3)`
	_, err := q.db.ExecContext(ctx, query, userID, songID)
	if err != nil {
		return fmt.Errorf("insert download fail: %w", err)
	}

	return nil
}
