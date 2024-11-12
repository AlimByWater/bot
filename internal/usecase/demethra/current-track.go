package demethra

import (
	"context"
	"elysium/internal/application/logger"
	"elysium/internal/entity"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
)

func (m *Module) CurrentTrackForStream(slug string) entity.TrackInfo {
	stream, ok := m.streams[slug]
	if !ok {
		return entity.TrackInfo{}
	}

	stream.RLock()
	defer stream.RUnlock()

	return stream.CurrentTrack
}

func (b *Bot) sendCurrentTrackMessage(ctx context.Context, chatID int64, songID int, current, prev entity.TrackInfo, coverFileID string, attrs []slog.Attr) (tgbotapi.Message, error) {
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
	radioBtn := tgbotapi.NewInlineKeyboardButtonWebApp("Радио", tgbotapi.WebAppInfo{
		URL: fmt.Sprintf("https://t.me/elysium_demethra_bot/radio"),
	})
	row := tgbotapi.NewInlineKeyboardRow(btn, radioBtn)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)

	msg := tgbotapi.PhotoConfig{
		BaseFile: tgbotapi.BaseFile{
			BaseChat: tgbotapi.BaseChat{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: chatID,
				},
				ReplyMarkup: keyboard,
			},
			File: cover,
		},
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

	responseMsg, err := b.Api.Send(msg)
	if err != nil {
		return tgbotapi.Message{}, fmt.Errorf("send: %w", err)
	}

	return responseMsg, nil
}
