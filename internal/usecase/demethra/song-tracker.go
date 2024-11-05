package demethra

import (
	"context"
	"elysium/internal/application/logger"
	"elysium/internal/entity"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

func (m *Module) NextSong(stream string, track entity.TrackInfo) {
	track.SanitizeInfo()

	_, err := url.ParseRequestURI(track.TrackLink)
	if err != nil {
		m.logger.LogAttrs(context.TODO(), slog.LevelError, "parse track link", logger.AppendErrorToLogs(nil, err)...)
		return
	}
	m.UpdateStreamTrack(stream, track)
}

func (m *Module) songPlayed(ctx context.Context, stream *entity.Stream, attributes []slog.Attr, songID int) error {
	songPlayed, err := m.repo.SongPlayed(ctx, stream.Slug, songID)
	if err != nil {
		return fmt.Errorf("song played: %w", err)
	}

	stream.LastPlayed = songPlayed
	return nil
}

func (m *Module) addPrevSongToCurrentListenersHistory(ctx context.Context, stream *entity.Stream) {
	if stream.LastPlayed.ID == 0 {
		return
	}

	attributes := []slog.Attr{
		slog.Int("song_id", stream.LastPlayed.SongID),
		slog.String("METHOD", "add prev song to current listeners history"),
		slog.String("stream", stream.Slug),
	}

	currentlyOnStream, err := m.cache.GetAllCurrentListeners(m.ctx)
	if err != nil {
		m.logger.Error("Failed to get listeners", slog.String("error", err.Error()), slog.String("method", "addPrevSongToCurrentListenersHistory"))
		return
	}

	histories := make([]entity.UserToSongHistory, 0, len(currentlyOnStream))
	for _, listener := range currentlyOnStream {
		if listener.Payload.StreamSlug != stream.Slug {
			continue
		}
		history := entity.UserToSongHistory{
			TelegramID: listener.TelegramID,
			SongID:     stream.LastPlayed.SongID,
			SongPlayID: stream.LastPlayed.ID,
			Timestamp:  stream.LastPlayed.PlayTime,
			StreamSlug: stream.Slug,
		}

		histories = append(histories, history)
	}

	err = m.repo.BatchAddSongToUserSongHistory(ctx, histories)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "add song to listener history", logger.AppendErrorToLogs(attributes, err)...)
	}

}

func (m *Module) downloadAndCreateNewSong(info entity.TrackInfo) (entity.Song, error) {
	ctx := context.TODO()

	attributes := []slog.Attr{
		slog.String("track_link", info.TrackLink),
		slog.String("METHOD", "download and create new song"),
	}

	fileName, songData, err := m.downloader.DownloadByLink(ctx, info.TrackLink, "mp3")
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "download track by url", logger.AppendErrorToLogs(attributes, err)...)
		return entity.Song{}, fmt.Errorf("download track by url: %w", err)
	}

	//удаляем mp3 с диска
	defer func(fileName string) {
		err := m.downloader.RemoveFile(ctx, fileName)
		if err != nil {
			m.logger.LogAttrs(ctx, slog.LevelError, "remove song", logger.AppendErrorToLogs(attributes, err)...)
		}
	}(fileName)

	// ************* ОТПРАВИТЬ ТРЕК В ГРУППУ *************** //
	audio := tgbotapi.AudioConfig{
		BaseFile: tgbotapi.BaseFile{
			BaseChat: tgbotapi.BaseChat{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: m.cfg.GetTracksDbChannel(),
				},
			},
			File: tgbotapi.FileBytes{
				Bytes: songData,
			},
		},
		Caption:   `[элизиум \[ラジオ\]](t.me/elysium_fm)`,
		ParseMode: "MarkdownV2",
		Title:     info.TrackTitle,
		Performer: info.ArtistName,
	}
	respMsg, err := m.Bot.Api.Send(audio)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "send audio", logger.AppendErrorToLogs(attributes, err)...)
		return entity.Song{}, fmt.Errorf("send audio: %w", err)
	}

	// ************* ****************** //

	// ************* СОЗДАЕМ НОВЫЙ ТРЕК *************** //

	song := entity.Song{
		URL:                       info.TrackLink,
		ArtistName:                info.ArtistName,
		Title:                     info.TrackTitle,
		CoverLink:                 info.CoverLink,
		SongTelegramMessageID:     respMsg.MessageID,
		SongTelegramMessageChatID: respMsg.Chat.ID,
		Tags:                      info.Tags,
	}

	song, err = m.repo.CreateSong(ctx, song)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "create song", logger.AppendErrorToLogs(attributes, err)...)
		return entity.Song{}, fmt.Errorf("create song: %w", err)
	}

	return song, nil
}

