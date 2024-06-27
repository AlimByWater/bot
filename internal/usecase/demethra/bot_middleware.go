package demethra

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (b *Bot) adminsOnlyMiddleware(channelID int64, next CommandFunc) CommandFunc {
	return func(ctx context.Context, update tgbotapi.Update) error {
		admins, err := b.Api.GetChatAdministrators(
			tgbotapi.ChatAdministratorsConfig{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: channelID,
				},
			},
		)

		if err != nil {
			return err
		}

		for _, admin := range admins {
			if admin.User.ID == update.SentFrom().ID {
				return next(ctx, update)
			}
		}

		if _, err := b.Api.Send(tgbotapi.NewMessage(
			update.FromChat().ID,
			"У вас нет прав на выполнение этой команды.",
		)); err != nil {
			return err
		}

		return nil
	}
}
