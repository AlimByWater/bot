package group

import (
	"context"
	"elysium/internal/usecase/use_message"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"
	"time"

	emoji_gen_utils "elysium/internal/controller/telegram/emoji-gen-utils"
	"elysium/internal/controller/telegram/emoji-gen-utils/queue"
	"elysium/internal/controller/telegram/emoji-gen-utils/uploader"
	"elysium/internal/entity"

	"github.com/cavaliergopher/grab/v3"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"
)

// EmojiChat отвечает за генерацию паков в групповом чате.
// Локализационные методы имеют префикс EmojiGen.
type EmojiChat struct {
	logger   *slog.Logger
	uploader *uploader.Module
	queue    *queue.StickerQueue
	userBot  interface {
		SendMessage(ctx context.Context, chat string, params telego.SendMessageParams) error
		SendMessageWithEmojis(ctx context.Context, chatID string, width int, packLink string, emojis []entity.EmojiMeta, replyTo int) error
	}
	processor interface {
		ParseArgs(arg string) (*entity.EmojiCommand, error)
		ExtractCommandArgs(msgText, msgCaption string) string
		SetupEmojiCommand(args *entity.EmojiCommand, user entity.User) *entity.EmojiCommand
		ProcessVideo(ctx context.Context, args *entity.EmojiCommand) ([]string, error)
		RegisterDirectory(dir string) error
	}
	userUC interface {
		UserByTelegramID(ctx context.Context, userID int64) (entity.User, error)
		CanGenerateEmojiPack(ctx context.Context, user entity.User, chatID int64) (bool, error)
	}
	repo interface {
		GetEmojiPackByPackLink(ctx context.Context, packLink string) (entity.EmojiPack, error)
		CreateNewEmojiPack(ctx context.Context, pack entity.EmojiPack) (entity.EmojiPack, error)
		UpdateEmojiCount(ctx context.Context, pack int64, emojiCount int) error
	}
	progressManager interface {
		SendMessage(ctx context.Context, b *telego.Bot, chatID telego.ChatID, replyToID int, userID int64, status string) (*entity.ProgressMessage, error)
		UpdateMessage(ctx context.Context, bot *telego.Bot, chatID telego.ChatID, msgID int, status string) error
		DeleteMessage(ctx context.Context, bot *telego.Bot, chatID telego.ChatID, msgID int) error
	}
}

// NewEmojiChat создаёт новый экземпляр EmojiChat.
func NewEmojiChat(
	userUC interface {
		UserByTelegramID(ctx context.Context, userID int64) (entity.User, error)
		CanGenerateEmojiPack(ctx context.Context, user entity.User, chatID int64) (bool, error)
	},
	processor interface {
		ParseArgs(arg string) (*entity.EmojiCommand, error)
		ExtractCommandArgs(msgText, msgCaption string) string
		SetupEmojiCommand(args *entity.EmojiCommand, user entity.User) *entity.EmojiCommand
		ProcessVideo(ctx context.Context, args *entity.EmojiCommand) ([]string, error)
		RegisterDirectory(dir string) error
	},
	repo interface {
		GetEmojiPackByPackLink(ctx context.Context, packLink string) (entity.EmojiPack, error)
		CreateNewEmojiPack(ctx context.Context, pack entity.EmojiPack) (entity.EmojiPack, error)
		UpdateEmojiCount(ctx context.Context, pack int64, emojiCount int) error
	},
	userBot interface {
		SendMessage(ctx context.Context, chat string, params telego.SendMessageParams) error
		SendMessageWithEmojis(ctx context.Context, chatID string, width int, packLink string, emojis []entity.EmojiMeta, replyTo int) error
	},
	progressManager interface {
		SendMessage(ctx context.Context, b *telego.Bot, chatID telego.ChatID, replyToID int, userID int64, status string) (*entity.ProgressMessage, error)
		UpdateMessage(ctx context.Context, bot *telego.Bot, chatID telego.ChatID, msgID int, status string) error
		DeleteMessage(ctx context.Context, bot *telego.Bot, chatID telego.ChatID, msgID int) error
	},
) *EmojiChat {
	queueModule := queue.New()
	uploaderModule := uploader.New()
	return &EmojiChat{
		userBot:         userBot,
		uploader:        uploaderModule,
		queue:           queueModule,
		progressManager: progressManager,
		processor:       processor,
		userUC:          userUC,
		repo:            repo,
	}
}

