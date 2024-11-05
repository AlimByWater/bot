package entity

import (
	"sync"
	"time"
)

const SongDownloadSourceWebApp string = "web_app"
const SongDownloadSourceBot string = "bot"

type StreamsMetaInfo struct {
	OnlineUsersCount int64              `json:"online_users_count"`
	CurrentTrack     TrackInfo          `json:"current_track"`
	Streams          map[string]*Stream `json:"streams"`
}

type Stream struct {
	Slug             string    `json:"slug"`
	CurrentTrack     TrackInfo `json:"current_track"`
	OnlineUsersCount int64     `json:"online_users_count"`
	PrevTrack        TrackInfo

	LastPlayed  SongPlay
	LastUpdated time.Time

	Mu sync.RWMutex
}
