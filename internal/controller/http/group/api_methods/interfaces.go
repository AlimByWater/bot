package api_methods

import "arimadj-helper/internal/entity"

type botUC interface {
	NextSong(track entity.TrackInfo)
	ProcessWebAppEvent(event entity.WebAppEvent) error
}
