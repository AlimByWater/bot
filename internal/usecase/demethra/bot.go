package demethra

import (
	"bytes"
	"context"
	"database/sql"
	"elysium/internal/application/logger"
	"elysium/internal/entity"
	"errors"
	"fmt"
	"github.com/bogem/id3v2"
	"log/slog"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Структурированный комментарии -1002044294733,
// Структурированный -1001934236726
// ElysiumFM -1002129034021
// ElysiumChat -1002124956071
// ElysiumFmComment  -1002164548613
// Bot Bot Bot forum -1002224939217

var (
	BotRepliesVariants = []string{"я им передам.", "ты был услышан.", "хорошо, я им передам", "это все что ты хотел сказать?"}
	defaultKeyboard    = tgbotapi.NewReplyKeyboard()
)

type Bot struct {
	Api *tgbotapi.BotAPI
	//sc       soundcloudDownloader
	downloader downloader
	logger     *slog.Logger
	cmdViews   map[string]CommandFunc
	name       string
	users      usersUseCase

	chatIDForLogs         int64
	mainChannelID         int64
	forumID               int64
	commentChatID         int64
	tracksDbChannelID     int64
	currentTrackMessageID int // TODO поменять на слайс который будет подтягиватьcя с базы данных

	repo                        repository
	ctx                         context.Context
	mu                          sync.RWMutex
	adminIds                    []int64
	DisableCommentSectionDelete bool
}

func newBot(ctx context.Context, repo repository, downloader downloader, users usersUseCase, name string, api *tgbotapi.BotAPI, chatIDForLogs, mainChannelID, forumID, commentChatID, tracksDbChannelID int64, currentTrackMessageID int, logger *slog.Logger) *Bot {
	b := &Bot{
		name: name,
		repo: repo,
		//sc:                    sc,
		downloader:            downloader,
		users:                 users,
		Api:                   api,
		chatIDForLogs:         chatIDForLogs,
		mainChannelID:         mainChannelID,
		forumID:               forumID,
		commentChatID:         commentChatID,
		tracksDbChannelID:     tracksDbChannelID,
		currentTrackMessageID: currentTrackMessageID,
		logger:                logger,
		ctx:                   ctx,
		mu:                    sync.RWMutex{},
		adminIds:              []int64{251636949, 548414066, 5534121833},
	}

	b.registerCommands()

	return b
}

func (b *Bot) registerCommands() {
	b.registerCommand("start", b.cmdStart())
	b.registerCommand("download", b.cmdDownloadInline())
	b.registerCommand("autodelete", b.cmdSwitchToggleForPostAutoDelete())
	b.registerCommand("online", b.cmdCheckCurrentOnline())
	//b.registerCommand("/download", b.cmdDownload())
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

	b.logger.Info("Bot started")

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
			b.logger.Error("panic recovered: ", slog.Any("err", p))
		}
	}()

	if update.SentFrom() != nil {
		tgUser := update.SentFrom()
		user := entity.User{
			TelegramID:       tgUser.ID,
			TelegramUsername: tgUser.UserName,
			Firstname:        tgUser.FirstName,
			DateCreate:       time.Now(),
		}

		_, err := b.repo.CreateOrUpdateUser(ctx, user)
		if err != nil {
			b.logger.Error("create user", slog.String("err", err.Error()), slog.Int64("user_id", tgUser.ID))
		}
	}

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

		// если сообщение пришло с форума, то проверить не является ли сообщение ссылкой на саундклауд, если является - скачать трек и прислать его в чат
		if !update.Message.IsCommand() {
			sent, err := b.checkDownloadUrlAndSend(ctx, update, attributes)
			if err != nil {
				b.logger.LogAttrs(ctx, slog.LevelError, "check soundcloud url and send", logger.AppendErrorToLogs(attributes, err)...)
			}

			if sent {
				return
			}
		}

		// TODO update раскомментить
		//если сообщение пришло в чате комментов - пропускаем #исключение
		if update.Message.Chat.ID == b.commentChatID {
			// если сообщение в авторстве элизиум_фм - удаляем его
			b.mu.RLock()
			defer b.mu.RUnlock()
			if update.Message.ForwardOrigin != nil && update.Message.ForwardOrigin.Chat.ID == b.mainChannelID {
				// тут проверяем нет ли его в исключениях
				if b.DisableCommentSectionDelete {
					return
				}

				resp, err := b.Api.Request(tgbotapi.NewDeleteMessage(b.commentChatID, update.Message.MessageID))
				if err != nil {
					b.logger.LogAttrs(ctx, slog.LevelError, "delete message", logger.AppendErrorToLogs(attributes, err)...)
					return
				}
				b.logger.LogAttrs(ctx, slog.LevelDebug, "delete message ", logger.AppendToLogs(attributes, slog.Any("resp", resp))...)

				return
			}

			return
		}

		// отвечаем на сообщения присланные боту
		if !update.Message.IsCommand() && update.CallbackQuery == nil && update.Message.Chat.Type != entity.ChatTypeSuperGroup {
			// отправляем сообщение в чат
			err := b.logToChat(update)
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
		_, err := b.Api.Request(callback)
		if err != nil {
			b.logger.LogAttrs(ctx, slog.LevelError, "request callback", logger.AppendErrorToLogs(attributes, err)...)
			return
		}

		data := update.CallbackQuery.Data
		dataSlice := strings.Split(data, "?")

		cmd = dataSlice[0]
	} else if update.Message.IsCommand() {
		cmd = update.Message.Command()
	}

	cmdView, ok := b.cmdViews[cmd]
	if !ok {
		return
	}

	view = cmdView

	if err := view(ctx, update); err != nil {
		b.logger.LogAttrs(ctx, slog.LevelError, "failed to execute view", logger.AppendToLogs(logger.AppendErrorToLogs(attributes, err), slog.String("username", update.SentFrom().UserName))...)

		if _, err := b.Api.Send(tgbotapi.NewMessage(b.chatIDForLogs,
			fmt.Sprintf("Internal error: %s;\nusername: %s;\ntelegram_user_id: %d", err.Error(), update.SentFrom().UserName, update.SentFrom().ID))); err != nil {
			b.logger.LogAttrs(ctx, slog.LevelError, "failed to send error message", logger.AppendErrorToLogs(attributes, err)...)
		}
	}
}

