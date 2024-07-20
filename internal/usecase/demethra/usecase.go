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
	SongByUrl(ctx context.Context, url string) (entity.Song, error)
	CreateSong(ctx context.Context, song entity.Song) (entity.Song, error)
	CreatePlay(ctx context.Context, songID int) error
	CreateSongAndAddToPlayed(ctx context.Context, song entity.Song) (entity.Song, error)
	SongPlayed(ctx context.Context, songID int) (entity.SongPlay, error)
	RemoveSong(ctx context.Context, songID int) error
	SetCoverTelegramFileIDForSong(ctx context.Context, songID int, fileID string) error
	GetPlayedCountByID(ctx context.Context, songID int) (int, error)
	GetPlayedCountByURL(ctx context.Context, url string) (int, error)
	GetAllPlaysByURL(ctx context.Context, url string) ([]entity.SongPlay, error)
}
type soundcloudDownloader interface {
	DownloadTrackByURL(ctx context.Context, trackUrl string, info entity.TrackInfo) (string, error)
}

type Module struct {
	bot          *Bot
	cfg          config
	repo         repository
	eventRepo    eventRepository
	soundcloud   soundcloudDownloader
	logger       *slog.Logger
	prevTrack    entity.TrackInfo // Предыдущий трек
	currentTrack entity.TrackInfo // Текущий трек
}

// New конструктор ...
func New(cfg config, repo repository, eventRepo eventRepository, sc soundcloudDownloader) *Module {
	return &Module{
		cfg:        cfg,
		repo:       repo,
		eventRepo:  eventRepo,
		soundcloud: sc,
	}
}

// Init инициализатор ...
func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.logger = logger.With(slog.StringValue("🦕 " + m.cfg.GetBotName()))
	if err := m.cfg.Validate(); err != nil {
		return err
	}

	tgapi, err := tgbotapi.NewBotAPI(m.cfg.GetBotToken())
	if err != nil {
		return fmt.Errorf("new bot api: %w", err)
	}

	m.bot = newBot(ctx, m.repo, m.soundcloud, m.cfg.GetBotName(), tgapi, m.cfg.GetChatIDForLogs(), m.logger)
	go func() {
		m.bot.Run(ctx)
	}()

	return nil
}