func (h *EmojiChat) AddLogger(logger *slog.Logger) {
	h.uploader.AddLogger(logger)
	h.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(h)).Type().PkgPath()))
}

func (h *EmojiChat) Command() string {
	return "emoji"
}

var validChatIDs = []int64{-1002400904088, -1002491830452}

func (h *EmojiChat) Predicate() telegohandler.Predicate {
	return func(update telego.Update) bool {
		if update.Message == nil {
			return false
		}
		// Обрабатываем только групповую переписку (не private)
		if update.Message.Chat.Type == "private" {
			return false
		}

		if slices.Contains(validChatIDs, update.Message.Chat.ID) {
			if update.Message.MessageThreadID == 3 {
				return true
			}
		}

		return false
	}
}

func (h *EmojiChat) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if update.Message == nil {
			return
		}

		lang := update.Message.From.LanguageCode

		// Получение пользователя
		user, err := h.userUC.UserByTelegramID(ctx, update.Message.From.ID)
		if err != nil {
			h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenError(lang))
			return
		}

		botUser, ok := update.Context().Value(entity.BotSelfCtxKey).(*telego.User)
		if !ok || botUser == nil {
			h.logger.Error("bot info not found in context")
			return
		}

		// Проверка возможности генерации
		canGenerate, err := h.userUC.CanGenerateEmojiPack(ctx, user, update.Message.Chat.ID)
		if err != nil {
			h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenError(lang))
			return
		}
		if !canGenerate {
			h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenLimitExceeded(lang))
			return
		}

		// Парсинг аргументов
		initialCommand := h.processor.ExtractCommandArgs(update.Message.Text, update.Message.Caption)
		emojiArgs, err := h.processor.ParseArgs(initialCommand)
		if err != nil {
			h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenError(lang))
			return
		}

		h.processor.SetupEmojiCommand(emojiArgs, user)

		// Подготовка окружения
		if err := h.prepareWorkingEnvironment(ctx, bot, &update, emojiArgs); err != nil {
			h.handleDownloadError(ctx, bot, &update, err)
			return
		}

		// Работа с паком
		var emojiPack entity.EmojiPack
		if strings.Contains(emojiArgs.PackLink, botUser.Username) {
			emojiPack, err = h.handleExistingPack(ctx, emojiArgs)
		} else {
			emojiPack, err = h.handleNewPack(ctx, botUser, user, emojiArgs)
		}
		if err != nil {
			h.logger.Error("Failed to handle new pack", slog.String("err", err.Error()))
			h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.Error(lang))
			return
		}

		// Отправка progressMessage
		progress, err := h.progressManager.SendMessage(ctx, bot, update.Message.Chat.ChatID(), update.Message.MessageID, user.TelegramID, use_message.GL.EmojiGenProcessingStart(lang))
		if err != nil {
			h.logger.Error("Failed to send progress message", slog.String("err", err.Error()))
		}
		defer func() {
			err := h.progressManager.DeleteMessage(ctx, bot, update.Message.Chat.ChatID(), progress.MessageID)
			if err != nil {
				h.logger.Error("Failed to delete progress message", slog.String("err", err.Error()))
			}
		}()

		// Acquire processing lock для пака
		canProcess, waitCh := h.queue.Acquire(emojiArgs.PackLink)
		if !canProcess {
			h.logger.Debug("Queue waiting", slog.String("pack_link", emojiArgs.PackLink))
			select {
			case <-ctx.Done():
				h.queue.Release(emojiArgs.PackLink)
				return
			case <-waitCh:
				h.logger.Debug("Queue released, processing started", slog.String("pack_link", emojiArgs.PackLink))
			}
		}
		defer h.queue.Release(emojiArgs.PackLink)

		// Основной цикл обработки
		var stickerSet *telego.StickerSet
		var emojiMetaRows [][]entity.EmojiMeta
		for {
			h.progressManager.UpdateMessage(ctx, bot, update.Message.Chat.ChatID(), progress.MessageID, use_message.GL.EmojiGenProcessingVideo(lang))
			files, err := h.processor.ProcessVideo(ctx, emojiArgs)
			if err != nil {
				h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID,
					use_message.GL.EmojiGenError(lang)+" "+err.Error())
				return
			}

			h.progressManager.UpdateMessage(ctx, bot, update.Message.Chat.ChatID(), progress.MessageID, use_message.GL.EmojiGenUploadingEmojis(lang))
			stickerSet, emojiMetaRows, err = h.uploader.AddEmojis(ctx, bot, emojiArgs, files)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				if ue, ok := err.(*uploader.UploaderError); ok {
					switch ue.Code {
					case uploader.ErrCodeNoFiles:
						h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenNoFiles(lang))
					case uploader.ErrCodeExceedLimit:
						if total, ok1 := ue.Params["totalStickers"]; ok1 {
							if max, ok2 := ue.Params["max"]; ok2 {
								h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, fmt.Sprintf(use_message.GL.EmojiGenEmojiInPackLimitExceeded(lang), total, max))
							} else {
								h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenEmojiInPackLimitExceeded(lang))
							}
						} else {
							h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenEmojiInPackLimitExceeded(lang))
						}
					case uploader.ErrCodeUploadSticker:
						h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenUploadStickerError(lang))
					case uploader.ErrCodeOpenFile:
						h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenOpenFileError(lang))
					case uploader.ErrCodeUploadTransparentSticker:
						h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenUploadTransparentStickerError(lang))
					case uploader.ErrCodeCreateStickerSet:
						h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenCreateStickerSetError(lang))
					case uploader.ErrCodeAddStickers:
						h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenAddStickersError(lang))
					case uploader.ErrCodeGetStickerSet:
						h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenGetStickerSetError(lang))
					default:
						h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenUploadError(lang)+": "+ue.Error())
					}
				} else {
					h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, use_message.GL.EmojiGenUploadError(lang)+": "+err.Error())
				}
				return
			}
			break
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		// Обновляем количество эмодзи в базе данных
		if err := h.repo.UpdateEmojiCount(ctx, emojiPack.ID, len(stickerSet.Stickers)); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			h.logger.Error("Failed to update emoji count",
				slog.String("err", err.Error()),
				slog.String("pack_link", emojiArgs.PackLink),
				slog.Int64("user_id", emojiArgs.TelegramUserID),
			)
		}

		// Отправка сообщения с эмодзи через userbot
		var topicID string
		if update.Message.MessageThreadID != 0 {
			topicID = fmt.Sprintf("%d_%d", update.Message.Chat.ID, update.Message.MessageThreadID)
		} else {
			topicID = fmt.Sprintf("%d", update.Message.Chat.ID)
		}

		// В данном примере выбранные эмодзи - первая строка
		selectedEmojis := emoji_gen_utils.GenerateEmojiMessage(emojiMetaRows, stickerSet, emojiArgs)

		err = h.userBot.SendMessageWithEmojis(ctx, topicID, emojiArgs.Width, emojiArgs.PackLink, selectedEmojis, update.Message.MessageID)
		if err != nil {
			h.logger.Error("Failed to send message with emojis",
				slog.String("err", err.Error()),
				slog.String("username", update.Message.From.Username),
				slog.Int64("user_id", update.Message.From.ID))
		}
	}
}

