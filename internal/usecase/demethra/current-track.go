package demethra

import "elysium/internal/entity"

func (m *Module) CurrentTrack() entity.TrackInfo {
	return m.currentTrack
}
