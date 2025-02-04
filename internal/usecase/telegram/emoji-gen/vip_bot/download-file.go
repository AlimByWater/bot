package vip_bot

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"github.com/cavaliergopher/grab/v3"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"slices"
)

func (d *DBot) prepareWorkingEnvironment(ctx context.Context, update *models.Update, args *entity.EmojiCommand) error {
	if err := d.processor.RegisterDirectory(args.WorkingDir); err != nil {
		return fmt.Errorf("failed to register working directory: %w", err)
	}

	fileName, err := d.downloadFile(ctx, update.Message, args)
	if err != nil {
		return err
	}
	args.DownloadedFile = fileName
	return nil
}

func (d *DBot) handleDownloadError(ctx context.Context, update *models.Update, err error) {
	d.logger.Debug("Failed to download file", slog.String("err", err.Error()))
	var message string
	switch err {
	case entity.ErrFileNotProvided:
		return
	case entity.ErrFileOfInvalidType:
		message = "Неподдерживаемый тип файла. Поддерживаются: GIF, JPEG, PNG, WebP, MP4, WebM, MPEG"
	case entity.ErrGetFileFromTelegram:
		message = "Не удалось получить файл из Telegram"
	case entity.ErrFileDownloadFailed:
		message = "Ошибка при загрузке файла"
	default:
		message = "Ошибка при загрузке файла"
	}
	d.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.ID, update.Message.MessageThreadID, message)
}

func (d *DBot) downloadFile(ctx context.Context, m *models.Message, args *entity.EmojiCommand) (string, error) {
	var fileID string
	var fileExt string
	var mimeType string

	if m.Video != nil {
		fileID = m.Video.FileID
		mimeType = m.Video.MimeType
	} else if m.Photo != nil && len(m.Photo) > 0 {
		fileID = m.Photo[len(m.Photo)-1].FileID
		mimeType = "image/jpeg"

	} else if m.Document != nil {
		if slices.Contains(entity.AllowedMimeTypes, m.Document.MimeType) {
			fileID = m.Document.FileID
			mimeType = m.Document.MimeType
		} else {
			return "", entity.ErrFileOfInvalidType
		}
	} else if m.ReplyToMessage != nil {
		if m.ReplyToMessage.Video != nil {
			fileID = m.ReplyToMessage.Video.FileID
			mimeType = m.ReplyToMessage.Video.MimeType
		} else if m.ReplyToMessage.Photo != nil && len(m.ReplyToMessage.Photo) > 0 {
			fileID = m.ReplyToMessage.Photo[len(m.ReplyToMessage.Photo)-1].FileID
			mimeType = "image/jpeg"
		} else if m.ReplyToMessage.Document != nil {
			fileID = m.ReplyToMessage.Document.FileID
			mimeType = m.ReplyToMessage.Document.MimeType
		} else if m.ReplyToMessage.Sticker != nil && m.ReplyToMessage.Sticker.Type == "regular" {
			fileID = m.ReplyToMessage.Sticker.FileID
			if m.ReplyToMessage.Sticker.IsVideo {
				mimeType = "video/webm"
			} else if !m.ReplyToMessage.Sticker.IsAnimated {
				mimeType = "image/webp"
			}
		} else {
			return "", entity.ErrFileNotProvided
		}

	} else {
		return "", entity.ErrFileNotProvided
	}

	file, err := d.b.BotApi.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("%w: %w", entity.ErrGetFileFromTelegram, err)
	}
	//args.File = file

	switch mimeType {
	case "image/gif":
		fileExt = ".gif"
	case "image/jpeg":
		fileExt = ".jpg"
	case "image/png":
		fileExt = ".png"
	case "image/webp":
		fileExt = ".webp"
	case "video/mp4":
		fileExt = ".mp4"
	case "video/webm":
		fileExt = ".webm"
	case "video/mpeg":
		fileExt = ".mpeg"
	default:
		return "", entity.ErrFileOfInvalidType
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", d.b.Token, file.FilePath)
	resp, err := grab.Get(args.WorkingDir+"/saved"+fileExt, fileURL)
	if err != nil {
		return "", fmt.Errorf("%w: %w", entity.ErrFileDownloadFailed, err)
	}

	return resp.Filename, nil
}
