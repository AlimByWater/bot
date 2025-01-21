package entity

import "time"

// Song represents a song in the database
// easyjson:json
type Song struct {
	ID                        int       `db:"id" json:"id"`
	URL                       string    `db:"url" json:"url"`
	ArtistName                string    `db:"artist_name" json:"artist_name"`
	Title                     string    `db:"title" json:"title"`
	CoverLink                 string    `db:"cover_link" json:"cover_link"`
	CoverTelegramFileID       string    `db:"cover_telegram_file_id" json:"cover_telegram_file_id"`
	SongTelegramMessageID     int       `db:"song_telegram_message_id" json:"song_telegram_message_id"`           // ID сообщения содержащего данный файл, для оперативного репоста
	SongTelegramMessageChatID int64     `db:"song_telegram_message_chat_id" json:"song_telegram_message_chat_id"` // ID чата, в котором находится сообщение с файлом
	DownloadCount             int       `db:"download_count" json:"download_count"`
	PlaysCount                int       `json:"plays_count"`
	Tags                      []string  `db:"tags" json:"tags"`
	DateCreate                time.Time `db:"date_create" json:"date_create"`
}

// SongPlay represents a record of a song being played
// easyjson:json
type SongPlay struct {
	ID       int       `db:"id" json:"id"`
	SongID   int       `db:"song_id" json:"song_id"`
	PlayTime time.Time `db:"play_time" json:"play_time"`
}
