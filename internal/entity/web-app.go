package entity

const SongDownloadSourceWebApp string = "web_app"
const SongDownloadSourceBot string = "bot"

// WebsocketInfo TODO переименовать
type WebsocketInfo struct {
	OnlineUsersCount int64             `json:"online_users_count"`
	CurrentTrack     TrackInfo         `json:"current_track"`
	Streams          map[string]Stream `json:"streams"`
}

type Stream struct {
	Slug             string    `json:"slug"`
	CurrentTrack     TrackInfo `json:"current_track"`
	OnlineUsersCount int64     `json:"online_users_count"`
}
