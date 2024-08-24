package entity

// WebsocketInfo TODO переименовать
type WebsocketInfo struct {
	OnlineUsersCount int64     `json:"online_users_count"`
	CurrentTrack     TrackInfo `json:"current_track"`
}
