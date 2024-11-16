package demethra

import (
	"context"
	"elysium/internal/application/logger"
	"elysium/internal/entity"
	"fmt"
	"log/slog"
	"sort"
	"time"
)

func (m *Module) initStreams() error {
	streams, err := m.repo.AvailableStreams(m.ctx)
	if err != nil {
		return fmt.Errorf("get available streams: %w", err)
	}
	for _, stream := range streams {
		m.streams[stream.Slug] = stream
		m.streamsList = append(m.streamsList, stream.Slug)
	}

	go m.availableStreamsWatcher()
	go m.streamOnlineUpdater()
	return nil
}

func (m *Module) availableStreamsWatcher() {
	for {
		streams, err := m.repo.AvailableStreams(m.ctx)
		if err != nil {
			m.logger.Error("Failed to get available streams", slog.String("error", err.Error()), slog.String("method", "availableStreamsWatcher"))
			continue
		}

		m.mu.Lock()
		for _, stream := range streams {
			if _, ok := m.streams[stream.Slug]; !ok {
				m.streams[stream.Slug] = stream
				continue
			}

			m.streams[stream.Slug].Lock()

			m.streams[stream.Slug].Name = stream.Name
			m.streams[stream.Slug].LogoLink = stream.LogoLink
			m.streams[stream.Slug].Link = stream.Link
			m.streams[stream.Slug].IconLink = stream.IconLink
			m.streams[stream.Slug].OnClickLink = stream.OnClickLink
			m.streams[stream.Slug].Priority = stream.Priority

			m.streamsList = append(m.streamsList, stream.Slug)

			m.streams[stream.Slug].Unlock()
		}

		m.mu.Unlock()
		time.Sleep(time.Minute * 1)
	}
}

func (m *Module) streamOnlineUpdater() {
	for {
		onlineUsersCount := m.users.GetOnlineUsersCount()
		for streamSlug, stream := range m.streams {
			stream.Lock()
			stream.OnlineUsersCount = onlineUsersCount[streamSlug]
			stream.Unlock()
		}

		time.Sleep(time.Second * 2)
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

	if slug == "elysium1" {
		m.updateCurrentTrackMessageForMainStream(ctx, stream, song, attributes)
	}

	if song.ID != 0 {
		stream.SetSong(song)
		err = m.songPlayed(ctx, stream, attributes, song.ID)
		if err != nil {
			m.logger.LogAttrs(ctx, slog.LevelError, "song played", logger.AppendErrorToLogs(attributes, err)...)
		}
	}

}

func (m *Module) updateCurrentTrackMessageForMainStream(ctx context.Context, stream *entity.Stream, song entity.Song, attributes []slog.Attr) {
	if stream.Slug != "elysium1" {
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
	v := make([]*entity.Stream, 0, len(m.streams))
	for _, stream := range m.streams {
		v = append(v, stream)
	}

	// Sort streams by priority, highest priority (0 first
	sort.Slice(v, func(i, j int) bool {
		return v[i].Priority > v[j].Priority
	})

	return entity.StreamsMetaInfo{
		OnlineUsersCount: m.streams["elysium1"].OnlineUsersCount,
		CurrentTrack:     m.streams["elysium1"].CurrentTrack,
		Streams:          v,
	}
}
