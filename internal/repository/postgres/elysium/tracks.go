package elysium

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"github.com/lib/pq"
)

func (r *Repository) SongByUrl(ctx context.Context, url string) (entity.Song, error) {
	query := fmt.Sprintf(`
SELECT 
	s.id,
	s.url,
	s.artist_name,
	s.title,
	s.cover_link,
	s.cover,
	s.cover_telegram_file_id,
	s.song_telegram_message_id,
	s.song_telegram_message_chat_id,
	COALESCE(count(usd.id), 0) download_count,
	COALESCE(count(sp.id), 0) plays_count,
	s.tags,
	s.date_create
FROM elysium.songs s
	LEFT JOIN elysium.user_song_downloads usd ON s.id = usd.song_id
	LEFT JOIN elysium.song_plays sp ON s.id = sp.song_id
WHERE s.url = $1 
GROUP BY s.id
`)

	var song entity.Song
	err := r.db.QueryRowContext(ctx, query, url).
		Scan(
			&song.ID,
			&song.URL,
			&song.ArtistName,
			&song.Title,
			&song.CoverLink,
			&song.CoverPath,
			&song.CoverTelegramFileID,
			&song.SongTelegramMessageID,
			&song.SongTelegramMessageChatID,
			&song.DownloadCount,
			&song.PlaysCount,
			pq.Array(&song.Tags),
			&song.DateCreate,
		)
	if err != nil {
		return entity.Song{}, fmt.Errorf("scan: %w", err)
	}

	return song, nil
}

// CreateSong создает новый трек в базе данных
func (r *Repository) CreateSong(ctx context.Context, song entity.Song) (entity.Song, error) {
	query := fmt.Sprintf(`
INSERT INTO songs(url, artist_name, title, cover_link, cover, cover_telegram_file_id, song_telegram_message_id, song_telegram_message_chat_id, tags)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, date_create
`)

	err := r.db.QueryRowContext(ctx, query, song.URL, song.ArtistName, song.Title, song.CoverLink, song.CoverPath, song.CoverTelegramFileID, song.SongTelegramMessageID, song.SongTelegramMessageChatID, pq.Array(song.Tags)).
		Scan(&song.ID, &song.DateCreate)
	if err != nil {
		return entity.Song{}, fmt.Errorf("query row context: %w", err)
	}

	return song, nil
}

func (r *Repository) SetCoverTelegramFileIDForSong(ctx context.Context, songID int, fileID string) error {
	query := fmt.Sprintf(`
UPDATE songs SET cover_telegram_file_id = $1 WHERE id = $2
`)
	_, err := r.db.ExecContext(ctx, query, fileID, songID)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	return nil
}

func (r *Repository) CreatePlay(ctx context.Context, songID int) error {
	query := `
INSERT INTO song_plays(song_id)
VALUES ($1)
    `
	_, err := r.db.ExecContext(ctx, query, songID)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	return nil
}

