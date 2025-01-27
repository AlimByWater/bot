package vip_bot

import (
	"context"
	"elysium/internal/entity"
	"errors"
	"fmt"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"strconv"
	"strings"
)

func (d *DBot) HandleEmojiCommandForDM(ctx context.Context, update *models.Update) {
	user, err := d.userUC.UserByTelegramID(ctx, update.Message.From.ID)
	if err != nil {
		d.logger.Error("Failed to get permissions", slog.String("err", err.Error()))
		d.sendMessageByBot(ctx, update.Message.Chat.ID, update.Message.ID, "Возникла внутреняя ошибка. Попробуйте позже", nil)
		return
	}

	{
		// Проверяем на возможность создать пак
		can, err := d.userUC.CanGenerateEmojiPack(ctx, user)
		if err != nil {
			d.logger.Error("Failed to get permissions", slog.String("err", err.Error()))
			d.sendMessageByBot(ctx, update.Message.Chat.ID, update.Message.ID, "Возникла внутреняя ошибка. Попробуйте позже", nil)
			return
		}

		if !can {
			// TODO тут надо проверит на возможность создать пак. Тут можем отправлять сообщение о пополнении токенов и т.д.
		}
	}

	// Extract command arguments
	initialCommand := d.processor.ExtractCommandArgs(update.Message.Text, update.Message.Caption)
	emojiArgs, err := d.processor.ParseArgs(initialCommand)
	if err != nil {
		d.logger.Error("Invalid arguments", slog.String("err", err.Error()))
		d.sendMessageByBot(ctx, update.Message.Chat.ID, update.Message.ID, err.Error(), nil)
		return
	}

	// Setup command defaults and working environment
	d.processor.SetupEmojiCommand(emojiArgs, user)

	// Create working directory and download file
	if err := d.prepareWorkingEnvironment(ctx, update, emojiArgs); err != nil {
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
		d.sendMessageByBot(ctx, update.Message.Chat.ID, update.Message.ID, message, nil)
		return
	}

	var emojiPack entity.EmojiPack
	if strings.Contains(emojiArgs.PackLink, d.tgbotApi.Self.UserName) {
		emojiPack, err = d.handleExistingPack(ctx, emojiArgs)
	} else {
		emojiPack, err = d.handleNewPack(ctx, user, d.b.ID, emojiArgs)
	}
	if err != nil {
		d.logger.Error("Failed to setup pack details", slog.String("err", err.Error()))
		d.sendMessageByBot(ctx, update.Message.Chat.ID, update.Message.ID, "пак с подобной ссылкой не найден", nil)
		return
	}

	// Create working directory and download file
	if err := d.prepareWorkingEnvironment(ctx, update, emojiArgs); err != nil {
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
		d.sendMessageByBot(ctx, update.Message.Chat.ID, update.Message.ID, message, nil)
		return
	}

	// Создаем контекст с отменой
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	defer cancel()

	// Отправляем сообщение с прогрессом и кнопкой отмены
	progress, err := d.sendProgressMessage(ctx, update.Message.Chat.ID, update.Message.ID, update.Message.From.ID, "⏳ Начинаем создание эмодзи-пака...")
	if err != nil {
		d.logger.Error("Failed to send initial progress message",
			slog.String("err", err.Error()),
			slog.Int64("telegram_user_id", emojiArgs.TelegramUserID))
	}
	defer d.deleteProgressMessage(ctx, update.Message.Chat.ID, progress.MessageID)

	// Запускаем горутину для отслеживания отмены
	cancelCh := d.progressManager.GetCancelChannel(progress.CancelKey)
	go func() {
		select {
		case <-cancelCh:
			cancel() // Отменяем контекст
			// Освобождаем очередь
			d.stickerQueue.Release(emojiArgs.PackLink)
			// Удаляем только новый пак, если он уже создан
			if emojiArgs.PackLink != "" && emojiArgs.NewSet {
				_, _ = d.b.BotApi.DeleteStickerSet(ctx, &bot.DeleteStickerSetParams{
					Name: emojiArgs.PackLink,
				})
			}
		case <-ctx.Done():
			return
		}
	}()

	var stickerSet *models.StickerSet
	var emojiMetaRows [][]entity.EmojiMeta

	for {
		// Обновляем статус: начало обработки видео
		d.updateProgressMessage(ctx, update.Message.Chat.ID, progress.MessageID, "🎬 Обрабатываем видео...")

		// Обрабатываем видео
		createdFiles, err := d.processor.ProcessVideo(ctx, emojiArgs)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			slog.LogAttrs(ctx, slog.LevelError, "Ошибка при обработке видео", emojiArgs.ToSlogAttributes(slog.String("err", err.Error()))...)
			d.sendMessageByBot(ctx, update.Message.Chat.ID, update.Message.ID, fmt.Sprintf("Ошибка при обработке видео: %s", err.Error()), nil)
			return
		}

		// Обновляем статус: создание стикеров
		d.updateProgressMessage(ctx, update.Message.Chat.ID, progress.MessageID, "✨ Создаем эмодзи...")

		// Создаем набор стикеров
		stickerSet, emojiMetaRows, err = d.uploader.AddEmojis(ctx, d.b.BotApi, emojiArgs, createdFiles)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}

			if strings.Contains(err.Error(), "PEER_ID_INVALID") || strings.Contains(err.Error(), "user not found") || strings.Contains(err.Error(), "bot was blocked by the user") {
				d.SendInitMessage(update.Message.Chat.ID, update.Message.ID)
				return
			}

			if strings.Contains(err.Error(), "STICKER_VIDEO_BIG") {
				_ = d.updateProgressMessage(ctx, update.Message.Chat.ID, progress.MessageID, "🔄 Оптимизируем размер видео...")
				emojiArgs.QualityValue++
				continue
			}

			// ошибка может означать что пользователь превысил внутренний лимит паков и ему нужно подождать какое то время
			if errors.Is(err, entity.ErrEmojiPacksLimitExceeded) {
				//d.sendErrorMessage(ctx, update.Message.Chat.ID, update.Message.ID, update.Message.MessageThreadID, err.Error())
				d.sendMessageByBot(ctx, update.Message.Chat.ID, update.Message.ID, err.Error(), nil)
				return
			}

			// ... остальная обработка ошибок ...
			d.sendMessageByBot(ctx, update.Message.Chat.ID, update.Message.ID, fmt.Sprintf("%s", err.Error()), nil)
			return
		}

		break
	}

	// Обновляем количество эмодзи в базе данных
	if err := d.repo.UpdateEmojiCount(ctx, emojiPack.ID, len(stickerSet.Stickers)); err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		d.logger.Error("Failed to update emoji count",
			slog.String("err", err.Error()),
			slog.String("pack_link", emojiArgs.PackLink),
			slog.Int64("user_id", emojiArgs.TelegramUserID))
	}

	if d.vip {
		selectedEmojis := d.processor.GenerateEmojiMessage(emojiMetaRows, stickerSet, emojiArgs)
		_, err := d.userBot.SendMessageWithEmojisToBot(ctx, strconv.FormatInt(d.tgbotApi.Self.ID, 10), emojiArgs.Width, emojiArgs.PackLink, selectedEmojis)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			d.logger.Error("Failed to send message with emojis",
				slog.String("err", err.Error()),
				slog.String("pack_link", emojiArgs.PackLink),
				slog.Int64("user_id", emojiArgs.TelegramUserID))
		} else {
			sent := d.waitForEmojiMessageAndForwardIt(ctx, update.Message.From.ID, emojiArgs.PackLink)
			if sent {
				return
			}
		}
	}

	d.sendMessageByBot(ctx, update.Message.Chat.ID, update.Message.ID, fmt.Sprintf("Ваш пак\n%s", "https://t.me/addemoji/"+emojiArgs.PackLink), nil)

}
func (d *DBot) onEmojiSelect(ctx context.Context, b *bot.Bot, mes models.MaybeInaccessibleMessage, data []byte) {
	d.sendInfoMessage(ctx, mes.Message.Chat.ID, 0)
}
