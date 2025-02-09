package group

import (
	"context"
	emoji_gen_utils "elysium/internal/controller/telegram/emoji-gen-utils"
	"elysium/internal/controller/telegram/emoji-gen-utils/queue"
	"elysium/internal/controller/telegram/emoji-gen-utils/uploader"
	"elysium/internal/usecase/use_message"
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/cavaliergopher/grab/v3"
	"github.com/mymmrac/telego"
	"github.com/mymmrac/telego/telegohandler"

	"elysium/internal/entity"
)

type EmojiDM struct {
	logger   *slog.Logger
	uploader *uploader.Module
	queue    *queue.StickerQueue
	cache    interface {
		LoadAndDelete(key string) (value any, loaded bool)
	}
	userUC interface {
		UserByTelegramID(ctx context.Context, userID int64) (entity.User, error)
		CanGenerateEmojiPack(ctx context.Context, user entity.User, chatID int64) (bool, error)
	}
	processor interface {
		ParseArgs(arg string) (*entity.EmojiCommand, error)
		ExtractCommandArgs(msgText, msgCaption string) string
		SetupEmojiCommand(args *entity.EmojiCommand, user entity.User) *entity.EmojiCommand
		RegisterDirectory(dir string) error
		ProcessVideo(ctx context.Context, args *entity.EmojiCommand) ([]string, error)
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
		GetCancelChannel(cancelKey string) chan struct{}
	}
	userBot interface {
		GetID() int64
		SendMessageWithEmojisToBot(ctx context.Context, chatID string, width int, packLink string, emojis []entity.EmojiMeta) (int, error)
	}
}

func NewEmojiDM(
	cache interface {
		LoadAndDelete(key string) (value any, loaded bool)
	},
	userUC interface {
		UserByTelegramID(ctx context.Context, userID int64) (entity.User, error)
		CanGenerateEmojiPack(ctx context.Context, user entity.User, chatID int64) (bool, error)
	},
	processor interface {
		ParseArgs(arg string) (*entity.EmojiCommand, error)
		ExtractCommandArgs(msgText, msgCaption string) string
		SetupEmojiCommand(args *entity.EmojiCommand, user entity.User) *entity.EmojiCommand
		RegisterDirectory(dir string) error
		ProcessVideo(ctx context.Context, args *entity.EmojiCommand) ([]string, error)
	},
	// TODO заменить на usecase?
	repo interface {
		GetEmojiPackByPackLink(ctx context.Context, packLink string) (entity.EmojiPack, error)
		CreateNewEmojiPack(ctx context.Context, pack entity.EmojiPack) (entity.EmojiPack, error)
		UpdateEmojiCount(ctx context.Context, pack int64, emojiCount int) error
	},
	//stickerQueue interface {
	//	Acquire(packLink string) (bool, chan struct{})
	//	Release(packLink string)
	//},
	userBot interface {
		GetID() int64
		SendMessageWithEmojisToBot(ctx context.Context, chatID string, width int, packLink string, emojis []entity.EmojiMeta) (int, error)
	},
	progressManager interface {
		SendMessage(ctx context.Context, b *telego.Bot, chatID telego.ChatID, replyToID int, userID int64, status string) (*entity.ProgressMessage, error)
		UpdateMessage(ctx context.Context, bot *telego.Bot, chatID telego.ChatID, msgID int, status string) error
		DeleteMessage(ctx context.Context, bot *telego.Bot, chatID telego.ChatID, msgID int) error
		GetCancelChannel(cancelKey string) chan struct{}
	},
) *EmojiDM {
	queueModule := queue.New()
	uploaderModule := uploader.New()
	return &EmojiDM{
		cache:           cache,
		userUC:          userUC,
		processor:       processor,
		repo:            repo,
		uploader:        uploaderModule,
		progressManager: progressManager,
		queue:           queueModule,
		userBot:         userBot,
	}
}

func (h *EmojiDM) AddLogger(logger *slog.Logger) {
	h.uploader.AddLogger(logger)
	h.logger = logger.With(slog.String("handler", reflect.Indirect(reflect.ValueOf(h)).Type().PkgPath()))
}

func (h *EmojiDM) Command() string { return "emoji" }