func (b *Bot) logToChat(u tgbotapi.Update) error {
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

var allowedDomains = map[string]bool{
	"youtube.com":      true,
	"youtu.be":         true,
	"bandcamp.com":     true,
	"soundcloud.com":   true,
	"open.spotify.com": true,
	"spotify.com":      true,
	"deezer.com":       true,
	"deezer.page.link": true,
	"music.yandex.ru":  true,
	"yandex.ru":        true,
}

func validateLink(link string) error {
	for domain := range allowedDomains {
		if strings.Contains(link, domain) {
			return nil
		}
	}
	return entity.ErrInvalidDownloadLink
}

func (b *Bot) checkDownloadUrlAndSend(ctx context.Context, update tgbotapi.Update, attrs []slog.Attr) (bool, error) {
	if update.Message.Text == "" {
		return false, nil
	}

	for domain, _ := range allowedDomains {
		if !strings.Contains(update.Message.Text, domain) {
			continue
		}
		tokens := strings.Split(update.Message.Text, " ")
		for _, token := range tokens {
			if !strings.Contains(token, domain) {
				continue
			}

			url, err := url.Parse(token)
			if err != nil {
				continue
			}

			if url.RequestURI() == "" {
				continue
			}

			song, err := b.repo.SongByUrl(ctx, url.String())
			if err != nil && !errors.Is(err, sql.ErrNoRows) {
				b.logger.LogAttrs(ctx, slog.LevelWarn, "get song by URL from chat", logger.AppendErrorToLogs(attrs, err)...)
			}

			if song.ID != 0 {
				err = b.forwardSongToChat(update.Message.Chat.ID, song)
				if err != nil {
					b.logger.LogAttrs(ctx, slog.LevelWarn, "forward track to chat from chat", logger.AppendErrorToLogs(attrs, err)...)
				}

				return true, nil
			}

			// TODO логировать в базу факт скачивания трека

			var replyMessage tgbotapi.Message
			if update.Message.Chat.Type == entity.ChatTypePrivate {
				chatId := update.FromChat().ChatConfig().ChatID
				//rand.Seed(time.Now().Unix())
				msg := tgbotapi.NewMessage(chatId, `Скачиваю\. ||Надеюсь у меня получится\.||`)
				msg.ReplyParameters.MessageID = update.Message.MessageID
				msg.ParseMode = "MarkdownV2"

				//msg.ReplyMarkup = inlineKeyboard
				if replyMessage, err = b.Api.Send(msg); err != nil {
					b.logger.LogAttrs(ctx, slog.LevelError, "send reply to user to audio request", logger.AppendErrorToLogs(attrs, err)...)
				}
			}

			fileName, songData, err := b.downloader.DownloadByLink(ctx, url.String(), "mp3")
			if err != nil {
				msg := tgbotapi.NewMessage(update.FromChat().ChatConfig().ChatID, "Не получилось скачать(")
				msg.ReplyParameters.MessageID = update.Message.MessageID

				//msg.ReplyMarkup = inlineKeyboard
				if _, err2 := b.Api.Send(msg); err2 != nil {
					b.logger.LogAttrs(ctx, slog.LevelError, "send reply to user to audio request", logger.AppendErrorToLogs(attrs, err)...)
				}

				return false, fmt.Errorf("download track by url: %w", err)
			}
			defer func(songPath string) {
				err := b.downloader.RemoveFile(ctx, fileName)
				if err != nil {
					b.logger.LogAttrs(ctx, slog.LevelError, "remove song", logger.AppendErrorToLogs(attrs, err)...)
				}
			}(fileName)

			err = b.sengSongToChat(update, songData)
			if err != nil {
				msg := tgbotapi.NewMessage(update.FromChat().ChatConfig().ChatID, "Не получилось отправить(")
				msg.ReplyParameters.MessageID = update.Message.MessageID

				//msg.ReplyMarkup = inlineKeyboard
				if _, err2 := b.Api.Send(msg); err2 != nil {
					b.logger.LogAttrs(ctx, slog.LevelError, "send reply to user to audio request", logger.AppendErrorToLogs(attrs, err)...)
				}
				return false, fmt.Errorf("send track to chat: %w", err)
			}

			if replyMessage.MessageID != 0 {
				delMessageReq := tgbotapi.NewDeleteMessage(replyMessage.Chat.ID, replyMessage.MessageID)
				_, err = b.Api.Send(delMessageReq)
				if err != nil {
					b.logger.LogAttrs(ctx, slog.LevelWarn, "delete reply message for audio request", logger.AppendErrorToLogs(attrs, err)...)
				}
			}

			return true, nil

		}
	}

	return true, nil
}

func (b *Bot) forwardSongToChat(chatID int64, song entity.Song) error {
	forwardMsg := tgbotapi.NewForward(chatID, song.SongTelegramMessageChatID, song.SongTelegramMessageID)

	_, err := b.Api.Send(forwardMsg)
	if err != nil {
		return fmt.Errorf("forward message: %w", err)
	}

	return nil
}

func (b *Bot) sengSongToChat(u tgbotapi.Update, song []byte) error {
	songReader := bytes.NewReader(song)

	tag, err := id3v2.ParseReader(songReader, id3v2.Options{Parse: true})
	if err != nil {
		return err
	}
	defer tag.Close()

	artist := tag.Artist()
	title := tag.Title()

	// ************* ОТПРАВИТЬ ТРЕК В ГРУППУ *************** //
	audio := tgbotapi.AudioConfig{
		BaseFile: tgbotapi.BaseFile{
			BaseChat: tgbotapi.BaseChat{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: u.Message.Chat.ID,
				},
				ReplyParameters: tgbotapi.ReplyParameters{
					MessageID: u.Message.MessageID,
				},
			},
			File: tgbotapi.FileBytes{
				Bytes: song,
			},
		},
		Caption:   `[элизиум \[ラジオ\]](t.me/elysium_fm)`,
		ParseMode: "MarkdownV2",
		Title:     title,
		Performer: artist,
	}
	_, err = b.Api.Send(audio)
	if err != nil {
		return fmt.Errorf("send audio: %w", err)
	}

	return nil
}