// CreateSongAndAddToPlayed создает новый трек и логирует его проигрывание
func (r *Repository) CreateSongAndAddToPlayed(ctx context.Context, song entity.Song) (entity.Song, error) {
	err := r.execTX(ctx, func(q *queries) error {
		var err error
		song, err = q.createSong(ctx, song)
		if err != nil {
			return fmt.Errorf("create song: %w", err)
		}

		_, err = q.createSongPlay(ctx, song.ID)
		if err != nil {
			return fmt.Errorf("create song play: %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.Song{}, fmt.Errorf("exec tx: %w", err)
	}

	return song, nil
}

func (r *Repository) SongPlayed(ctx context.Context, songID int) (entity.SongPlay, error) {
	var songPlay entity.SongPlay
	err := r.execTX(ctx, func(q *queries) error {
		var err error
		songPlay, err = q.createSongPlay(ctx, songID)
		if err != nil {
			return fmt.Errorf("create song play: %w", err)
		}

		return nil
	})

	if err != nil {
		return entity.SongPlay{}, fmt.Errorf("exec tx: %w", err)
	}

	return songPlay, nil

}

func (r *Repository) RemoveSong(ctx context.Context, songID int) error {
	query := fmt.Sprintf(`
DELETE FROM songs WHERE id = $1
`)
	_, err := r.db.ExecContext(ctx, query, songID)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	return nil
}

func (r *Repository) GetPlayedCountByID(ctx context.Context, songID int) (int, error) {
	query := fmt.Sprintf(`
SELECT count(id) FROM song_plays WHERE song_id = $1
`)

	var count int
	err := r.db.QueryRowContext(ctx, query, songID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("query row context: %w", err)

	}

	return count, nil
}

func (r *Repository) GetPlayedCountByURL(ctx context.Context, url string) (int, error) {
	query := fmt.Sprintf(`
SELECT count(sp.id) FROM song_plays sp
	JOIN songs s ON sp.song_id = s.id
WHERE s.url = $1
`)

	var count int
	err := r.db.QueryRowContext(ctx, query, url).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("query row context: %w", err)

	}

	return count, nil
}

func (r *Repository) GetAllPlaysByURL(ctx context.Context, url string) ([]entity.SongPlay, error) {
	query := fmt.Sprintf(`
SELECT sp.id, sp.song_id, sp.play_time FROM song_plays sp
	JOIN songs s ON sp.song_id = s.id
WHERE s.url = $1
`)

	rows, err := r.db.QueryContext(ctx, query, url)
	if err != nil {
		return nil, fmt.Errorf("query context: %w", err)
	}

	defer rows.Close()
	var songPlays []entity.SongPlay
	for rows.Next() {
		var songPlay entity.SongPlay
		err := rows.Scan(&songPlay.ID, &songPlay.SongID, &songPlay.PlayTime)
		if err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		songPlays = append(songPlays, songPlay)
	}

	return songPlays, nil
}

func (q *queries) removeSong(ctx context.Context, songID int) error {
	query := fmt.Sprintf(`
DELETE FROM songs WHERE id = $1
`)
	_, err := q.db.ExecContext(ctx, query, songID)
	if err != nil {
		return fmt.Errorf("exec context: %w", err)
	}

	return nil
}

// createSong создает новый трек в базе данных (без логирования проигрывания)
func (q *queries) createSong(ctx context.Context, song entity.Song) (entity.Song, error) {
	query := fmt.Sprintf(`
INSERT INTO songs(url, artist_name, title, cover_link, cover, cover_telegram_file_id, song_telegram_message_id, song_telegram_message_chat_id, tags)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, date_create
`)

	err := q.db.QueryRowContext(ctx, query, song.URL, song.ArtistName, song.Title, song.CoverLink, song.CoverPath, song.CoverTelegramFileID, song.SongTelegramMessageID, song.SongTelegramMessageChatID, pq.Array(song.Tags)).
		Scan(&song.ID, &song.DateCreate)
	if err != nil {
		return entity.Song{}, fmt.Errorf("query row context: %w", err)
	}

	return song, nil
}

func (q *queries) createSongPlay(ctx context.Context, songID int) (entity.SongPlay, error) {
	query := fmt.Sprintf(`
INSERT INTO song_plays(song_id)
VALUES ($1)
RETURNING id, play_time
`)
	var songPlay entity.SongPlay
	songPlay.SongID = songID
	err := q.db.QueryRowContext(ctx, query, songID).Scan(&songPlay.ID, &songPlay.PlayTime)
	if err != nil {
		return entity.SongPlay{}, fmt.Errorf("query row context: %w", err)
	}

	return songPlay, nil
}

//func (r *Repository) SongPlayed1(info entity.TrackInfo) error {
//	err := r.execTX(ctx, func(q *queries) error {
//		inValues := "phone, telegram, username"
//		inPlaceholders := "$1, $2, $3"
//		placeholders := []interface{}{user.Phone, user.Telegram}
//
//		// Администраторы могут иметь пароль и логин для аутенфикации в АДМИНКУ
//		if !user.IsTelegramUser {
//			inValues = fmt.Sprintf("%s, %s", inValues, "login, password")
//			inPlaceholders = fmt.Sprintf("%s, %s", inPlaceholders, "$4, $5")
//			placeholders = append(placeholders, user.Login, user.Password)
//		}
//
//		query := fmt.Sprintf(`INSERT INTO users(%s) VALUES (%s) RETURNING id`, inValues, inPlaceholders)
//
//		err := q.db.QueryRowContext(ctx, query, placeholders...).Scan(&user.Id)
//		if err != nil {
//			return fmt.Errorf("new user: query row context: %w", err)
//		}
//
//		return nil
//	})
//	if err != nil {
//		return domain.User{}, fmt.Errorf("exec tx: %w", err)
//	}
//
//	return user, nil
//}