func (h *EmojiDM) Handler() telegohandler.Handler {
	return func(bot *telego.Bot, update telego.Update) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		if update.Message == nil {
			return
		}

		lang := update.Message.From.LanguageCode

		botUser, ok := update.Context().Value(entity.BotSelfCtxKey).(*telego.User)
		if !ok || botUser == nil {
			h.logger.Error("bot info not found in context")
			return
		}

		// Получение пользователя
		user, err := h.userUC.UserByTelegramID(ctx, update.Message.From.ID)
		if err != nil {
			h.logger.Error("Failed to get user", slog.String("err", err.Error()))
			h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenError(lang))
			return
		}

		// Проверка возможности генерации
		canGenerate, err := h.userUC.CanGenerateEmojiPack(ctx, user, user.TelegramID)
		if err != nil {
			h.logger.Error("CanGenerate check failed", slog.String("err", err.Error()))
			h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.Error(lang))
			return
		}
		if !canGenerate {
			h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenLimitExceeded(lang))
			return
		}

		// Парсинг аргументов
		initialCommand := h.processor.ExtractCommandArgs(update.Message.Text, update.Message.Caption)
		emojiArgs, err := h.processor.ParseArgs(initialCommand)
		if err != nil {
			h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenError(lang))
			return
		}

		// Настройка команды
		h.processor.SetupEmojiCommand(emojiArgs, user)

		// Подготовка окружения
		if err := h.prepareWorkingEnvironment(ctx, bot, &update, emojiArgs); err != nil {
			h.handleDownloadError(bot, &update, err)
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
			h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.Error(lang))
			return
		}

		// Прогресс сообщение
		progress, err := h.progressManager.SendMessage(
			ctx, bot,
			update.Message.Chat.ChatID(),
			update.Message.MessageID,
			user.TelegramID,
			use_message.GL.EmojiGenProcessingStart(lang),
		)
		if err != nil {
			h.logger.Error("Progress message failed", slog.String("err", err.Error()))
		}
		defer func() {
			err := h.progressManager.DeleteMessage(ctx, bot, telego.ChatID{ID: update.Message.Chat.ID}, progress.MessageID)
			if err != nil {
				h.logger.Error("Failed to delete progress message", slog.String("err", err.Error()))
			}
		}()

		// Acquire processing lock for the emoji pack
		canProcess, waitCh := h.queue.Acquire(emojiArgs.PackLink)
		if !canProcess {
			// TODO тут можно обновлять статус прогресс-сообщения на 'очередь'
			h.logger.Debug("В ОЧЕРЕДИ", slog.String("pack_link", emojiArgs.PackLink))
			select {
			case <-ctx.Done():
				h.queue.Release(emojiArgs.PackLink)
				return
			case <-waitCh:
				h.logger.Debug("ОЧЕРЕДЬ ПРИШЛА, НАЧИНАЕТСЯ ОБРАБОТКА", slog.String("pack_link", emojiArgs.PackLink))
			}
		}
		defer h.queue.Release(emojiArgs.PackLink)
		// Обработка отмены
		cancelCh := h.progressManager.GetCancelChannel(progress.CancelKey)
		go func() {
			select {
			case <-cancelCh:
				cancel()
				if emojiArgs.NewSet {
					_ = bot.DeleteStickerSet(&telego.DeleteStickerSetParams{Name: emojiArgs.PackLink})
				}
			case <-ctx.Done():
				cancel()
			}
		}()

		// Основной цикл обработки
		var stickerSet *telego.StickerSet
		var emojiMetaRows [][]entity.EmojiMeta
		for {
			// Обработка видео
			h.progressManager.UpdateMessage(ctx, bot, telego.ChatID{ID: update.Message.Chat.ID}, progress.MessageID, use_message.GL.EmojiGenProcessingVideo(lang))
			files, err := h.processor.ProcessVideo(ctx, emojiArgs)
			if err != nil {
				h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenProcessingError(lang)+": "+err.Error())
				return
			}

			// Загрузка стикеров
			h.progressManager.UpdateMessage(ctx, bot, telego.ChatID{ID: update.Message.Chat.ID}, progress.MessageID, use_message.GL.EmojiGenUploadingEmojis(lang))
			stickerSet, emojiMetaRows, err = h.uploader.AddEmojis(ctx, bot, emojiArgs, files)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					return
				}
				if ue, ok := err.(*uploader.UploaderError); ok {
					switch ue.Code {
					case uploader.ErrCodeNoFiles:
						h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenNoFiles(lang))
					case uploader.ErrCodeExceedLimit:
						if total, ok1 := ue.Params["totalStickers"]; ok1 {
							if max, ok2 := ue.Params["max"]; ok2 {
								h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, fmt.Sprintf(use_message.GL.EmojiGenEmojiInPackLimitExceeded(lang), total, max))
							} else {
								h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenEmojiInPackLimitExceeded(lang))
							}
						} else {
							h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenEmojiInPackLimitExceeded(lang))
						}
					case uploader.ErrCodeUploadSticker:
						h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenUploadStickerError(lang))
					case uploader.ErrCodeOpenFile:
						h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenOpenFileError(lang))
					case uploader.ErrCodeUploadTransparentSticker:
						h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenUploadTransparentStickerError(lang))
					case uploader.ErrCodeCreateStickerSet:
						h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenCreateStickerSetError(lang))
					case uploader.ErrCodeAddStickers:
						h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenAddStickersError(lang))
					case uploader.ErrCodeGetStickerSet:
						h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenGetStickerSetError(lang))
					default:
						h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenUploadError(lang)+": "+ue.Error())
					}
				} else {
					h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, use_message.GL.EmojiGenUploadError(lang)+": "+err.Error())
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

		selectedEmojis := emoji_gen_utils.GenerateEmojiMessage(emojiMetaRows, stickerSet, emojiArgs)
		_, err = h.userBot.SendMessageWithEmojisToBot(ctx, strconv.FormatInt(botUser.ID, 10), emojiArgs.Width, emojiArgs.PackLink, selectedEmojis)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			h.logger.Error("Failed to send message with emojis",
				slog.String("err", err.Error()),
				slog.String("pack_link", emojiArgs.PackLink),
				slog.Int64("user_id", emojiArgs.TelegramUserID),
			)
		} else {
			sent := h.waitForEmojiMessageAndForwardIt(update.Context(), bot, update.Message.From.ID, emojiArgs.PackLink)
			if sent {
				return
			}
		}

		// Финализация
		h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID,
			use_message.GL.EmojiGenYourPack(lang, emojiArgs.PackLink))
	}
}

