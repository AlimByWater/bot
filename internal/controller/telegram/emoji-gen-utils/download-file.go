package emoji_gen_utils

import (
	"elysium/internal/entity"
	"fmt"
	"github.com/cavaliergopher/grab/v3"
	"github.com/mymmrac/telego"
	"slices"
)

func DownloadFile(bot *telego.Bot, msg *telego.Message, args *entity.EmojiCommand) (string, error) {
	var fileID string
	var mimeType string

	switch {
	case msg.Video != nil:
		fileID = msg.Video.FileID
		mimeType = msg.Video.MimeType
	case msg.Photo != nil && len(msg.Photo) > 0:
		fileID = msg.Photo[len(msg.Photo)-1].FileID
		mimeType = "image/jpeg"
	case msg.Document != nil:
		if slices.Contains(entity.AllowedMimeTypes, msg.Document.MimeType) {
			fileID = msg.Document.FileID
			mimeType = msg.Document.MimeType
		} else {
			return "", entity.ErrFileOfInvalidType
		}
	case msg.ReplyToMessage != nil:
		switch {
		case msg.ReplyToMessage.Video != nil:
			fileID = msg.ReplyToMessage.Video.FileID
			mimeType = msg.ReplyToMessage.Video.MimeType
		case msg.ReplyToMessage.Photo != nil && len(msg.ReplyToMessage.Photo) > 0:
			fileID = msg.ReplyToMessage.Photo[len(msg.ReplyToMessage.Photo)-1].FileID
			mimeType = "image/jpeg"
		case msg.ReplyToMessage.Document != nil:
			fileID = msg.ReplyToMessage.Document.FileID
			mimeType = msg.ReplyToMessage.Document.MimeType
		case msg.ReplyToMessage.Sticker != nil && msg.ReplyToMessage.Sticker.Type == "regular":
			fileID = msg.ReplyToMessage.Sticker.FileID
			if msg.ReplyToMessage.Sticker.IsVideo {
				mimeType = "video/webm"
			} else if !msg.ReplyToMessage.Sticker.IsAnimated {
				mimeType = "image/webp"
			}
		default:
			return "", entity.ErrFileNotProvided
		}
	default:
		return "", entity.ErrFileNotProvided
	}

	// Get file info from Telegram
	file, err := bot.GetFile(&telego.GetFileParams{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("%w: %w", entity.ErrGetFileFromTelegram, err)
	}
	args.File = file

	// Determine file extension
	var fileExt string
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

	// Download file
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", bot.Token(), file.FilePath)
	resp, err := grab.Get(args.WorkingDir+"/saved"+fileExt, fileURL)
	if err != nil {
		return "", fmt.Errorf("%w: %w", entity.ErrFileDownloadFailed, err)
	}

	return resp.Filename, nil
}
