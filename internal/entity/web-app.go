package entity

import (
	"sync"
	"time"
)

const SongDownloadSourceWebApp string = "web_app"
const SongDownloadSourceBot string = "bot"

type StreamsMetaInfo struct {
	OnlineUsersCount int64     `json:"online_users_count"`
	CurrentTrack     TrackInfo `json:"current_track"`
	Streams          []*Stream `json:"streams"`
}

type Stream struct {
	Slug             string    `json:"slug"`
	CurrentTrack     TrackInfo `json:"current_track"`
	OnlineUsersCount int64     `json:"online_users_count"`
	Link             string    `json:"link"`
	LogoLink         string    `json:"logo_link"`
	prevTrack        TrackInfo

	song Song

	lastPlayed  SongPlay
	lastUpdated time.Time

	mu sync.RWMutex
}

func (s *Stream) GetPrevTrack() TrackInfo {
	return s.prevTrack
}

func (s *Stream) SetPrevTrack(prevTrack TrackInfo) {
	s.prevTrack = prevTrack
}

func (s *Stream) GetSong() Song {
	return s.song
}

func (s *Stream) SetSong(song Song) {
	s.song = song
}

func (s *Stream) GetLastPlayed() SongPlay {
	return s.lastPlayed
}

func (s *Stream) SetLastPlayed(lastPlayed SongPlay) {
	s.lastPlayed = lastPlayed
}

func (s *Stream) GetLastUpdated() time.Time {
	return s.lastUpdated
}

func (s *Stream) SetLastUpdated(lastUpdated time.Time) {
	s.lastUpdated = lastUpdated
}

func (s *Stream) Lock() {
	s.mu.Lock()
}

func (s *Stream) Unlock() {
	s.mu.Unlock()
}

func (s *Stream) RLock() {
	s.mu.RLock()
}

func (s *Stream) RUnlock() {
	s.mu.RUnlock()
}
