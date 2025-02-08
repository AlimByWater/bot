package uploader

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"log/slog"
	"os"

	"github.com/mymmrac/telego"
)

type Queuer interface {
	Acquire(packLink string) (bool, chan struct{})
	Release(packLink string)
}

type Module struct {
	logger *slog.Logger
}

func New() *Module {
	return &Module{}
}

func (m *Module) AddLogger(logger *slog.Logger) {
	m.logger = logger.With(slog.String("module", "emoji-uploader"))
}

func (m *Module) AddEmojis(ctx context.Context, b *telego.Bot, args *entity.EmojiCommand, emojiFiles []string) (*telego.StickerSet, [][]entity.EmojiMeta, error) {
	select {
	case <-ctx.Done():
		return nil, nil, nil
	default:
	}

	if err := m.ValidateEmojiFiles(emojiFiles); err != nil {
		return nil, nil, err
	}

	var set *telego.StickerSet
	if !args.NewSet {
		var err error
		set, err = b.GetStickerSet(&telego.GetStickerSetParams{
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
		return &UploaderError{Code: "NoFiles", Err: fmt.Errorf("no files provided")}
	}
	if len(emojiFiles) > entity.MaxStickersTotal {
		return &UploaderError{Code: "ExceedLimit", Err: fmt.Errorf("files exceed allowed limit")}
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
