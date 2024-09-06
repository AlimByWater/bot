package song_methods

import (
	"context"
)

type songDownloadUC interface {
	SendSongByTrackLink(ctx context.Context, userID int, trackLink string) error
}
