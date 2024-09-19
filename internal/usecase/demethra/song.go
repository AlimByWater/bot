package demethra

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (m *Module) SendSongByTrackLink(ctx context.Context, userID int, trackLink string) error {
	user, err := m.repo.GetUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("get user by id: %w", err)
	}

	if user.TelegramID == 0 {
		return fmt.Errorf("user telegram id is empty")
	}

	song, err := m.repo.SongByUrl(ctx, trackLink)
	if err != nil {
		return fmt.Errorf("get song by url: %w", err)
	}

	if song.ID == 0 {
		return fmt.Errorf("song not found")
	}

	err = m.Bot.sendSongToTelegramUser(ctx, user.TelegramID, song)
	if err != nil {
		return fmt.Errorf("send song to chat: %w", err)
	}

	err = m.repo.LogSongDownload(ctx, user.ID, song.ID, entity.SongDownloadSourceWebApp)
	if err != nil {
		m.logger.Error("Failed to log song download", slog.String("error", err.Error()), slog.Int("user_id", user.ID), slog.Int("song_id", song.ID), slog.String("method", "SendSongByTrackLink"))
	}

	return nil
}

func (b *Bot) sendSongToTelegramUser(_ context.Context, telegramUserID int64, song entity.Song) error {
	forwardMsg := tgbotapi.NewForward(telegramUserID, song.SongTelegramMessageChatID, song.SongTelegramMessageID)

	_, err := b.Api.Send(forwardMsg)
	if err != nil {
		return fmt.Errorf("forward message: %w", err)
	}

	return nil
}
