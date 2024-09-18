package web_app_methods

import (
	"context"
	"elysium/internal/entity"
)

type botUC interface {
	NextSong(track entity.TrackInfo)
	ProcessWebAppEvent(ctx context.Context, event entity.WebAppEvent)
	ProcessWebAppState(ctx context.Context, event entity.WebAppState)
}

type usersUC interface {
	GetOnlineUsersCount() int64
}

type songTrackerUC interface {
	CurrentTrack() entity.TrackInfo
}
