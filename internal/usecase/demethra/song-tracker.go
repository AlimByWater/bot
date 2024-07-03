package demethra

import (
	"arimadj-helper/internal/application/logger"
	"arimadj-helper/internal/entity"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (m *Module) NextSong(track entity.TrackInfo) {
	track.TrackLink = strings.Split(track.TrackLink, "?")[0]
	if track.TrackLink == m.currentTrack.TrackLink {
		return
	}

	track.TrackTitle = strings.Replace(track.TrackTitle, "Current track: ", "", 1)
	track.CoverLink = strings.Replace(track.CoverLink, "t120x120", "t500x500", 1)

	ctx := context.TODO()

	j, _ := json.MarshalIndent(track, "", "  ")
	fmt.Println(string(j))

	attributes := []slog.Attr{
		slog.String("track_link", track.TrackLink),
		slog.String("method", "next song"),
	}

	songChan := make(chan entity.Song)
	defer close(songChan)

	song, err := m.repo.SongByUrl(ctx, track.TrackLink)
	if err != nil { // при любой ошибки и если трек не найден
		go func(sCh chan entity.Song) {
			song, err = m.downloadAndCreateNewSong(track)
			if err != nil {
				m.logger.LogAttrs(ctx, slog.LevelWarn, "download and create new song", logger.AppendErrorToLogs(attributes, err)...)
			}

			sCh <- song
		}(songChan)

		if !errors.Is(err, sql.ErrNoRows) {
			m.logger.LogAttrs(ctx, slog.LevelWarn, "get song by url", logger.AppendErrorToLogs(attributes, err)...)
		}
	}

	m.prevTrack = m.currentTrack
	m.currentTrack = track

	msg, err := m.bot.updateCurrentTrackMessage(m.currentTrack, m.prevTrack, song.CoverTelegramFileID)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "update current track", logger.AppendErrorToLogs(attributes, err)...)
	}

	if song.ID == 0 { // Если трек найден
		createdSong := <-songChan
		if createdSong.ID != 0 && msg.MessageID != 0 { // Если трек создан
			err = m.repo.SetCoverTelegramFileIDForSong(ctx, createdSong.ID, msg.Photo[0].FileID)
			if err != nil {
				m.logger.LogAttrs(ctx, slog.LevelError, "set cover telegram file id", logger.AppendErrorToLogs(attributes, err)...)
			}
		}
	}
}

func (m *Module) downloadAndCreateNewSong(info entity.TrackInfo) (entity.Song, error) {
	ctx := context.TODO()

	attributes := []slog.Attr{
		slog.String("track_link", info.TrackLink),
		slog.String("method", "download and create new song"),
	}

	//************ КАЧАЕМ ОБЛОЖКУ  *************** //
	coverFile, err := m.downloadCover(info.CoverLink)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "download cover", logger.AppendErrorToLogs(attributes, err)...)
		return entity.Song{}, fmt.Errorf("download cover: %w", err)
	}
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
					ChatID: TracksDbChannel,
				},
			},
			File: tgbotapi.FileReader{
				Reader: file,
			},
		},
		Caption:   `||[elysium fm](t.me/elysium_fm)||`,
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
		CoverPath:                 coverFile,
		SongTelegramMessageID:     respMsg.MessageID,
		SongTelegramMessageChatID: respMsg.Chat.ID,
		Tags:                      info.Tags,
	}

	song, err = m.repo.CreateSongAndAddToPlayed(ctx, song)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "create song", logger.AppendErrorToLogs(attributes, err)...)
		return entity.Song{}, fmt.Errorf("create song: %w", err)
	}

	return song, nil
}

func (m *Module) downloadCover(link string) (string, error) {
	res, err := http.Get(link)
	if err != nil {
		return "", fmt.Errorf("http get: %w", err)
	}
	defer res.Body.Close()

	img, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("read img body: %w", err)
	}

	//// Сохранение обложки на файловую систему
	fileName := filepath.Join("./app/images", time.Now().Format("20060102_150405")+".jpg")
	err = os.WriteFile(fileName, img, 0644)
	if err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return fileName, nil
}

// updateCurrentTrackMessage обновляет сообщение с текущим треком в чате и возвращает обновленное сообщение
func (b *Bot) updateCurrentTrackMessage(current, prev entity.TrackInfo, coverFileID string) (tgbotapi.Message, error) {
	coverURl := current.CoverLink
	trackURL := current.TrackLink
	currentFmt := current.Format()
	prevFmt := prev.Format()
	visual := formatEscapeChars(fmt.Sprintf(`0:35 ━❍──────── -%s
             *↻     ⊲  Ⅱ  ⊳     ↺*
VOLUME: ▁▂▃▄▅▆▇ 100%%`, currentFmt.Duration))

	//b.logger.Debug("song update", slog.Any("current", current), slog.Any("prev", prev))

	var cover tgbotapi.RequestFileData
	if coverFileID != "" {
		cover = tgbotapi.FileID(coverFileID)
	} else {
		cover = tgbotapi.FileURL(coverURl)
	}

	urlPart := strings.Split(trackURL, "https://soundcloud.com/")
	data := "download?"
	if len(urlPart) >= 1 {
		data = data + urlPart[1]
	}

	data = "https://t.me"

	btn := tgbotapi.NewInlineKeyboardButtonURL("Добавить в плеер", data)
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
				MessageID:  CurrentTrackMessageID,
				ChatConfig: tgbotapi.ChatConfig{ChatID: ElysiumFmID},
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
