package tampermonkey_methods

import (
	"context"
	"elysium/internal/entity"
)

type botUC interface {
	NextSong(track entity.TrackInfo)
	ProcessWebAppEvent(ctx context.Context, event entity.WebAppEvent)
	ProcessWebAppState(ctx context.Context, event entity.WebAppState)
}
