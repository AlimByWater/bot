package telegram

import (
	"context"
	"elysium/internal/entity"
	"elysium/internal/usecase/telegram/emoji-gen/userbot"
	"elysium/internal/usecase/telegram/emoji-gen/vip_bot"
	"fmt"
	botapi "github.com/go-telegram/bot"
	"log/slog"
	"sync"
)

type config interface {
	GetUserBotAppID() int
	GetUserBotAppHash() string
	GetUserBotTgPhone() string
	IsLocal() bool
	GetSessionDir() string
	Validate() error
}

type repository interface {
	GetAllBots(ctx context.Context) ([]*entity.Bot, error)
	GetBotByID(ctx context.Context, botID int64) (*entity.Bot, error)

	CreateOrUpdateAccessHash(ctx context.Context, accessHash entity.AccessHash) error
	GetAccessHash(ctx context.Context, chatID string) (entity.AccessHash, error)
	GetAllAccessHashes(ctx context.Context) ([]entity.AccessHash, error)

	GetEmojiPackByPackLink(ctx context.Context, packLink string) (entity.EmojiPack, error)
	GetEmojiPacksByCreator(ctx context.Context, creator int64, deleted bool) ([]entity.EmojiPack, error)
	SetEmojiPackDeleted(ctx context.Context, packName string) error
	UnsetEmojiPackDeleted(ctx context.Context, packName string) error
	CreateNewEmojiPack(ctx context.Context, pack entity.EmojiPack) (entity.EmojiPack, error)
	UpdateEmojiCount(ctx context.Context, pack int64, emojiCount int) error
}

type UserUC interface {
	CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)
	UserByTelegramID(ctx context.Context, userID int64) (entity.User, error)
	SetUserToBotActive(ctx context.Context, userID int, botID int64) error
	CanGenerateEmojiPack(ctx context.Context, user entity.User) (bool, error)
}

type Manager struct {
	ctx    context.Context
	logger *slog.Logger
	cfg    config

	mu sync.RWMutex
	wg sync.WaitGroup // обработка обновлений

	repo   repository
	userUC UserUC

	allBots map[int64]*entity.Bot
	vip     *vip_bot.DBot
	//dripbot *drip_bot.DBot
	//threeDViewer threed_bot.TBot
	userBot *userbot.User
}

func NewManager(config config, repo repository, userUC UserUC) *Manager {
	return &Manager{
		allBots: make(map[int64]*entity.Bot),
		cfg:     config,
		repo:    repo,
		userUC:  userUC,
	}
}

func (m *Manager) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.String("module", "💻 telegram"))

	err := m.cfg.Validate()
	if err != nil {
		return fmt.Errorf("telegram config validate: %w", err)
	}

	err = m.initUserBot()
	if err != nil {
		return err
	}

	bots, err := m.repo.GetAllBots(m.ctx)
	if err != nil {
		return err
	}

	for _, b := range bots {
		if !m.cfg.IsLocal() && b.Test {
			m.logger.Info("Пропуск бота в тестовом режиме",
				slog.Int64("bot_id", b.ID),
				slog.String("bot_name", b.Name))
			continue
		}

		var api *botapi.Bot
		switch b.Purpose {
		case entity.ProjectEmojiGenVip:
			api, err = m.initVipBot(b, logger)
			m.logger.Info("emoji-gen-vip бот инициализирован", slog.Int64("bot_id", b.ID), slog.String("bot_name", b.Name))
		default:
			m.logger.Info("Неизвестное назначение бота",
				slog.Int64("bot_id", b.ID),
				slog.String("bot_name", b.Name),
				slog.String("purpose", b.Purpose))
			continue
		}

		if err != nil {
			return fmt.Errorf("failed to init bot: %w", err)
		}

		b.BotApi = api
		m.allBots[b.ID] = b
	}

	for _, b := range m.allBots {
		b.BotApi.Start(m.ctx)
	}

	return nil
}

func (m *Manager) initUserBot() error {
	m.userBot = userbot.NewBot(m.repo, m.cfg)
	err := m.userBot.Init(m.ctx, m.logger)
	if err != nil {
		m.logger.Error("Failed to init userbot", slog.String("error", err.Error()), slog.String("module", "💻 telegram"))
		return fmt.Errorf("failed to init userbot: %w", err)
	}

	return nil
}
