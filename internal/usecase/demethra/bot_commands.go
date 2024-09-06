package demethra

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"

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
		msg := tgbotapi.NewMessage(chatId, fmt.Sprintf(`Что вершит судьбу человечества в этом мире? Некое незримое существо или закон, подобно Длани Господней парящей над миром? По крайне мере истинно то, что человек не властен даже над своей волей.`))

		//msg.ReplyMarkup = inlineKeyboard

		if _, err := b.Api.Send(msg); err != nil {
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

		if _, err := b.Api.Send(msg); err != nil {
			return err
		}

		return nil
	}
}

func (b *Bot) cmdDownload() CommandFunc {
	return func(ctx context.Context, update tgbotapi.Update) error {
		chatId := update.FromChat().ChatConfig().ChatID
		errMsg := tgbotapi.NewMessage(chatId, "неверная ссылка")

		argString := update.Message.CommandArguments()
		if argString == "" {
			_, err := b.Api.Send(errMsg)
			if err != nil {
				return fmt.Errorf("empty arguments; send err message: %w", err)
			}
			return fmt.Errorf("empty arguments")
		}

		args := strings.Split(argString, " ")

		_, err := url.Parse(args[0])
		if err != nil {
			_, err := b.Api.Send(errMsg)
			if err != nil {
				return fmt.Errorf("not url; : %w", err)
			}
			return fmt.Errorf("empty arguments")
		}

		return nil
	}
}

func (b *Bot) cmdDownloadInline() CommandFunc {
	return func(ctx context.Context, update tgbotapi.Update) error {
		data := strings.Split(update.CallbackQuery.Data, "?")
		if len(data) != 2 {
			return fmt.Errorf("invalid data format: %s", update.CallbackQuery.Data)
		}
		url := "https://soundcloud.com/" + data[1]

		song, err := b.repo.SongByUrl(ctx, url)
		if err != nil {
			return fmt.Errorf("get song by URL: %w", err)
		}

		forwardMsg := tgbotapi.NewForward(update.CallbackQuery.From.ID, song.SongTelegramMessageChatID, song.SongTelegramMessageID)

		_, err = b.Api.Send(forwardMsg)
		if err != nil {
			return fmt.Errorf("forward message: %w", err)
		}

		// просто логируем факт скачивания
		user, err := b.repo.GetUserByTelegramID(ctx, update.CallbackQuery.From.ID)
		if err != nil {
			b.logger.Error("get user by telegram id", slog.String("error", err.Error()), slog.Int64("telegram_id", update.CallbackQuery.From.ID), slog.String("method", "cmdDownloadInline"))
		} else {
			err := b.repo.LogSongDownload(ctx, song.ID, user.ID, entity.SongDownloadSourceBot)
			if err != nil {
				b.logger.Error("log song download", slog.String("error", err.Error()), slog.Int("song_id", song.ID), slog.Int("user_id", user.ID), slog.String("method", "cmdDownloadInline"))
			}

		}

		return nil
	}
}

func (b *Bot) cmdSwitchToggleForPostAutoDelete() CommandFunc {
	return func(ctx context.Context, update tgbotapi.Update) error {
		chatId := update.FromChat().ChatConfig().ChatID
		if !slices.Contains(b.adminIds, chatId) {
			return fmt.Errorf("несанкционированный доступ к /autodelete") // TODO возвращать ошибку или логировать ее для не-админов
		}

		var msg tgbotapi.MessageConfig

		b.mu.Lock()
		defer b.mu.Unlock()
		if b.DisableCommentSectionDelete {
			b.DisableCommentSectionDelete = false
			msg = tgbotapi.NewMessage(chatId, "Автоматическое удаление постов включено.")

		} else {
			b.DisableCommentSectionDelete = true
			msg = tgbotapi.NewMessage(chatId, "Автоматическое удаление постов отключено.")
		}
		_, err := b.Api.Send(msg)
		if err != nil {
			return fmt.Errorf("empty arguments; send err message: %w", err)
		}

		return nil
	}
}
