package uploader

import (
	"bytes"
	"context"
	"elysium/internal/entity"
	"errors"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"os"
)

func (m *Module) uploadSticker(ctx context.Context, b *bot.Bot, userID int64, filename string, data []byte) (string, error) {
	newSticker, err := b.UploadStickerFile(ctx, &bot.UploadStickerFileParams{
		UserID: userID,
		Sticker: &models.InputFileUpload{
			Filename: filename,
			Data:     bytes.NewReader(data),
		},
		StickerFormat: entity.DefaultStickerFormat,
	})

	if err != nil {
		if errors.Is(err, context.Canceled) {
			return "", context.Canceled
		}
		return "", fmt.Errorf("upload sticker: %w; filename: %s", err, filename)
	}

	return newSticker.FileID, nil
}

// uploadEmojiFiles загружает все файлы эмодзи и возвращает их fileIDs и метаданные
func (m *Module) uploadEmojiFiles(ctx context.Context, b *bot.Bot, args *entity.EmojiCommand, set *models.StickerSet, emojiFiles []string) ([]string, [][]entity.EmojiMeta, error) {
	m.logger.Debug("uploading emoji stickers", slog.Int("count", len(emojiFiles)))

	totalEmojis := len(emojiFiles)
	rows := (totalEmojis + args.Width - 1) / args.Width // Округляем вверх
	emojiMetaRows := make([][]entity.EmojiMeta, rows)

	// Проверка на превышение максимального количества стикеров
	totalStickers := len(emojiFiles)
	if args.Width < entity.DefaultWidth {
		totalStickers += (entity.DefaultWidth - args.Width) * rows
	}

	if set != nil {
		if set.Stickers != nil {
			totalStickers += len(set.Stickers)
		}
	}

	if totalStickers > entity.MaxStickersTotal {
		return nil, nil, fmt.Errorf("будет превышено максимальное количество эмодзи в паке (%d из %d)", totalStickers, entity.MaxStickersTotal)
	}

	// Подготавливаем прозрачный стикер только если он нужен
	var transparentData []byte
	var err error
	if args.Width < entity.DefaultWidth {
		transparentData, err = PrepareTransparentData(args.Width)
		if err != nil {
			return nil, nil, err
		}
	}

	for i := range emojiMetaRows {
		if args.Width < entity.DefaultWidth {
			emojiMetaRows[i] = make([]entity.EmojiMeta, entity.DefaultWidth) // Инициализируем каждый ряд с полной шириной
		} else {
			emojiMetaRows[i] = make([]entity.EmojiMeta, args.Width) // Инициализируем каждый ряд с полной шириной
		}
	}

	// Сначала загружаем все эмодзи и заполняем метаданные
	for i, emojiFile := range emojiFiles {
		fileData, err := os.ReadFile(emojiFile)
		if err != nil {
			return nil, nil, fmt.Errorf("open emoji file: %w", err)
		}

		fileID, err := m.uploadSticker(ctx, b, args.TelegramUserID, emojiFile, fileData)
		if err != nil {
			return nil, nil, err
		} else {
			//slog.Debug("upload sticker SUCCESS",
			//	slog.String("file", emojiFile),
			//	slog.String("pack", args.PackLink),
			//	slog.Int64("user_id", args.UserID),
			//	slog.Bool("transparent", false),
			//)
		}

		// Вычисляем позицию в сетке
		row := i / args.Width
		col := i % args.Width

		// Вычисляем отступы для центрирования
		totalPadding := entity.DefaultWidth - args.Width
		leftPadding := totalPadding / 2
		if totalPadding > 0 && totalPadding%2 != 0 {
			// Для нечетного количества отступов, слева меньше на 1
			leftPadding = (totalPadding - 1) / 2
		}
		transparentFileID, err := m.uploadSticker(ctx, b, args.TelegramUserID, "transparent.webm", transparentData)
		if err != nil {
			return nil, nil, fmt.Errorf("upload transparent sticker: %w", err)
		}

		// Загружаем прозрачные эмодзи слева только если нужно
		if args.Width < entity.DefaultWidth {
			for j := 0; j < leftPadding; j++ {
				if emojiMetaRows[row][j].FileID == "" {

					emojiMetaRows[row][j] = entity.EmojiMeta{
						FileID:      transparentFileID,
						FileName:    "transparent.webm",
						Transparent: true,
					}
				}
			}
		}

		// Записываем метаданные эмодзи
		pos := col
		if args.Width < entity.DefaultWidth {
			pos = col + leftPadding
		}
		emojiMetaRows[row][pos] = entity.EmojiMeta{
			FileID:      fileID,
			FileName:    emojiFile,
			Transparent: false,
		}

		// Загружаем прозрачные эмодзи справа только если нужно
		if args.Width < entity.DefaultWidth {
			for j := col + leftPadding + 1; j < entity.DefaultWidth; j++ {
				if emojiMetaRows[row][j].FileID == "" {
					emojiMetaRows[row][j] = entity.EmojiMeta{
						FileID:      transparentFileID,
						FileName:    "transparent.webm",
						Transparent: true,
					}
				}
			}
		}

	}

	// Теперь собираем emojiFileIDs в правильном порядке
	emojiFileIDs := make([]string, 0, rows*entity.DefaultWidth)
	for i := range emojiMetaRows {
		for j := range emojiMetaRows[i] {
			if emojiMetaRows[i][j].FileID != "" {
				emojiFileIDs = append(emojiFileIDs, emojiMetaRows[i][j].FileID)
			}
		}
	}

	return emojiFileIDs, emojiMetaRows, nil
}
