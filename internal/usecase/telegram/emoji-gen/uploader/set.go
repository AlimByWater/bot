package uploader

import (
	"context"
	"elysium/internal/entity"
	"errors"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"strings"
	"time"
)

// createNewStickerSet создает новый набор стикеров
func (m *Module) createNewStickerSet(ctx context.Context, b *bot.Bot, args *entity.EmojiCommand, emojiFileIDs []string) (*models.StickerSet, error) {
	totalWithTransparent := len(emojiFileIDs)
	if totalWithTransparent > entity.MaxStickersTotal {
		return nil, fmt.Errorf("общее количество стикеров (%d) с прозрачными превысит максимум (%d)", totalWithTransparent, entity.MaxStickersTotal)
	}

	return m.createStickerSetWithBatches(ctx, b, args, emojiFileIDs)
}

// createStickerSetWithBatches создает новый набор стикеров
func (m *Module) createStickerSetWithBatches(ctx context.Context, b *bot.Bot, args *entity.EmojiCommand, emojiFileIDs []string) (*models.StickerSet, error) {
	count := len(emojiFileIDs)
	if count > entity.MaxStickersInBatch {
		count = entity.MaxStickersInBatch
	}

	firstBatch := make([]models.InputSticker, count)

	for i := 0; i < count; i++ {
		firstBatch[i] = models.InputSticker{
			Sticker: &models.InputFileString{Data: emojiFileIDs[i]},
			Format:  entity.DefaultStickerFormat,
			EmojiList: []string{
				entity.DefaultEmojiIcon,
			},
		}
	}

	var err error

	_, err = b.CreateNewStickerSet(ctx, &bot.CreateNewStickerSetParams{
		UserID:      args.TelegramUserID,
		Name:        args.PackLink,
		Title:       args.SetTitle,
		StickerType: "custom_emoji",
		Stickers:    firstBatch,
	})

	if err != nil && !strings.Contains(err.Error(), "STICKER_VIDEO_NOWEBM") {
		slog.Debug("new sticker set FAILED", slog.String("name", args.PackLink), slog.String("error", err.Error()))
		return nil, fmt.Errorf("create sticker set: %w", err)
	} else if err != nil && strings.Contains(err.Error(), "STICKER_VIDEO_NOWEBM") {
		count = 1

		_, err := b.CreateNewStickerSet(ctx, &bot.CreateNewStickerSetParams{
			UserID:      args.TelegramUserID,
			Name:        args.PackLink,
			Title:       args.SetTitle,
			StickerType: "custom_emoji",
			Stickers:    firstBatch,
		})
		if err != nil {
			slog.Debug("new sticker set FAILED", slog.String("name", args.PackLink), slog.String("error", err.Error()))
			return nil, fmt.Errorf("create sticker set: %w", err)
		}
	}

	emojiFileIDs = emojiFileIDs[count:]

	// Добавляем оставшиеся стикеры по одному
	err = m.addStickersToSet(ctx, b, args, emojiFileIDs)
	if err != nil {
		return nil, fmt.Errorf("add stickers to set: %w", err)
	}

	// Получаем финальное состояние набора
	set, err := b.GetStickerSet(ctx, &bot.GetStickerSetParams{
		Name: args.PackLink,
	})
	if err != nil {
		return nil, fmt.Errorf("get sticker set: %w", err)
	}

	return set, nil
}

// addToExistingStickerSet добавляет эмодзи в существующий набор
func (m *Module) addToExistingStickerSet(ctx context.Context, b *bot.Bot, args *entity.EmojiCommand, stickerSet *models.StickerSet, emojiFileIDs []string) (*models.StickerSet, error) {

	// Проверяем, что не превысим лимит
	if len(stickerSet.Stickers)+len(emojiFileIDs) > entity.MaxStickersTotal {
		return nil, fmt.Errorf(
			"превышен лимит стикеров в наборе (%d + %d > %d)",
			len(stickerSet.Stickers),
			len(emojiFileIDs),
			entity.MaxStickersTotal,
		)
	}

	// Добавляем стикеры батчами
	err := m.addStickersToSet(ctx, b, args, emojiFileIDs)
	if err != nil {
		return nil, fmt.Errorf("add stickers to set: %w", err)
	}

	return b.GetStickerSet(ctx, &bot.GetStickerSetParams{
		Name: args.PackLink,
	})
}

var maxRetries = 5

func (m *Module) addStickersToSet(ctx context.Context, b *bot.Bot, args *entity.EmojiCommand, emojiFileIDs []string) error {
	for i := 0; i < len(emojiFileIDs); i++ {

		var err error
		for j := 1; j <= maxRetries; j++ {
			_, err = b.AddStickerToSet(ctx, &bot.AddStickerToSetParams{
				UserID: args.TelegramUserID,
				Name:   args.PackLink,
				Sticker: models.InputSticker{
					Sticker: &models.InputFileString{Data: emojiFileIDs[i]},
					Format:  entity.DefaultStickerFormat,
					EmojiList: []string{
						entity.DefaultEmojiIcon,
					},
				},
			})
			if err == nil || errors.Is(err, context.Canceled) {
				//slog.Debug("add sticker to set SUCCESS",
				//	slog.String("file_id", emojiFileIDs[i]),
				//	slog.String("pack", args.PackLink),
				//	slog.Int64("user_id", args.UserID),
				//)

				break
			} else {
				slog.Debug("error sending sticker", "err", err.Error())
				time.Sleep(time.Second * 1)
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}
