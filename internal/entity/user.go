package entity

import "time"

// easyjson:json
type User struct {
	ID               int         `db:"id" json:"id"`
	TelegramID       int64       `db:"telegram_id" json:"telegram_id"`
	TelegramUsername string      `db:"telegram_username" json:"username"`
	Firstname        string      `db:"firstname" json:"firstname"`
	Permissions      Permissions `db:"permissions" json:"permissions"`
	BotsActivated    []*Bot      `db:"bots_activated" json:"bots_activated"`
	Balance          int         `db:"balance" json:"balance"`
	DateCreate       time.Time   `db:"date_create" json:"date_create"`
}

// easyjson:json
type Permissions struct {
	UserID            int       `json:"user_id" db:"user_id"`
	PrivateGeneration bool      `json:"private_generation" db:"private_generation"`
	UseByChannelName  bool      `json:"use_by_channel_name" db:"use_by_channel_name"`
	Vip               bool      `json:"vip" db:"vip"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// UserSongDownload represents a record of a user downloading a song
// easyjson:json
type UserSongDownload struct {
	ID           int       `db:"id" json:"id"`
	UserID       int       `db:"user_id" json:"user_id"`
	SongID       int       `db:"song_id" json:"song_id"`
	DownloadDate time.Time `db:"download_date" json:"download_date"`
}
