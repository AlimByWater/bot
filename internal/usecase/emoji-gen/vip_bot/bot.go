package vip_bot

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/ogen-go/ogen/http"
	"log/slog"
	"strings"
	"sync"
)

var (
	validchatIDs            = []string{"-1002400904088_3", "-1002491830452_3", "-1002002718381"}
	packDeletePrefixMessage = "pack_delete:"
)

type UserBot interface {
	SendMessageWithEmojis(ctx context.Context, chatID string, width int, packLink string, emojis []entity.EmojiMeta, replyTo int) error
	SendMessageWithEmojisToBot(ctx context.Context, chatID string, width int, packLink string, emojis []entity.EmojiMeta) (int, error)
	SendMessage(ctx context.Context, chatID string, msg bot.SendMessageParams) error
	GetID() int64
}

type Queuer interface {
	Acquire(packLink string) (bool, chan struct{})
	Release(packLink string)
	Clear()
}

// ProgressManager отвечает за отслеживание прогресса обработки
type ProgressManager interface {
	SendMessage(ctx context.Context, b *bot.Bot, chatID int64, replyToID int, userID int64, status string) (*entity.ProgressMessage, error)
	UpdateMessage(ctx context.Context, b *bot.Bot, chatID int64, msgID int, status string) error
	DeleteMessage(ctx context.Context, b *bot.Bot, chatID int64, msgID int) error
	GetCancelChannel(cancelKey string) chan struct{}
	Cancel(cancelKey string)
}

type Processinger interface {
	// Methods from ArgsParser
	ParseArgs(arg string) (*entity.EmojiCommand, error)
	ExtractCommandArgs(msgText, msgCaption string) string
	SetupEmojiCommand(args *entity.EmojiCommand, user entity.User) *entity.EmojiCommand

	// Methods from DirectoryManager
	RegisterDirectory(dir string) error
	CheckAndRemoveOldDirectories()

	// Methods from VideoProcessor
	ProcessVideo(ctx context.Context, args *entity.EmojiCommand) ([]string, error)

	// Methods from MessageGenerator
	GenerateEmojiMessage(emojiMetaRows [][]entity.EmojiMeta, stickerSet *models.StickerSet, emojiArgs *entity.EmojiCommand) []entity.EmojiMeta
}

type Repository interface {
	GetEmojiPackByPackLink(ctx context.Context, packLink string) (entity.EmojiPack, error)
	GetEmojiPacksByCreator(ctx context.Context, creator int64, deleted bool) ([]entity.EmojiPack, error)
	SetEmojiPackDeleted(ctx context.Context, packName string) error
	UnsetEmojiPackDeleted(ctx context.Context, packName string) error
	CreateNewEmojiPack(ctx context.Context, pack entity.EmojiPack) (entity.EmojiPack, error)
	UpdateEmojiCount(ctx context.Context, pack int64, emojiCount int) error
}

type Uploader interface {
	AddEmojis(ctx context.Context, b *bot.Bot, args *entity.EmojiCommand, emojiFiles []string) (*models.StickerSet, [][]entity.EmojiMeta, error)
}

type Userer interface {
	UserByTelegramID(ctx context.Context, userID int64) (entity.User, error)
	CanGenerateEmojiPack(ctx context.Context, user entity.User) (bool, error)
}

type DBot struct {
	b *entity.Bot

	logger           *slog.Logger
	httpclient       http.Client
	tgbotApi         *tgbotapi.BotAPI
	messagesToDelete sync.Map
	wg               sync.WaitGroup
	userBot          UserBot
	stickerQueue     Queuer
	progressManager  ProgressManager
	userUC           Userer
	processor        Processinger
	repo             Repository
	uploader         Uploader
	vip              bool
}

func NewBot(b *entity.Bot, client http.Client, userBot UserBot, processor Processinger, userUC Userer, repository Repository, uploader Uploader, progressManager ProgressManager, queue Queuer) *DBot {
	dbot := &DBot{
		b:               b,
		httpclient:      client,
		userBot:         userBot,
		stickerQueue:    queue,
		progressManager: progressManager,
		userUC:          userUC,
		processor:       processor,
		repo:            repository,
		uploader:        uploader,
		vip:             true,
	}

	return dbot
}

func (d *DBot) Init(_ context.Context, logger *slog.Logger) error {
	d.logger = logger.With(slog.String("module", "VIP BOT"))

	tgbotApi, err := tgbotapi.NewBotAPIWithClient(d.b.Token, tgbotapi.APIEndpoint, d.httpclient)
	if err != nil {
		return fmt.Errorf("error creating tgbotapi: %w", err)
	}

	tgbotApi.StopReceivingUpdates()

	d.tgbotApi = tgbotApi
	return nil
}

func (d *DBot) handler(ctx context.Context, b *bot.Bot, update *models.Update) {
	d.wg.Add(1)
	defer d.wg.Done()

	// Добавляем обработку callback-запросов
	if update.CallbackQuery != nil {
		if strings.HasPrefix(update.CallbackQuery.Data, "cancel_") {
			d.HandleCancelCallback(ctx, update)
			return
		}
	}

	if update.Message == nil {
		return
	}

	if update.Message.Chat.Type == models.ChatTypePrivate {
		if strings.HasPrefix(update.Message.Text, "start") || strings.HasPrefix(update.Message.Caption, "start ") {
			d.HandleStartCommand(ctx, b, update)
			return
		}

		if strings.HasPrefix(update.Message.Text, "/info") {
			d.HandleInfoCommand(ctx, update)
			return
		}

		if strings.HasPrefix(update.Message.Text, "/emoji") ||
			strings.HasPrefix(update.Message.Caption, "/emoji ") {
			d.HandleEmojiCommandForDM(ctx, update)
			return
		}

		if update.Message.From.ID == d.userBot.GetID() {
			d.TrackEmojiMessage(update)
		}
	}

}
