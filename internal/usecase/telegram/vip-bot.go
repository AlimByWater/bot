package telegram

import (
	"context"
	"elysium/internal/entity"
	"elysium/internal/usecase/emoji-gen/queue"
	"elysium/internal/usecase/telegram/emoji-gen/processing"
	progress "elysium/internal/usecase/telegram/emoji-gen/progress-manager"
	"elysium/internal/usecase/telegram/emoji-gen/uploader"
	"elysium/internal/usecase/telegram/emoji-gen/vip_bot"
	"elysium/internal/usecase/telegram/httpclient"
	"fmt"
	botapi "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"golang.org/x/time/rate"
	"log/slog"
	"strings"
	"time"
)

func (m *Manager) initVipBot(b *entity.Bot, logger *slog.Logger) (*botapi.Bot, error) {
	rl := rate.NewLimiter(rate.Every(1*time.Second), 100)
	c := httpclient.NewClient(rl, logger)

	fmt.Println(b.Token)
	api, err := botapi.New(b.Token,
		botapi.WithHTTPClient(time.Minute, c),
		botapi.WithMiddlewares(m.saveUserMiddleware(b)),
		botapi.WithDefaultHandler(m.vipHandlers))
	if err != nil {
		return nil, fmt.Errorf("error creating botapi: %w", err)
	}
	b.BotApi = api

	processor := processing.NewProcessingModule(logger)

	queueModule := queue.New()
	uploaderModule := uploader.New(queueModule, logger)
	progressManager := progress.NewManager()

	vipBot := vip_bot.NewBot(b, c, m.userBot, processor, m.userUC, m.repo, uploaderModule, progressManager, queueModule)
	err = vipBot.Init(m.ctx, logger)
	if err != nil {
		return nil, fmt.Errorf("error initializing vip bot: %w", err)
	}

	m.vip = vipBot
	return api, nil
}

func (m *Manager) vipHandlers(ctx context.Context, b *botapi.Bot, update *models.Update) {
	m.wg.Add(1)
	defer m.wg.Done()

	// Добавляем обработку callback-запросов
	if update.CallbackQuery != nil {
		if strings.HasPrefix(update.CallbackQuery.Data, "cancel_") {
			m.vip.HandleCancelCallback(ctx, update)
			return
		}
	}

	if update.Message == nil {
		return
	}

	if update.Message.Chat.Type == models.ChatTypePrivate {
		if strings.HasPrefix(update.Message.Text, "start") || strings.HasPrefix(update.Message.Caption, "start ") {
			m.vip.HandleStartCommand(ctx, b, update)
			return
		}

		if strings.HasPrefix(update.Message.Text, "/info") {
			m.vip.HandleInfoCommand(ctx, update)
			return
		}

		if strings.HasPrefix(update.Message.Text, "/emoji") ||
			strings.HasPrefix(update.Message.Caption, "/emoji ") {
			m.vip.HandleEmojiCommandForDM(ctx, update)
			return
		}

		if update.Message.From.ID == m.userBot.GetID() {
			m.vip.TrackEmojiMessage(update)
		}
	}
}