func (h *EmojiChat) sendErrorMessage(ctx context.Context, chatID int64, replyTo int, threadID int, errMsg string) {
	params := telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   fmt.Sprintf("%s", errMsg),
	}

	if replyTo != 0 {
		params.ReplyParameters = &telego.ReplyParameters{
			MessageID: replyTo,
			ChatID:    telego.ChatID{ID: chatID},
		}
	}

	chat := fmt.Sprintf("%d", chatID)
	if threadID != 0 {
		chat = fmt.Sprintf("%s_%d", chat, threadID)
	}

	err := h.userBot.SendMessage(ctx, chat, params)
	if err != nil {
		if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "context canceled") {
			return
		}
		slog.Error("Failed to send error message", slog.String("err", err.Error()))
	}
}

func (h *EmojiChat) handleDownloadError(ctx context.Context, bot *telego.Bot, update *telego.Update, err error) {
	lang := update.Message.From.LanguageCode
	var message string
	switch {
	case errors.Is(err, entity.ErrFileOfInvalidType):
		message = use_message.GL.EmojiGenUnsupportedFileType(lang)
	case errors.Is(err, entity.ErrFileNotProvided):
		return
	default:
		h.logger.Error("Download error", slog.String("err", err.Error()))
		message = use_message.GL.EmojiGenDownloadError(lang)
	}
	h.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.MessageID, update.Message.MessageThreadID, message)
}

