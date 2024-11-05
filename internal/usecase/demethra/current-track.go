package demethra

import "elysium/internal/entity"

func (m *Module) CurrentTrackForStream(slug string) entity.TrackInfo {
	stream, ok := m.streams[slug]
	if !ok {
		return entity.TrackInfo{}
	}

	stream.Mu.RLock()
	defer stream.Mu.RUnlock()

	return stream.CurrentTrack
}
