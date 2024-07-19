package api_methods

import "arimadj-helper/internal/entity"

type botUC interface {
	NextSong(track entity.TrackInfo)
	ProcessInitWebAppData(data entity.InitWebApp) error
}
