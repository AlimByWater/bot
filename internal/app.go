package app

import (
	"arimadj-helper/internal/bot"
	"arimadj-helper/internal/domain"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
	"os/signal"
	"syscall"
)

var tokens = map[string]string{
	domain.BotArima:           "7447550770:AAHaO6tRmqNtb53fD9cIXPJVjYi1mHN3i_0",
	domain.BotElysiumDemethra: "7445477091:AAGOqZ_0_5vTkhHRNfK2iHWgk4ejM8UkL_8",
}

type App struct {
	ctx  context.Context
	stop context.CancelFunc
	bots map[string]*bot.Bot
}

func (a *App) init() error {

	bots := make(map[string]*bot.Bot)

	for name, token := range tokens {
		tgapi, err := tgbotapi.NewBotAPI(token)
		if err != nil {
			return fmt.Errorf("new bot api: %w", err)
		}

		b := bot.New(name, tgapi)

		switch name {
		case domain.BotArima:
			b.RegisterCommand(arimaCmdStart())
		case domain.BotElysiumDemethra:
			b.RegisterCommand(elysiumCmdStart())
		}
		bots[name] = b

	}

	a.bots = bots
	a.ctx, a.stop = signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	return nil
}

func (a *App) Run() {
	logrus.Info("application running")

	defer a.shutdown()

	err := a.init()
	if err != nil {
		logrus.Fatalln("init", err.Error())
	}

	for name, b := range a.bots {
		go func(name string, b *bot.Bot) {
			err := b.Run(a.ctx)
			if err != nil {
				logrus.Error(name, err.Error())
			}
		}(name, b)
	}

	<-a.ctx.Done()
}

func (a *App) shutdown() {
	logrus.Info("🦩Application shutdown")

	a.stop()
}

func elysiumCmdStart() func(b *bot.Bot) (string, bot.CommandFunc) {
	return func(b *bot.Bot) (string, bot.CommandFunc) {
		return "start", func(ctx context.Context, update tgbotapi.Update) error {
			//inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			//	tgbotapi.NewInlineKeyboardRow(
			//		tgbotapi.NewInlineKeyboardButtonData(`⁉️`, "/info"),
			//	),
			//)

			chatId := update.FromChat().ChatConfig().ChatID
			msg := tgbotapi.NewMessage(chatId, fmt.Sprintf(`Что вершит судьбу человечества в этом мире? Некое незримое существо или закон, подобно Длани Господней парящей над миром? По крайне мере истинно то, что человек не властен даже над своей волей.`))

			//msg.ReplyMarkup = inlineKeyboard

			if _, err := b.Api.Send(msg); err != nil {
				return err
			}

			return nil
		}
	}

}

func arimaCmdStart() func(b *bot.Bot) (string, bot.CommandFunc) {
	return func(b *bot.Bot) (string, bot.CommandFunc) {
		return "start", func(ctx context.Context, update tgbotapi.Update) error {
			//inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
			//	tgbotapi.NewInlineKeyboardRow(
			//		tgbotapi.NewInlineKeyboardButtonData(`⁉️`, "/info"),
			//	),
			//)

			chatId := update.FromChat().ChatConfig().ChatID
			msg := tgbotapi.NewMessage(chatId, fmt.Sprintf(`я не Арима ДЖ, но все ему передам`))

			//msg.ReplyMarkup = inlineKeyboard

			if _, err := b.Api.Send(msg); err != nil {
				return err
			}

			return nil
		}
	}
}