// downloadCover depricated
func (b *Bot) downloadCover(link string) ([]byte, error) {
	res, err := http.Get(link)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer res.Body.Close()

	img, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("read img body: %w", err)
	}

	return img, nil
}

// updateCurrentTrackMessage обновляет сообщение с текущим треком в чате и возвращает обновленное сообщение
func (b *Bot) updateCurrentTrackMessage(ctx context.Context, songID int, current, prev entity.TrackInfo, coverFileID string, attrs []slog.Attr) (tgbotapi.Message, error) {
	coverURl := current.CoverLink
	currentFmt := current.Format()
	prevFmt := prev.Format()
	visual := formatEscapeChars(fmt.Sprintf(`0:35 ━❍──────── %s
             *↻     ⊲  Ⅱ  ⊳     ↺*
VOLUME: ▁▂▃▄▅▆▇ 100%%`, current.Duration))

	var cover tgbotapi.RequestFileData
	if coverFileID != "" {
		cover = tgbotapi.FileID(coverFileID)
	} else {
		img, err := b.downloadCover(coverURl)
		if err != nil {
			b.logger.LogAttrs(ctx, slog.LevelError, "download cover", logger.AppendErrorToLogs(attrs, err)...)
		}
		cover = tgbotapi.FileBytes{
			Name:  "cover",
			Bytes: img,
		}
	}

	data := fmt.Sprintf("download?%d", songID)

	btn := tgbotapi.NewInlineKeyboardButtonData("Добавить в плеер", data)
	row := tgbotapi.NewInlineKeyboardRow(btn)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)

	baseInputMedia := tgbotapi.BaseInputMedia{
		Type:      "photo", // Set the desired media type
		Media:     cover,
		ParseMode: "MarkdownV2", // Set the desired parse mode
		Caption: fmt.Sprintf(`
*[%s \- %s](%s)*
%s

||Предыдущий: [%s \- %s](%s)||
`,
			currentFmt.ArtistName, currentFmt.TrackTitle, currentFmt.TrackLink,
			visual,
			prevFmt.ArtistName, prevFmt.TrackTitle, prevFmt.TrackLink),
	}

	msg := tgbotapi.EditMessageMediaConfig{
		BaseEdit: tgbotapi.BaseEdit{
			BaseChatMessage: tgbotapi.BaseChatMessage{
				MessageID:  b.currentTrackMessageID,
				ChatConfig: tgbotapi.ChatConfig{ChatID: b.forumID},
			},
			ReplyMarkup: &keyboard,
		},
		Media: tgbotapi.InputMediaPhoto{
			BaseInputMedia: baseInputMedia,
		},
	}

	responseMsg, err := b.Api.Send(msg)
	if err != nil {
		return tgbotapi.Message{}, fmt.Errorf("send: %w", err)
	}

	return responseMsg, nil
}

// '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'
func formatEscapeChars(oldS string) string {
	s := oldS
	s = strings.ReplaceAll(s, `_`, `\_`)
	//s = strings.ReplaceAll(s, `*`, `\*`)
	s = strings.ReplaceAll(s, `[`, `\[`)
	s = strings.ReplaceAll(s, `]`, `\]`)
	s = strings.ReplaceAll(s, `(`, `\(`)
	s = strings.ReplaceAll(s, `)`, `\)`)
	s = strings.ReplaceAll(s, `~`, `\~`)
	//s = strings.ReplaceAll(s, "`", "\`")
	s = strings.ReplaceAll(s, `>`, `\>`)
	s = strings.ReplaceAll(s, `#`, `\#`)
	s = strings.ReplaceAll(s, `+`, `\+`)
	s = strings.ReplaceAll(s, `-`, `\-`)
	s = strings.ReplaceAll(s, `=`, `\=`)
	s = strings.ReplaceAll(s, `|`, `\|`)
	s = strings.ReplaceAll(s, `{`, `\{`)
	s = strings.ReplaceAll(s, `}`, `\}`)
	s = strings.ReplaceAll(s, `.`, `\.`)
	s = strings.ReplaceAll(s, `!`, `\!`)

	return s
}
