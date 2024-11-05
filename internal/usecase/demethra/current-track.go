package demethra

import "elysium/internal/entity"

func (m *Module) CurrentTrackForStream(slug string) entity.TrackInfo {
	stream, ok := m.streams[slug]
	if !ok {
		return entity.TrackInfo{}
	}

	stream.mu.RLock()
	defer stream.mu.RUnlock()

	return stream.CurrentTrack
}
