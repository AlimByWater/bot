package demethra

import (
	"arimadj-helper/internal/application/logger"
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
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

	//coverPath, err := m.downloadCover(track.CoverLink)

	// ************* КАЧАЕМ ОБЛОЖКУ  *************** //
	res, err := http.Get(track.CoverLink)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "http get cover link", logger.AppendErrorToLogs(attributes, err)...)
	}
	defer res.Body.Close()

	img, err := io.ReadAll(res.Body)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "read img body", logger.AppendErrorToLogs(attributes, err)...)
	}

	//// Сохранение обложки на файловую систему
	fileName := filepath.Join("/app/images", time.Now().Format("20060102_150405")+".jpg")
	err = os.WriteFile(fileName, img, 0644)
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "write file", logger.AppendErrorToLogs(attributes, err)...)
	}
	// ****************** ******************////

	m.prevTrack = m.currentTrack
	m.currentTrack = track

	err = m.bot.updateCurrentTrackMessage(m.currentTrack, m.prevTrack, "")
	if err != nil {
		m.logger.LogAttrs(ctx, slog.LevelError, "update current track", logger.AppendErrorToLogs(attributes, err)...)
	}

	//err := m.repo.SongPlayed(track)

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