// Добавленные методы для EmojiDM
func (h *EmojiChat) prepareWorkingEnvironment(ctx context.Context, bot *telego.Bot, update *telego.Update, args *entity.EmojiCommand) error {
	workingDir := fmt.Sprintf("/tmp/%d_%d", update.Message.Chat.ID, time.Now().Unix())
	if err := h.processor.RegisterDirectory(workingDir); err != nil {
		return fmt.Errorf("failed to register working directory: %w", err)
	}
	args.WorkingDir = workingDir

	fileName, err := h.downloadFile(ctx, bot, update.Message, args)
	if err != nil {
		return err
	}
	args.DownloadedFile = fileName
	return nil
}

func (h *EmojiChat) handleNewPack(ctx context.Context, botUser *telego.User, user entity.User, args *entity.EmojiCommand) (entity.EmojiPack, error) {
	args.NewSet = true
	packName := fmt.Sprintf("%s%d_by_%s", "dt", time.Now().Unix(), botUser.Username)
	if len(packName) > entity.TelegramPackLinkAndNameLength {
		args.PackLink = args.PackLink[:len(packName)-entity.TelegramPackLinkAndNameLength]
		packName = fmt.Sprintf("%s_%s", args.PackLink, botUser.Username)
	}
	args.PackLink = packName

	emojiPack := entity.EmojiPack{
		CreatorTelegramID: user.TelegramID,
		PackTitle:         args.SetTitle,
		PackLink:          packName,
		InitialCommand:    args.InitialCommand,
		EmojiCount:        0,
		BotID:             botUser.ID,
		BotUsername:       botUser.Username,
	}

	return h.repo.CreateNewEmojiPack(ctx, emojiPack)
}

func (h *EmojiChat) handleExistingPack(ctx context.Context, args *entity.EmojiCommand) (entity.EmojiPack, error) {
	args.NewSet = false
	if strings.Contains(args.PackLink, "t.me/addemoji/") {
		splited := strings.Split(args.PackLink, ".me/addemoji/")
		args.PackLink = strings.TrimSpace(splited[len(splited)-1])
	}

	pack, err := h.repo.GetEmojiPackByPackLink(ctx, args.PackLink)
	if err != nil {
		return entity.EmojiPack{}, err
	}
	args.SetTitle = ""
	return pack, nil
}

func (h *EmojiChat) downloadFile(ctx context.Context, bot *telego.Bot, msg *telego.Message, args *entity.EmojiCommand) (string, error) {
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
