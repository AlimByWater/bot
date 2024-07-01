package demethra

import (
	"arimadj-helper/internal/application/logger"
	"arimadj-helper/internal/entity"
	"context"
	"database/sql"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (m *Module) NextSong(track entity.TrackInfo) {
	if track.TrackLink == m.currentTrack.TrackLink {
		return
	}

	ctx := context.TODO()

	attributes := []slog.Attr{
		slog.String("track_link", track.TrackLink),
		slog.String("method", "next song"),
	}

	var song entity.Song
	var err error

	song, err = m.repo.SongByUrl(ctx, track.TrackLink)
	if err != nil && err != sql.ErrNoRows {
		m.logger.LogAttrs(ctx, slog.LevelError, "get song by url", logger.AppendErrorToLogs(attributes, err)...)
	}
	if err == sql.ErrNoRows {
		coverPath, err := m.downloadCover(track.CoverLink)
		if err != nil {
			m.logger.LogAttrs(ctx, slog.LevelError, "download cover", logger.AppendErrorToLogs(attributes, err)...)
		}

		go func() {
			song, err = m.downloadAndCreateNewSong(track)
			if err != nil {
				m.logger.LogAttrs(ctx, slog.LevelError, "download and create new song", logger.AppendErrorToLogs(attributes, err)...)
			}
		}()

	}

	m.prevTrack = m.currentTrack
	m.currentTrack = track

	msg, err := m.bot.updateCurrentTrackMessage(m.currentTrack, m.prevTrack, song.CoverTelegramFileID)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "update current track", logger.AppendErrorToLogs(attributes, err)...)
	}

	// ************* TODO обновить CoverTelegramFileID в бд ****************** //
	if song.ID != 0 {
		err = m.repo.SetCoverTelegramFileID(ctx, song.ID, msg.Photo[0].FileID)
		if err != nil {
			m.logger.LogAttrs(ctx, slog.LevelError, "set cover telegram file id", logger.AppendErrorToLogs(attributes, err)...)
		}
	}
}

func (m *Module) downloadAndCreateNewSong(info entity.TrackInfo) (entity.Song, error) {
	ctx := context.TODO()

	attributes := []slog.Attr{
		slog.String("track_link", info.TrackLink),
		slog.String("method", "download and create new song"),
	}

	// ************* КАЧАЕМ ОБЛОЖКУ  *************** //
	coverFile, err := m.downloadCover(info.CoverLink)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "download cover", logger.AppendErrorToLogs(attributes, err)...)
		return entity.Song{}, fmt.Errorf("download cover: %w", err)
	}
	// ****************** ******************////

	songPath, err := m.soundcloud.DownloadTrackByURL(ctx, info.TrackLink, info)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "download track by url", logger.AppendErrorToLogs(attributes, err)...)
		return entity.Song{}, fmt.Errorf("download track by url: %w", err)
	}

	// удаляем mp3 с диска
	defer func() {
		err := os.Remove(songPath)
		if err != nil {
			m.logger.LogAttrs(ctx, slog.LevelError, "remove song", logger.AppendErrorToLogs(attributes, err)...)
		}
	}()

	// ************* ОТПРАВИТЬ ТРЕК В ГРУППУ *************** //

	audio := tgbotapi.AudioConfig{
		BaseFile: tgbotapi.BaseFile{
			BaseChat: tgbotapi.BaseChat{},
			File:     tgbotapi.FilePath(songPath),
		},
		Caption:   `||[elysium fm](t.me/elysium_fm)||`,
		ParseMode: "MarkdownV2",
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

		//ReleaseDate: time.Now(), TODO отправлять с саундклауда
		//Tags:        []string{}, TODO отправлять с саундклауда
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
	fileName := filepath.Join("/app/images", time.Now().Format("20060102_150405")+".jpg")
	err = os.WriteFile(fileName, img, 0644)
	if err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return fileName, nil
}
