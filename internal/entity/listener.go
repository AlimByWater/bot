package entity

import "time"

// easyjson:json
type UserSessionDuration struct {
	ID                int64     `json:"id"`
	TelegramID        int64     `json:"telegram_id"`
	StartTime         time.Time `json:"start_time"`
	EndTime           time.Time `json:"end_time"`
	DurationInSeconds int64     `json:"duration_in_seconds"`
}

// easyjson:json
type ListenerCache struct {
	TelegramID int64
	Payload    ListenerCachePayload
}

// easyjson:json
type ListenerCachePayload struct {
	InitTimestamp int64  `redis:"init_timestamp"`
	LastActivity  int64  `redis:"last_activity"`
	StreamSlug    string `redis:"stream_slug"`
}

// easyjson:json
type UserToSongHistory struct {
	TelegramID int64     `json:"telegram_id"`
	SongID     int       `json:"song_id"`
	SongPlayID int       `json:"song_play_id"`
	Timestamp  time.Time `json:"timestamp"`
	StreamSlug string    `json:"stream_slug"`
}
