package vip_bot

import (
	"context"
	"elysium/internal/entity"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"log/slog"
	"strings"
	"time"
)

func (d *DBot) SendInitMessage(chatID int64, msgID int) {
	inlineKeyboard := tgbotapi.NewInlineKeyboardButtonURL("/start", fmt.Sprintf("t.me/%s?start=start", d.tgbotApi.Self.UserName))
	row := tgbotapi.NewInlineKeyboardRow(inlineKeyboard)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(row)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("Чтобы бот мог создать пак \nнажмите кнопку ниже\n↓↓↓↓↓↓↓↓"))
	msg.ReplyMarkup = keyboard
	msg.ParseMode = "MarkdownV2"
	msg.ReplyParameters = tgbotapi.ReplyParameters{
		MessageID: msgID,
		ChatID:    chatID,
	}

	_, err2 := d.tgbotApi.Send(msg)
	if err2 != nil {
		slog.Error("Failed to send message with emojis", slog.Int64("user_id", chatID), slog.String("err2", err2.Error()))
	}
}

func (d *DBot) sendErrorMessage(ctx context.Context, chatID int64, replyTo int, threadID int, errToSend string) {
	params := bot.SendMessageParams{
		ChatID: chatID,
		Text:   fmt.Sprintf("%s", errToSend),
	}

	if replyTo != 0 {
		params.ReplyParameters = &models.ReplyParameters{
			MessageID: replyTo,
			ChatID:    chatID,
		}
	}

	chat := fmt.Sprintf("%d", chatID)
	if threadID != 0 {
		chat = fmt.Sprintf("%s_%d", chat, threadID)
	}

	err := d.userBot.SendMessage(ctx, chat, params)
	if err != nil {
		if errors.Is(err, context.Canceled) || strings.Contains(err.Error(), "context canceled") {
			return
		}
		d.logger.Error("Failed to send error message", slog.String("err", err.Error()))
	}
}

func (d *DBot) editMessageByBot(ctx context.Context, chatID int64, msgId int, msgToSend string, keyboard models.ReplyMarkup) {
	params := &bot.EditMessageTextParams{
		ChatID:    chatID,
		MessageID: msgId,
		Text:      fmt.Sprintf("%s", msgToSend),
	}

	if keyboard != nil {
		params.ReplyMarkup = keyboard
	}

	_, err := d.b.BotApi.EditMessageText(ctx, params)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		d.logger.Error("Failed to edit message by bot", slog.String("err", err.Error()), slog.Int64("user_id", chatID), slog.Int("msg_id", msgId))
	}
}

func (d *DBot) forwardMessage(ctx context.Context, fromChatID, toChatID int64, msgID int) {
	params := &bot.ForwardMessageParams{
		FromChatID: fromChatID,
		ChatID:     toChatID,
		MessageID:  msgID,
	}

	_, err := d.b.BotApi.ForwardMessage(ctx, params)
	if err != nil && !errors.Is(err, context.Canceled) {
		d.logger.Error("Failed to forward message", slog.Int("msg_id", msgID), slog.Int64("to", toChatID), slog.String("err", err.Error()))
	}
}

func (d *DBot) sendMessageByBot(ctx context.Context, chatID int64, replyTo int, msgToSend string, keyboard models.ReplyMarkup) {
	params := &bot.SendMessageParams{
		ChatID: chatID,
		Text:   fmt.Sprintf("%s", msgToSend),
	}

	if keyboard != nil {
		params.ReplyMarkup = keyboard
	}

	if replyTo != 0 {
		params.ReplyParameters = &models.ReplyParameters{
			MessageID: replyTo,
			ChatID:    chatID,
		}
	}

	_, err := d.b.BotApi.SendMessage(ctx, params)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		d.logger.Error("Failed to send error message", slog.String("err", err.Error()), slog.Int64("user_id", chatID))
	}
	return
}

func (d *DBot) handleExistingPack(ctx context.Context, args *entity.EmojiCommand) (entity.EmojiPack, error) {
	args.NewSet = false
	if strings.Contains(args.PackLink, "t.me/addemoji/") {
		splited := strings.Split(args.PackLink, ".me/addemoji/")
		args.PackLink = strings.TrimSpace(splited[len(splited)-1])
	}

	pack, err := d.repo.GetEmojiPackByPackLink(ctx, args.PackLink)
	if err != nil {
		return entity.EmojiPack{}, err
	}
	args.SetTitle = ""
	return pack, nil
}

func (d *DBot) handleNewPack(ctx context.Context, user entity.User, internalBotID int64, args *entity.EmojiCommand) (entity.EmojiPack, error) {
	args.NewSet = true
	packName := fmt.Sprintf("%s%d_by_%s", "dt", time.Now().Unix(), d.tgbotApi.Self.UserName)
	if len(packName) > entity.TelegramPackLinkAndNameLength {
		args.PackLink = args.PackLink[:len(packName)-entity.TelegramPackLinkAndNameLength]
		packName = fmt.Sprintf("%s_%s", args.PackLink, d.tgbotApi.Self.UserName)
	}
	args.PackLink = packName

	emojiPack := entity.EmojiPack{
		CreatorTelegramID: user.TelegramID,
		PackTitle:         args.SetTitle,
		//TelegramFileID:    args.File.FileID,
		PackLink:       args.PackLink,
		InitialCommand: args.InitialCommand,
		Bot:            entity.Bot{ID: internalBotID},
		EmojiCount:     0,
	}

	return d.repo.CreateNewEmojiPack(ctx, emojiPack)
}

func (d *DBot) sendProgressMessage(ctx context.Context, chatID int64, replyToID int, userID int64, status string) (*entity.ProgressMessage, error) {
	return d.progressManager.SendMessage(ctx, d.b.BotApi, chatID, replyToID, userID, status)
}

func (d *DBot) deleteProgressMessage(_ context.Context, chatID int64, msgID int) {
	// ctx может быть уже отменен, поэтому создаем новый
	err := d.progressManager.DeleteMessage(context.Background(), d.b.BotApi, chatID, msgID)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return
		}
		d.logger.Error("Error deleting progress message", slog.String("chatID", fmt.Sprintf("%d", chatID)), slog.Int("msgID", msgID), slog.String("error", err.Error()))
	}
}

func (d *DBot) updateProgressMessage(ctx context.Context, chatID int64, msgID int, status string) error {
	return d.progressManager.UpdateMessage(ctx, d.b.BotApi, chatID, msgID, status)
}
