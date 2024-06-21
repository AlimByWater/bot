package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type CommandFunc func(ctx context.Context, update tgbotapi.Update) error

func (b *Bot) cmdStart() CommandFunc {
	return func(ctx context.Context, update tgbotapi.Update) error {
		//inlineKeyboard := tgbotapi.NewInlineKeyboardMarkup(
		//	tgbotapi.NewInlineKeyboardRow(
		//		tgbotapi.NewInlineKeyboardButtonData(`⁉️`, "/info"),
		//	),
		//)

		chatId := update.FromChat().ChatConfig().ChatID
		msg := tgbotapi.NewMessage(chatId, fmt.Sprintf(`я не Арима ДЖ, но все ему передам`))

		//msg.ReplyMarkup = inlineKeyboard

		if _, err := b.api.Send(msg); err != nil {
			return err
		}

		return nil
	}
}

func (b *Bot) cmdInfo() CommandFunc {
	return func(ctx context.Context, update tgbotapi.Update) error {
		chatId := update.FromChat().ChatConfig().ChatID

		msg := tgbotapi.NewMessage(chatId, fmt.Sprintf(`
Правила пользования

1. //....
2. //......
3. //.......
4. //.........
`))

		msg.ReplyMarkup = defaultKeyboard

		if _, err := b.api.Send(msg); err != nil {
			return err
		}

		return nil
	}
}
