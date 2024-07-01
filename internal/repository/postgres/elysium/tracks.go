package elysium

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
)

func (r *Repository) AddNewTrack() error {

	return nil
}

func (r *Repository) SongPlayed(info entity.TrackInfo) error {
	ctx := context.TODO()
	err := r.execTX(ctx, func(q *queries) error {

		return nil
	})
	if err != nil {
		return fmt.Errorf("exec tx: %w", err)
	}

	return nil
}

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
	s.release_date,
	count(usd.id) download_count,
	count(sp.id) plays_count,
	s.tags,
	s.telegram_message_link,
	s.date_create
FROM songs s
	JOIN user_song_downloads usd ON s.id = usd.song_id
	JOIN song_plays sp ON s.id = sp.song_id
WHERE s.url = $1 
GROUP BY s.id
`)

	var song entity.Song
	err := r.db.QueryRowContext(ctx, query, url).
		Scan(
			song.ID,
			song.URL,
			song.ArtistName,
			song.Title,
			song.CoverLink,
			song.CoverPath,
			song.CoverTelegramFileID,
			song.ReleaseDate,
			song.DownloadCount,
			song.PlaysCount,
			song.Tags,
			song.TelegramMessageLink,
			song.DateCreate,
		)
	if err != nil {
		return entity.Song{}, fmt.Errorf("scan: %w", err)
	}

	return song, nil
}

func (r *Repository) CreateSong(ctx context.Context, song entity.Song) (entity.Song, error) {

	return entity.Song{}, nil
}

func (q *queries) createSong() {
	q := fmt.Sprintf(`
INSERT INTO `)
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
