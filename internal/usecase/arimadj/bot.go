package arimadj

import (
	"context"
	"elysium/internal/application/logger"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"math/rand"
	"time"
)

const ArimaDJ int64 = -1002132552731

var (
	BotRepliesVariants = []string{"я ему передам.", "хорошо, я ему передам"}
	defaultKeyboard    = tgbotapi.NewReplyKeyboard()
)

type Bot struct {
	Api           *tgbotapi.BotAPI
	logger        *slog.Logger
	cmdViews      map[string]CommandFunc
	name          string
	chatIDForLogs int64
}

func newBot(name string, api *tgbotapi.BotAPI, chatIDForLogs int64, logger *slog.Logger) *Bot {
	b := &Bot{
		name:          name,
		Api:           api,
		chatIDForLogs: chatIDForLogs,
		logger:        logger,
	}

	b.registerCommands()

	return b
}

func (b *Bot) registerCommands() {
	b.registerCommand("start", b.cmdStart())
	//b.registerCommand("/info", b.cmdInfo())
	//b.registerCommand("⁉️Инфа", b.cmdInfo())
	//b.registerCommand("/calendar", b.cmdCalendar())
}

func (b *Bot) registerCommand(cmd string, view CommandFunc) {
	if b.cmdViews == nil {
		b.cmdViews = make(map[string]CommandFunc)
	}

	b.cmdViews[cmd] = view
}

func (b *Bot) Run(ctx context.Context) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.Api.GetUpdatesChan(u)

	b.logger.Info("bot started")

	for {
		select {
		case update := <-updates:
			updateCtx, updateCancel := context.WithTimeout(context.Background(), 5*time.Minute)
			b.handleUpdate(updateCtx, update)
			updateCancel()
		case <-ctx.Done():
			if ctx.Err() != nil {
				b.logger.Error("ctx error", slog.StringValue(ctx.Err().Error()))
			}

		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			b.logger.Error("panic recovered: ", slog.AnyValue(p))
		}
	}()

	// для логов
	// если не проверять пустое ли сообщение то можно словить панику
	var attributes []slog.Attr
	if update.Message != nil {
		attributes = []slog.Attr{
			slog.String("user", update.Message.From.FirstName),
			slog.String("username", update.Message.From.UserName),
			slog.Int64("user_id", update.Message.From.ID),
			slog.Int64("chat_id", update.Message.Chat.ID),
		}
	}

	// отвечаем на сообщения присланные боту
	if update.Message != nil && !update.Message.IsCommand() && update.CallbackQuery == nil {
		err := b.sendToChat(update)
		if err != nil {
			b.logger.LogAttrs(ctx, slog.LevelError, "send to chat", logger.AppendErrorToLogs(attributes, err)...)
			return
		}

		chatId := update.FromChat().ChatConfig().ChatID
		//rand.Seed(time.Now().Unix())
		msg := tgbotapi.NewMessage(chatId, BotRepliesVariants[rand.Intn(len(BotRepliesVariants))])

		//msg.ReplyMarkup = inlineKeyboard
		if _, err := b.Api.Send(msg); err != nil {
			b.logger.LogAttrs(ctx, slog.LevelError, "send reply to user", logger.AppendErrorToLogs(attributes, err)...)
			return
		}
		return
	}

	if (update.Message == nil || !update.Message.IsCommand()) && update.CallbackQuery == nil {
		return
	}

	var view CommandFunc
	var cmd string

	if update.CallbackQuery != nil {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
		if _, err := b.Api.Request(callback); err != nil {
			b.logger.LogAttrs(ctx, slog.LevelError, "request callback", logger.AppendErrorToLogs(attributes, err)...)
			return
		}

		cmd = update.CallbackQuery.Data
	} else if update.Message.IsCommand() {
		cmd = update.Message.Command()
	}

	cmdView, ok := b.cmdViews[cmd]
	if !ok {
		return
	}

	view = cmdView

	if err := view(ctx, update); err != nil {
		b.logger.LogAttrs(ctx, slog.LevelError, "failed to execute view", logger.AppendErrorToLogs(attributes, err)...)

		if _, err := b.Api.Send(tgbotapi.NewMessage(update.FromChat().ChatConfig().ChatID, "Internal error")); err != nil {
			b.logger.LogAttrs(ctx, slog.LevelError, "failed to send error message", logger.AppendErrorToLogs(attributes, err)...)
		}
	}
}

func (b *Bot) sendToChat(u tgbotapi.Update) error {
	text := fmt.Sprintf(`
%s
	
id: %d
имя: %s @%s
%s
	`, u.Message.Text, u.Message.From.ID, u.Message.From.FirstName, u.Message.From.UserName, time.Unix(int64(u.Message.Date), 0).String())

	msg := tgbotapi.NewMessage(b.chatIDForLogs, text)
	_, err := b.Api.Send(msg)
	if err != nil {
		return fmt.Errorf("forward details msg: %w", err)
	}

	if u.Message != nil && u.Message.MessageID != 0 {
		fwd := tgbotapi.NewForward(b.chatIDForLogs, u.Message.Chat.ID, u.Message.MessageID)
		_, err = b.Api.Send(fwd)
		if err != nil {
			return fmt.Errorf("forward msg: %w", err)
		}
	}

	return nil
}
