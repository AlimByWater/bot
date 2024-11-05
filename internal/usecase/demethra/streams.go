package demethra

import (
	"context"
	"elysium/internal/application/logger"
	"elysium/internal/entity"
	"fmt"
	"log/slog"
	"time"
)

func (m *Module) StreamOnlineUpdater() {
	for {
		onlineUsersCount := m.users.GetOnlineUsersCount()
		for streamSlug, stream := range m.streams {
			stream.Lock()
			stream.OnlineUsersCount = onlineUsersCount[streamSlug]
			stream.Unlock()
		}

		time.Sleep(time.Second * 10)
	}
}

func (m *Module) UpdateStreamTrack(slug string, track entity.TrackInfo) {
	stream, ok := m.streams[slug]
	if !ok {
		m.logger.Error("stream not found", slog.String("slug", slug))
		return
	}

	stream.Lock()
	defer stream.Unlock()
	defer func() {
		stream.SetLastUpdated(time.Now())
	}()

	if track.TrackLink == stream.CurrentTrack.TrackLink {
		return
	}

	stream.SetPrevTrack(stream.CurrentTrack)
	stream.CurrentTrack = track

	ctx := context.Background()
	go m.addPrevSongToCurrentListenersHistory(ctx, stream)

	attributes := []slog.Attr{
		slog.String("stream", slug),
		slog.String("track_link", track.TrackLink),
		slog.String("METHOD", "update stream track"),
	}

	song, err := m.repo.SongByUrl(ctx, track.TrackLink)
	if err != nil { // при любой ошибки и если трек не найден
		song, err = m.downloadAndCreateNewSong(track)
		if err != nil {
			fmt.Println(track.TrackLink)
			m.logger.LogAttrs(ctx, slog.LevelWarn, "download and create new song", logger.AppendErrorToLogs(attributes, err)...)
		}
	}

	if slug == "main" {
		m.updateCurrentTrackMessageForMainStream(ctx, stream, song, attributes)
	}

	if song.ID != 0 {
		err = m.songPlayed(ctx, stream, attributes, song.ID)
		if err != nil {
			m.logger.LogAttrs(ctx, slog.LevelError, "song played", logger.AppendErrorToLogs(attributes, err)...)
		}
	}

}

func (m *Module) updateCurrentTrackMessageForMainStream(ctx context.Context, stream *entity.Stream, song entity.Song, attributes []slog.Attr) {
	if stream.Slug != "main" {
		return
	}
	msg, err := m.Bot.updateCurrentTrackMessage(ctx, song.ID, stream.CurrentTrack, stream.GetPrevTrack(), song.CoverTelegramFileID, attributes)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "update current track", logger.AppendErrorToLogs(attributes, err)...)
	}

	// все еще может сложиться такая ситуация, что song не создался в базе
	// тогда не имеет смысла создавать запись о проигранном треке в базе данных
	if song.ID != 0 {
		if song.ID != 0 && msg.MessageID != 0 { // Если трек создан
			err = m.repo.SetCoverTelegramFileIDForSong(ctx, song.ID, msg.Photo[0].FileID)
			if err != nil {
				m.logger.LogAttrs(ctx, slog.LevelError, "set cover telegram file id", logger.AppendErrorToLogs(attributes, err)...)
			}
		}
	}
}

func (m *Module) GetStreamsMetaInfo() entity.StreamsMetaInfo {
	return entity.StreamsMetaInfo{
		OnlineUsersCount: m.streams["main"].OnlineUsersCount,
		CurrentTrack:     m.streams["main"].CurrentTrack,
		Streams:          m.streams,
	}
}
