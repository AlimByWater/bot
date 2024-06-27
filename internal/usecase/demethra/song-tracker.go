package demethra

import (
	"arimadj-helper/internal/entity"
	"log/slog"
)

func (m *Module) SongUpdate(track entity.TrackInfo) {
	if track.TrackLink == m.currentTrack.TrackLink {
		return
	}

	m.prevTrack = m.currentTrack
	m.currentTrack = track

	err := m.bot.updateCurrentTrackMessage(m.currentTrack, m.prevTrack)
	if err != nil {
		m.logger.Error("update current track", slog.String("err", err.Error()))
		return
	}

}
