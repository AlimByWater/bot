package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"time"
)

var defaultKeyboard = tgbotapi.NewReplyKeyboard()

type Bot struct {
	api      *tgbotapi.BotAPI
	cmdViews map[string]CommandFunc
}

func New(api *tgbotapi.BotAPI) *Bot {
	b := &Bot{
		api: api,
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

func (b *Bot) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	logrus.Info("bot started")

	for {
		select {
		case update := <-updates:
			updateCtx, updateCancel := context.WithTimeout(context.Background(), 5*time.Minute)
			b.handleUpdate(updateCtx, update)
			updateCancel()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			logrus.Error("panic recovered: ", p)
		}
	}()

	log := logrus.WithFields(logrus.Fields{
		"name": update.Message.From.FirstName, "username": update.Message.From.UserName, "user_id": update.Message.From.ID, "chat_id": update.Message.Chat.ID,
	})

	if update.Message != nil && !update.Message.IsCommand() {
		err := b.sendToChat(update)
		if err != nil {
			log.Error("send to chat", err)
			return
		}

		chatId := update.FromChat().ChatConfig().ChatID
		msg := tgbotapi.NewMessage(chatId, fmt.Sprintf(`я ему передам.`))

		//msg.ReplyMarkup = inlineKeyboard

		if _, err := b.api.Send(msg); err != nil {
			log.Error("send reply to user", err)
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
		if _, err := b.api.Request(callback); err != nil {
			panic(err)
		}
		//
		//// And finally, send a message containing the data received.
		//msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
		//if _, err := b.api.Send(msg); err != nil {
		//	panic(err)
		//}

		cmd = update.CallbackQuery.Data
	} else if update.Message.IsCommand() {
		cmd = update.Message.Command()
	}

	log.Info(cmd)

	cmdView, ok := b.cmdViews[cmd]
	if !ok {
		return
	}

	view = cmdView

	if err := view(ctx, update); err != nil {
		log.Error("failed to execute view", err)

		if _, err := b.api.Send(tgbotapi.NewMessage(update.FromChat().ChatConfig().ChatID, "Internal error")); err != nil {
			log.Error("failed to send error message", err)
		}
	}
}

func (b *Bot) sendToChat(u tgbotapi.Update) error {
	text := fmt.Sprintf(`
%s
	
имя: %s
ник и id: @%s %d
чат: %d
%s
	`, u.Message.Text, u.Message.From.FirstName, u.Message.From.UserName, u.Message.From.ID, u.Message.Chat.ID, time.Unix(int64(u.Message.Date), 0).String())

	msg := tgbotapi.NewMessage(-1002044294733, text)
	_, err := b.api.Send(msg)
	if err != nil {
		return fmt.Errorf("forward details msg: %w", err)
	}

	fwd := tgbotapi.NewForward(-1002044294733, u.Message.Chat.ID, u.Message.MessageID)
	_, err = b.api.Send(fwd)
	if err != nil {
		return fmt.Errorf("forward msg: %w", err)
	}

	return nil

}

//Комментарии = -1002044294733
