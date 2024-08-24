package demethra

import "arimadj-helper/internal/entity"

func (m *Module) CurrentTrack() entity.TrackInfo {
	return m.currentTrack
}