func (h *EmojiDM) waitForEmojiMessageAndForwardIt(ctx context.Context, bot *telego.Bot, userID int64, packLink string) bool {
	t := time.NewTicker(time.Millisecond * 500)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			// TODO подумать над тем чтобы перенсти это дело в redis
			msgIDRaw, ok := h.cache.LoadAndDelete("https://t.me/addemoji/" + packLink)
			if !ok {
				continue
			}

			msgID, ok := msgIDRaw.(int)
			if !ok {
				continue
			}

			h.forwardMessage(bot, h.userBot.GetID(), userID, msgID)
			return true
		case <-ctx.Done():
			return false
		}
	}

}

func (h *EmojiDM) Predicate() telegohandler.Predicate {
	return telegohandler.And(privateChatPredicate(), telegohandler.Or(telegohandler.CommandEqual(h.Command()), telegohandler.CaptionCommandEqual(h.Command())))

}

func privateChatPredicate() telegohandler.Predicate {
	return func(update telego.Update) bool {
		if update.Message == nil {
			return false
		}

		return update.Message.Chat.Type == "private"
	}
}

func (h *EmojiDM) handleDownloadError(bot *telego.Bot, update *telego.Update, err error) {
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
	h.sendMessage(bot, update.Message.Chat.ID, update.Message.MessageID, message)
}

// Добавленные методы для EmojiDM
func (h *EmojiDM) prepareWorkingEnvironment(ctx context.Context, bot *telego.Bot, update *telego.Update, args *entity.EmojiCommand) error {
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

func (h *EmojiDM) handleNewPack(ctx context.Context, botUser *telego.User, user entity.User, args *entity.EmojiCommand) (entity.EmojiPack, error) {
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

func (h *EmojiDM) handleExistingPack(ctx context.Context, args *entity.EmojiCommand) (entity.EmojiPack, error) {
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

func (h *EmojiDM) downloadFile(ctx context.Context, bot *telego.Bot, msg *telego.Message, args *entity.EmojiCommand) (string, error) {
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

func (h *EmojiDM) sendMessage(bot *telego.Bot, chatID int64, replyTo int, msgToSend string) {
	params := &telego.SendMessageParams{
		ChatID: telego.ChatID{ID: chatID},
		Text:   msgToSend,
	}

	if replyTo != 0 {
		params.ReplyParameters = &telego.ReplyParameters{
			MessageID: replyTo,
		}
	}

	_, err := bot.SendMessage(params)
	if err != nil {
		h.logger.Error("Failed to send message", slog.String("err", err.Error()), slog.Int64("user_id", chatID))
	}
}

func (h *EmojiDM) forwardMessage(bot *telego.Bot, userBotID int64, chatID int64, msgID int) {
	_, err := bot.ForwardMessage(&telego.ForwardMessageParams{
		FromChatID: telego.ChatID{ID: userBotID},
		ChatID:     telego.ChatID{ID: chatID},
		MessageID:  msgID,
	})
	if err != nil {
		h.logger.Error("Failed to forward message",
			slog.Int("msg_id", msgID),
			slog.Int64("to", chatID),
			slog.String("err", err.Error()))
	}
}
