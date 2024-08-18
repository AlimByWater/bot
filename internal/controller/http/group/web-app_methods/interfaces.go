package web_app_methods

import (
	"arimadj-helper/internal/entity"
	"context"
)

type botUC interface {
	NextSong(track entity.TrackInfo)
	ProcessWebAppEvent(ctx context.Context, event entity.WebAppEvent)
	ProcessWebAppState(ctx context.Context, event entity.WebAppState)
}

type usersUC interface {
	WebsocketInfo() (entity.WebsocketInfo, error)
}
