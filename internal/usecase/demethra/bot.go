package demethra

import (
	"arimadj-helper/internal/application/logger"
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"slices"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Структурированный комментарии -1002044294733,
// ElysiumFM -1002129034021
// ElysiumChat -1002124956071
// ElysiumFmComment  -1002164548613
// bot bot bot forum -1002224939217

const (
	ElysiumFmID           int64 = -1002129034021
	ElysiumChatID         int64 = -1002124956071
	ElysiumFmCommentID    int64 = -1002164548613
	CurrentTrackMessageID       = 13832
	ArimaDJ               int64 = -1002132552731
)

var (
	WhiteListPostsId   = []int{509, 391}
	BotRepliesVariants = []string{"я им передам.", "ты был услышан.", "хорошо, я им передам", "это все что ты хотел сказать?"}
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
			//if ctx.Err() != nil {
			//	b.logger.Error("ctx error", slog.StringValue(ctx.Err().Error()))
			//}

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

		// если сообщение пришло в чате комментов - пропускаем #исключение
		if update.Message.Chat.ID == ElysiumFmCommentID {
			// если сообщение в авторстве элизиум_фм - удаляем его
			if update.Message.ForwardOrigin != nil && update.Message.ForwardOrigin.Chat.ID == ElysiumFmID {
				// тут проверяем нет ли его в исключениях
				if !slices.Contains(WhiteListPostsId, update.Message.ForwardOrigin.MessageID) {
					resp, err := b.Api.Request(tgbotapi.NewDeleteMessage(ElysiumFmCommentID, update.Message.MessageID))
					if err != nil {
						b.logger.LogAttrs(ctx, slog.LevelError, "delete message", logger.AppendErrorToLogs(attributes, err)...)
						return
					}
					b.logger.LogAttrs(ctx, slog.LevelDebug, "delete message ", logger.AppendToLogs(attributes, slog.Any("resp", resp))...)
				}

				return
			}

			return
		}

		// отвечаем на сообщения присланные боту
		if !update.Message.IsCommand() && update.CallbackQuery == nil && update.Message.Chat.Type != entity.ChatTypeSuperGroup {

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

	if u.Message != nil && u.Message.Text != "" {
		fwd := tgbotapi.NewForward(b.chatIDForLogs, u.Message.Chat.ID, u.Message.MessageID)
		_, err = b.Api.Send(fwd)
		if err != nil {
			return fmt.Errorf("forward msg: %w", err)
		}
	}

	return nil
}

func (b *Bot) updateCurrentTrackMessage(current, prev entity.TrackInfo) error {
	previewUrl := current.TrackLink
	current.Format()
	prev.Format()
	visual := formatEscapeChars(fmt.Sprintf(`0:35 ━❍──────── -%s
             *↻     ⊲  Ⅱ  ⊳     ↺*
VOLUME: ▁▂▃▄▅▆▇ 100%%`, current.Duration))

	b.logger.Debug("song update", slog.Any("current", current), slog.Any("prev", prev))

	msg := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			BaseChatMessage: tgbotapi.BaseChatMessage{
				MessageID:  CurrentTrackMessageID,
				ChatConfig: tgbotapi.ChatConfig{ChatID: ElysiumChatID},
			},
		},
		LinkPreviewOptions: tgbotapi.LinkPreviewOptions{
			ShowAboveText:    true,
			PreferLargeMedia: true,
			URL:              previewUrl,
			IsDisabled:       false,
		},
		ParseMode: "MarkdownV2",
		Text: fmt.Sprintf(`
*[%s \- %s](%s)*
%s

||Предыдущий: [%s \- %s](%s)||
`,
			current.ArtistName, current.TrackTitle, current.TrackLink,
			visual,
			prev.ArtistName, prev.TrackTitle, prev.TrackLink),
		//		Text: fmt.Sprintf(`
		//Текущий трек: %s - %s
		//%s
		//
		//Предыдущий: %s - %s
		//%s
		//`,
		//			current.ArtistName, strings.Replace(current.TrackTitle, "Current track: ", "", 1), current.TrackLink,
		//			prev.ArtistName, strings.Replace(prev.TrackTitle, "Current track: ", "", 1), prev.TrackLink),
	}

	_, err := b.Api.Send(msg)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return nil
}

// '_', '*', '[', ']', '(', ')', '~', '`', '>', '#', '+', '-', '=', '|', '{', '}', '.', '!'
func formatEscapeChars(oldS string) string {
	s := oldS
	s = strings.ReplaceAll(s, `_`, `\_`)
	//s = strings.ReplaceAll(s, `*`, `\*`)
	s = strings.ReplaceAll(s, `[`, `\[`)
	s = strings.ReplaceAll(s, `]`, `\]`)
	s = strings.ReplaceAll(s, `(`, `\(`)
	s = strings.ReplaceAll(s, `)`, `\)`)
	s = strings.ReplaceAll(s, `~`, `\~`)
	//s = strings.ReplaceAll(s, "`", "\`")
	s = strings.ReplaceAll(s, `>`, `\>`)
	s = strings.ReplaceAll(s, `#`, `\#`)
	s = strings.ReplaceAll(s, `+`, `\+`)
	s = strings.ReplaceAll(s, `-`, `\-`)
	s = strings.ReplaceAll(s, `=`, `\=`)
	s = strings.ReplaceAll(s, `|`, `\|`)
	s = strings.ReplaceAll(s, `{`, `\{`)
	s = strings.ReplaceAll(s, `}`, `\}`)
	s = strings.ReplaceAll(s, `.`, `\.`)
	s = strings.ReplaceAll(s, `!`, `\!`)

	return s
}
