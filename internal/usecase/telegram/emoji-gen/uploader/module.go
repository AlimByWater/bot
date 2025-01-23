package uploader

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"os"
)

type Queuer interface {
	Acquire(packLink string) (bool, chan struct{})
	Release(packLink string)
	Clear()
}

type Module struct {
	stickerQueue Queuer
	logger       *slog.Logger
}

func New(queuer Queuer, logger *slog.Logger) *Module {
	return &Module{
		stickerQueue: queuer,
		logger:       logger,
	}

}

func (m *Module) AddEmojis(ctx context.Context, b *bot.Bot, args *entity.EmojiCommand, emojiFiles []string) (*models.StickerSet, [][]entity.EmojiMeta, error) {
	if err := m.ValidateEmojiFiles(emojiFiles); err != nil {
		return nil, nil, err
	}

	// Пытаемся получить доступ к обработке пака
	canProcess, waitCh := m.stickerQueue.Acquire(args.PackLink)
	if !canProcess {
		// Если нельзя обрабатывать сейчас - ждем своей очереди
		slog.Debug("В ОЧЕРЕДИ", slog.String("pack_link", args.PackLink))
		select {
		case <-ctx.Done():
			m.stickerQueue.Release(args.PackLink)
			return nil, nil, nil
		case <-waitCh:
			slog.Debug("ОЧЕРЕДЬ ПРИШЛА, НАЧИНАЕТСЯ ОБРАБОТКА", slog.String("pack_link", args.PackLink))
		}
	}
	defer m.stickerQueue.Release(args.PackLink)

	var set *models.StickerSet
	if !args.NewSet {
		var err error
		set, err = b.GetStickerSet(ctx, &bot.GetStickerSetParams{
			Name: args.PackLink,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("get sticker set: %w", err)
		}
	}

	// Загружаем все файлы эмодзи и возвращаем их fileIDs и метаданные
	emojiFileIDs, emojiMetaRows, err := m.uploadEmojiFiles(ctx, b, args, set, emojiFiles)
	if err != nil {
		return nil, nil, err
	}

	// Создаем набор стикеров

	if args.NewSet {
		set, err = m.createNewStickerSet(ctx, b, args, emojiFileIDs)
	} else {
		set, err = m.addToExistingStickerSet(ctx, b, args, set, emojiFileIDs)
	}
	if err != nil {
		return nil, nil, err
	}

	m.logger.Debug("addEmojis",
		slog.Int("emojiFileIDS count", len(emojiFileIDs)),
		slog.Int("width", args.Width),
		slog.Int("transparent_spacing", entity.DefaultWidth-args.Width),
		slog.Int("stickers in set", len(set.Stickers)),
		slog.String("working dir", args.WorkingDir))

	// Обновляем emojiMetaRows только для последних стикеров
	idx := 0
	if !args.NewSet {
		idx = len(set.Stickers) - len(emojiFileIDs) - 1
	}

	for i := range emojiMetaRows {
		for j := range emojiMetaRows[i] {
			emojiMetaRows[i][j].DocumentID = set.Stickers[idx].CustomEmojiID
			idx++
		}
	}

	return set, emojiMetaRows, nil
}

func (m *Module) ValidateEmojiFiles(emojiFiles []string) error {
	if len(emojiFiles) == 0 {
		return fmt.Errorf("нет файлов для создания набора")
	}

	if len(emojiFiles) > entity.MaxStickersTotal {
		return fmt.Errorf("слишком много файлов для создания набора (максимум %d)", entity.MaxStickersTotal)
	}

	return nil
}

func PrepareTransparentData(width int) ([]byte, error) {
	// Подготавливаем прозрачные стикеры если нужно
	transparentSpacing := entity.DefaultWidth - width
	transparentData, err := os.ReadFile("transparent.webm")
	if err != nil || transparentSpacing <= 0 {
		return nil, nil
	} else if transparentSpacing > 0 {
		transparentData, err = os.ReadFile("transparent.webm")
		if err != nil {
			return nil, fmt.Errorf("open transparent file: %w", err)
		}
		return transparentData, nil
	}

	return nil, nil
}
