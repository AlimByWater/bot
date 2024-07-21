package demethra

import (
	"arimadj-helper/internal/application/logger"
	"arimadj-helper/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/bogem/id3v2"
	"log/slog"
	"math/rand"
	"os"
	"slices"
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
// bot bot bot forum -1002224939217

const (
	ElysiumFmID           int64 = -1001934236726
	ElysiumChatID         int64 = -1002224939217
	ElysiumFmCommentID    int64 = -1002044294733
	CurrentTrackMessageID       = 271
	TracksDbChannel             = -1002243776940
)

var (
	BotRepliesVariants = []string{"я им передам.", "ты был услышан.", "хорошо, я им передам", "это все что ты хотел сказать?"}
	defaultKeyboard    = tgbotapi.NewReplyKeyboard()
)

type Bot struct {
	Api                         *tgbotapi.BotAPI
	sc                          soundcloudDownloader
	logger                      *slog.Logger
	cmdViews                    map[string]CommandFunc
	name                        string
	chatIDForLogs               int64
	repo                        repository
	ctx                         context.Context
	mu                          sync.RWMutex
	WhiteListPostsID            []int // рудемент, но пусть пока что будет
	adminIds                    []int64
	DisableCommentSectionDelete bool
}

func newBot(ctx context.Context, repo repository, sc soundcloudDownloader, name string, api *tgbotapi.BotAPI, chatIDForLogs int64, logger *slog.Logger) *Bot {
	b := &Bot{
		name:             name,
		repo:             repo,
		sc:               sc,
		Api:              api,
		chatIDForLogs:    chatIDForLogs,
		logger:           logger,
		ctx:              ctx,
		mu:               sync.RWMutex{},
		adminIds:         []int64{251636949, 548414066, 5534121833},
		WhiteListPostsID: []int{509, 391}, // TODO перенести их в базу
	}

	b.registerCommands()

	return b
}

func (b *Bot) registerCommands() {
	b.registerCommand("start", b.cmdStart())
	b.registerCommand("download", b.cmdDownloadInline())
	b.registerCommand("autodelete", b.cmdSwitchToggleForPostAutoDelete())
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
			b.logger.Error("create user", slog.StringValue(err.Error()), slog.Int64("user_id", tgUser.ID))
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

		// если update.Message.Chat.ID == ElysiumChatID то проверить не является ли сообщение ссылкой на саундклауд, если является - скачать трек и прислать его в чат
		if update.Message.Chat.ID == ElysiumChatID && !update.Message.IsCommand() {
			fmt.Println(update.Message.Text)
			sent, err := b.checkSoundCloudUrlAndSend(ctx, update, attributes)
			if err != nil {
				b.logger.LogAttrs(ctx, slog.LevelError, "check soundcloud url and send", logger.AppendErrorToLogs(attributes, err)...)
			}

			if sent {
				return
			}
		}

		// TODO update раскомментить
		//если сообщение пришло в чате комментов - пропускаем #исключение
		if update.Message.Chat.ID == ElysiumFmCommentID {
			// если сообщение в авторстве элизиум_фм - удаляем его
			b.mu.RLock()
			defer b.mu.RUnlock()
			if update.Message.ForwardOrigin != nil && update.Message.ForwardOrigin.Chat.ID == ElysiumFmID {
				// тут проверяем нет ли его в исключениях
				if b.DisableCommentSectionDelete {
					return
				}
				if !slices.Contains(b.WhiteListPostsID, update.Message.ForwardOrigin.MessageID) {
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
			sent, err := b.checkSoundCloudUrlAndSend(ctx, update, attributes)
			if err != nil {
				b.logger.LogAttrs(ctx, slog.LevelError, "check soundcloud url and send", logger.AppendErrorToLogs(attributes, err)...)
			}

			if sent {
				return
			}

			// отправляем сообщение в чат
			err = b.sendToChat(update)
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
		b.logger.LogAttrs(ctx, slog.LevelError, "failed to execute view", logger.AppendErrorToLogs(attributes, err)...)

		//if _, err := b.Api.Send(tgbotapi.NewMessage(update.FromChat().ChatConfig().ChatID, "Internal error")); err != nil {
		//	b.logger.LogAttrs(ctx, slog.LevelError, "failed to send error message", logger.AppendErrorToLogs(attributes, err)...)
		//}
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

func (b *Bot) checkSoundCloudUrlAndSend(ctx context.Context, update tgbotapi.Update, attrs []slog.Attr) (bool, error) {
	if update.Message.Text != "" {
		if strings.Contains(update.Message.Text, "soundcloud.com") {
			text := strings.Split(update.Message.Text, " ")
			for _, v := range text {
				if strings.Contains(v, "soundcloud.com") {
					song, err := b.repo.SongByUrl(ctx, v)
					if err != nil && !errors.Is(err, sql.ErrNoRows) {
						b.logger.LogAttrs(ctx, slog.LevelWarn, "get song by URL from chat", logger.AppendErrorToLogs(attrs, err)...)
					}

					if song.ID != 0 {
						err = b.forwardSongToChat(update.Message.Chat.ID, song)
						if err != nil {
							b.logger.LogAttrs(ctx, slog.LevelWarn, "forward track to chat from chat", logger.AppendErrorToLogs(attrs, err)...)
						}
					}

					songPath, err := b.sc.DownloadTrackByURL(ctx, v, entity.TrackInfo{})
					if err != nil {
						return false, fmt.Errorf("download track by url: %w", err)
					}
					defer func(songPath string) {
						err := os.Remove(songPath)
						if err != nil {
							b.logger.LogAttrs(ctx, slog.LevelError, "remove song", logger.AppendErrorToLogs(attrs, err)...)
						}
					}(songPath)

					err = b.sengSongToChat(update.Message.Chat.ID, songPath)
					if err != nil {
						return false, fmt.Errorf("send track to chat: %w", err)
					}

					return true, nil
				}
			}
			// отправить в чат
			// удалить сообщение
			// отправить сообщение об успешной загрузке
			// если трек не удалось скачать - отправить сообщение об ошибке
			// если трек не удалось отправить - отправить сообщение об ошибке
			// если сообщение не удалось удалить - отправить сообщение об ошибке
			// если сообщение об успешной загрузке не удалось отправить - отправить сообщение об ошибке
		} else {
			return false, nil
		}
	} else {
		return false, nil
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

func (b *Bot) sengSongToChat(chatID int64, songPath string) error {
	file, err := os.Open(songPath)
	if err != nil {
		return fmt.Errorf("open song: %w", err)
	}

	tag, err := id3v2.Open(songPath, id3v2.Options{Parse: true})
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
					ChatID: chatID,
				},
			},
			File: tgbotapi.FileReader{
				Reader: file,
			},
		},
		Caption:   `||[elysium fm](t.me/elysium_fm)||`,
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
