package demethra

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
)

type config interface {
	GetBotToken() string
	GetBotName() string
	GetChatIDForLogs() int64
	Validate() error
}

type repository interface {
	SongPlayed(info entity.TrackInfo) error
}

type Module struct {
	bot          *Bot
	cfg          config
	repo         repository
	logger       *slog.Logger
	prevTrack    entity.TrackInfo // Предыдущий трек
	currentTrack entity.TrackInfo // Текущий трек
}

// New конструктор ...
func New(cfg config, repo repository) *Module {
	return &Module{
		cfg:  cfg,
		repo: repo,
	}
}

// Init инициализатор ...
func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	//m.ctx = ctx
	m.logger = logger.With(slog.StringValue("🦕 " + m.cfg.GetBotName()))
	if err := m.cfg.Validate(); err != nil {
		return err
	}

	tgapi, err := tgbotapi.NewBotAPI(m.cfg.GetBotToken())
	if err != nil {
		return fmt.Errorf("new bot api: %w", err)
	}

	m.bot = newBot(ctx, m.cfg.GetBotName(), tgapi, m.cfg.GetChatIDForLogs(), m.logger)
	go func() {
		m.bot.Run(ctx)
	}()

	return nil
}
