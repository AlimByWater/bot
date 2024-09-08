package demethra

import (
	"arimadj-helper/internal/application/logger"
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

func (m *Module) NextSong(track entity.TrackInfo) {
	track.TrackLink = strings.Split(track.TrackLink, "?")[0]
	if track.TrackLink == m.currentTrack.TrackLink {
		return
	}

	track.TrackTitle = strings.Replace(track.TrackTitle, "Current track: ", "", 1)
	track.CoverLink = strings.Replace(track.CoverLink, "t50x50", "t500x500", 1)
	track.CoverLink = strings.Replace(track.CoverLink, "t120x120", "t500x500", 1)

	ctx := context.TODO()

	// если текущий трек не пустой, то добавляем его в историю прослушивания всем текущим слушателям
	go m.addPrevSongToCurrentListenersHistory(ctx)

	attributes := []slog.Attr{
		slog.String("track_link", track.TrackLink),
		slog.String("METHOD", "next song"),
	}

	song, err := m.repo.SongByUrl(ctx, track.TrackLink)
	if err != nil { // при любой ошибки и если трек не найден

		song, err = m.downloadAndCreateNewSong(track)
		if err != nil {
			fmt.Println(track.TrackLink)
			m.logger.LogAttrs(ctx, slog.LevelWarn, "download and create new song", logger.AppendErrorToLogs(attributes, err)...)
		}
	}

	m.mu.Lock()
	m.prevTrack = m.currentTrack
	m.currentTrack = track
	m.mu.Unlock()

	msg, err := m.bot.updateCurrentTrackMessage(ctx, song.ID, m.currentTrack, m.prevTrack, song.CoverTelegramFileID, attributes)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "update current track", logger.AppendErrorToLogs(attributes, err)...)
	}

	go m.UpdateSongMetadataFile(track)

	// все еще может сложиться такая ситуация, что song не создался в базе
	// тогда не имеет смысла создавать запись о проигранном треке в базе данных
	if song.ID != 0 {
		if song.ID != 0 && msg.MessageID != 0 { // Если трек создан
			err = m.repo.SetCoverTelegramFileIDForSong(ctx, song.ID, msg.Photo[0].FileID)
			if err != nil {
				m.logger.LogAttrs(ctx, slog.LevelError, "set cover telegram file id", logger.AppendErrorToLogs(attributes, err)...)
			}
		}
		m.songPlayed(ctx, attributes, song.ID)
	}
}

func (m *Module) songPlayed(ctx context.Context, attributes []slog.Attr, songID int) {
	songPlayed, err := m.repo.SongPlayed(ctx, songID)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "song played", logger.AppendErrorToLogs(attributes, err)...)
		return
	}

	m.mu.Lock()
	m.lastPlayed = songPlayed
	m.mu.Unlock()
}

func (m *Module) addPrevSongToCurrentListenersHistory(ctx context.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.lastPlayed.ID == 0 {
		return
	}

	attributes := []slog.Attr{
		slog.Int("song_id", m.lastPlayed.SongID),
		slog.String("METHOD", "add prev song to current listeners history"),
	}

	currentlyOnStream, err := m.cache.GetAllCurrentListeners(m.ctx)
	if err != nil {
		m.logger.Error("Failed to get listeners", slog.String("error", err.Error()), slog.String("method", "addPrevSongToCurrentListenersHistory"))
		return
	}

	histories := make([]entity.UserToSongHistory, 0, len(currentlyOnStream))
	for _, listener := range currentlyOnStream {
		history := entity.UserToSongHistory{
			TelegramID: listener.TelegramID,
			SongID:     m.lastPlayed.SongID,
			SongPlayID: m.lastPlayed.ID,
			Timestamp:  m.lastPlayed.PlayTime,
		}

		histories = append(histories, history)
	}

	err = m.repo.BatchAddSongToUserSongHistory(ctx, histories)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "add song to listener history", logger.AppendErrorToLogs(attributes, err)...)
	}

}

func (m *Module) UpdateSongMetadataFile(track entity.TrackInfo) {
	//open file and create it if it doesn't exist
	file, err := os.OpenFile(m.cfg.GetSongMetadataFilePath(), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		m.logger.LogAttrs(context.TODO(), slog.LevelError, "open song metadata file", logger.AppendErrorToLogs(nil, err)...)
		return
	}

	// write or rewrite first line

	_, err = file.WriteAt([]byte(fmt.Sprintf("TrackTitle1::%s ArtistName2::%s TrackLink3::%s CoverLink4::%s", track.TrackTitle, track.ArtistName, track.TrackLink, track.CoverLink)), 0)
	if err != nil {
		m.logger.LogAttrs(context.TODO(), slog.LevelError, "write to song metadata file", logger.AppendErrorToLogs(nil, err)...)
		return
	}
}

func (m *Module) downloadAndCreateNewSong(info entity.TrackInfo) (entity.Song, error) {
	ctx := context.TODO()

	attributes := []slog.Attr{
		slog.String("track_link", info.TrackLink),
		slog.String("METHOD", "download and create new song"),
	}

	//************ КАЧАЕМ ОБЛОЖКУ  *************** //
	//coverFile, err := m.downloadCover(info.CoverLink)
	//if err != nil {
	//	m.logger.LogAttrs(ctx, slog.LevelError, "download cover", logger.AppendErrorToLogs(attributes, err)...)
	//	return entity.Song{}, fmt.Errorf("download cover: %w", err)
	//}
	//****************** ******************////
	songPath, err := m.soundcloud.DownloadTrackByURL(ctx, info.TrackLink, info)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "download track by url", logger.AppendErrorToLogs(attributes, err)...)
		return entity.Song{}, fmt.Errorf("download track by url: %w", err)
	}

	//удаляем mp3 с диска
	defer func() {
		err := os.Remove(songPath)
		if err != nil {
			m.logger.LogAttrs(ctx, slog.LevelError, "remove song", logger.AppendErrorToLogs(attributes, err)...)
		}
	}()

	file, err := os.Open(songPath)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "open song", logger.AppendErrorToLogs(attributes, err)...)
	}

	// ************* ОТПРАВИТЬ ТРЕК В ГРУППУ *************** //
	audio := tgbotapi.AudioConfig{
		BaseFile: tgbotapi.BaseFile{
			BaseChat: tgbotapi.BaseChat{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: m.cfg.GetTracksDbChannel(),
				},
			},
			File: tgbotapi.FileReader{
				Reader: file,
			},
		},
		Caption:   `[элизиум \[ラジオ\]](t.me/elysium_fm)`,
		ParseMode: "MarkdownV2",
		Title:     info.TrackTitle,
		Performer: info.ArtistName,
	}
	respMsg, err := m.bot.Api.Send(audio)
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
		return tgbotapi.Message{}, fmt.Errorf("send: %w %s", err, coverURl)
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
