package entity

import "time"

type User struct {
	ID               int       `db:"id" json:"id"`
	TelegramID       int64     `db:"telegram_id" json:"telegram_id"`
	TelegramUsername string    `db:"telegram_username" json:"username"`
	Firstname        string    `db:"firstname" json:"firstname"`
	DateCreate       time.Time `db:"date_create" json:"date_create"`
}

// Song represents a song in the database
type Song struct {
	ID                        int       `db:"id" json:"id"`
	URL                       string    `db:"url" json:"url"`
	ArtistName                string    `db:"artist_name" json:"artist_name"`
	Title                     string    `db:"title" json:"title"`
	CoverLink                 string    `db:"cover_link" json:"cover_link"`
	CoverPath                 string    `db:"cover" json:"cover"`
	CoverTelegramFileID       string    `db:"cover_telegram_file_id" json:"cover_telegram_file_id"`
	SongTelegramMessageID     int       `db:"song_telegram_message_id" json:"song_telegram_message_id"`           // ID сообщения содержащего данный файл, для оперативного репоста
	SongTelegramMessageChatID int64     `db:"song_telegram_message_chat_id" json:"song_telegram_message_chat_id"` // ID чата, в котором находится сообщение с файлом
	DownloadCount             int       `db:"download_count" json:"download_count"`
	PlaysCount                int       `json:"plays_count"`
	Tags                      []string  `db:"tags" json:"tags"`
	DateCreate                time.Time `db:"date_create" json:"date_create"`
}

// UserSongDownload represents a record of a user downloading a song
type UserSongDownload struct {
	ID           int       `db:"id" json:"id"`
	UserID       int       `db:"user_id" json:"user_id"`
	SongID       int       `db:"song_id" json:"song_id"`
	DownloadDate time.Time `db:"download_date" json:"download_date"`
}

// SongPlay represents a record of a song being played
type SongPlay struct {
	ID       int       `db:"id" json:"id"`
	SongID   int       `db:"song_id" json:"song_id"`
	PlayTime time.Time `db:"play_time" json:"play_time"`
}
