package uploader

import (
	"context"
	"elysium/internal/entity"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegoutil"
)

// createNewStickerSet создает новый набор стикеров
func (m *Module) createNewStickerSet(ctx context.Context, b *telego.Bot, args *entity.EmojiCommand, emojiFileIDs []string) (*telego.StickerSet, error) {
	totalWithTransparent := len(emojiFileIDs)
	if totalWithTransparent > entity.MaxStickersTotal {
		return nil, fmt.Errorf("общее количество стикеров (%d) с прозрачными превысит максимум (%d)", totalWithTransparent, entity.MaxStickersTotal)
	}

	return m.createStickerSetWithBatches(ctx, b, args, emojiFileIDs)
}

// createStickerSetWithBatches создает новый набор стикеров
func (m *Module) createStickerSetWithBatches(ctx context.Context, b *telego.Bot, args *entity.EmojiCommand, emojiFileIDs []string) (*telego.StickerSet, error) {
	count := len(emojiFileIDs)
	if count > entity.MaxStickersInBatch {
		count = entity.MaxStickersInBatch
		//count = 1
	}

	firstBatch := make([]telego.InputSticker, count)

	for i := 0; i < count; i++ {
		firstBatch[i] = telego.InputSticker{
			Sticker: telegoutil.FileFromID(emojiFileIDs[i]),
			Format:  entity.DefaultStickerFormat,
			EmojiList: []string{
				entity.DefaultEmojiIcon,
			},
		}
	}

	var err error
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	err = b.CreateNewStickerSet(&telego.CreateNewStickerSetParams{
		UserID:      args.TelegramUserID,
		Name:        args.PackLink,
		Title:       args.SetTitle,
		StickerType: "custom_emoji",
		Stickers:    firstBatch,
	})

	if err != nil && !strings.Contains(err.Error(), "STICKER_VIDEO_NOWEBM") {
		m.logger.Debug("new sticker set FAILED", slog.Int("t", 1), slog.String("name", args.PackLink), slog.String("error", err.Error()))
		return nil, fmt.Errorf("create sticker set: %w", err)
	} else if err != nil && strings.Contains(err.Error(), "STICKER_VIDEO_NOWEBM") {
		count = 1

		err := b.CreateNewStickerSet(&telego.CreateNewStickerSetParams{
			UserID:      args.TelegramUserID,
			Name:        args.PackLink,
			Title:       args.SetTitle,
			StickerType: "custom_emoji",
			Stickers:    firstBatch,
		})
		if err != nil {
			m.logger.Debug("new sticker set FAILED", slog.Int("t", 2), slog.String("name", args.PackLink), slog.String("error", err.Error()))
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
	set, err := b.GetStickerSet(&telego.GetStickerSetParams{
		Name: args.PackLink,
	})
	if err != nil {
		return nil, fmt.Errorf("get sticker set: %w", err)
	}

	return set, nil
}

// addToExistingStickerSet добавляет эмодзи в существующий набор
func (m *Module) addToExistingStickerSet(ctx context.Context, b *telego.Bot, args *entity.EmojiCommand, stickerSet *telego.StickerSet, emojiFileIDs []string) (*telego.StickerSet, error) {

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

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	return b.GetStickerSet(&telego.GetStickerSetParams{
		Name: args.PackLink,
	})
}

var maxRetries = 5

func (m *Module) addStickersToSet(ctx context.Context, b *telego.Bot, args *entity.EmojiCommand, emojiFileIDs []string) error {
	for i := 0; i < len(emojiFileIDs); i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var err error
		for j := 1; j <= maxRetries; j++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			err = b.AddStickerToSet(&telego.AddStickerToSetParams{
				UserID: args.TelegramUserID,
				Name:   args.PackLink,
				Sticker: telego.InputSticker{
					Sticker: telegoutil.FileFromID(emojiFileIDs[i]),
					Format:  entity.DefaultStickerFormat,
					EmojiList: []string{
						entity.DefaultEmojiIcon,
					},
				},
			})
			if err == nil || errors.Is(err, context.Canceled) {
				break
			} else {
				m.logger.Debug("error sending sticker", "err", err.Error())
				time.Sleep(time.Second * 1)
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}
