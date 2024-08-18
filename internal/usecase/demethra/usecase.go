package demethra

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
	"sync"
	"time"
)

const (
	batchItemsCount = 100
)

type config interface {
	GetBotToken() string
	GetBotName() string
	GetChatIDForLogs() int64
	GetListenerIdleTimeoutInMinutes() int
	GetElysiumFmID() int64
	GetElysiumForumID() int64
	GetElysiumFmCommentID() int64
	GetTracksDbChannel() int64
	GetCurrentTrackMessageID() int
	Validate() error
}

type cache interface {
	SaveOrUpdateListener(ctx context.Context, c entity.ListenerCache) error
	GetListenerByTelegramID(ctx context.Context, telegramID int64) (entity.ListenerCache, error)
	GetAllCurrentListeners(ctx context.Context) ([]entity.ListenerCache, error)
	RemoveListenerTelegramID(ctx context.Context, telegramID int64) error
}

type repository interface {
	SongByUrl(ctx context.Context, url string) (entity.Song, error)
	CreateSong(ctx context.Context, song entity.Song) (entity.Song, error)
	CreateSongAndAddToPlayed(ctx context.Context, song entity.Song) (entity.Song, error)
	SongPlayed(ctx context.Context, songID int) (entity.SongPlay, error)
	RemoveSong(ctx context.Context, songID int) error
	SetCoverTelegramFileIDForSong(ctx context.Context, songID int, fileID string) error
	GetPlayedCountByID(ctx context.Context, songID int) (int, error)
	GetPlayedCountByURL(ctx context.Context, url string) (int, error)
	GetAllPlaysByURL(ctx context.Context, url string) ([]entity.SongPlay, error)

	CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)

	SaveWebAppEvent(ctx context.Context, event entity.WebAppEvent) error
	SaveWebAppEvents(ctx context.Context, events []entity.WebAppEvent) error
	GetEventsByTelegramUserID(ctx context.Context, telegramUserID int64, since time.Time) ([]entity.WebAppEvent, error)

	BatchAddSongToUserSongHistory(ctx context.Context, histories []entity.UserToSongHistory) error
}
type soundcloudDownloader interface {
	DownloadTrackByURL(ctx context.Context, trackUrl string, info entity.TrackInfo) (string, error)
}

type Module struct {
	ctx        context.Context
	bot        *Bot
	cfg        config
	repo       repository
	soundcloud soundcloudDownloader
	cache      cache
	logger     *slog.Logger

	prevTrack    entity.TrackInfo // Предыдущий трек
	currentTrack entity.TrackInfo // Текущий трек

	mu         sync.RWMutex
	lastPlayed entity.SongPlay // Последний проигранный трек

	batchEventUpdate chan entity.WebAppEvent
}

// New конструктор ...
func New(cfg config, repo repository, cache cache, sc soundcloudDownloader) *Module {
	return &Module{
		cfg:        cfg,
		repo:       repo,
		cache:      cache,
		soundcloud: sc,

		mu:               sync.RWMutex{},
		batchEventUpdate: make(chan entity.WebAppEvent, batchItemsCount),
	}
}

// Init инициализатор ...
func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.StringValue("🦕 " + m.cfg.GetBotName()))
	if err := m.cfg.Validate(); err != nil {
		return err
	}

	tgapi, err := tgbotapi.NewBotAPI(m.cfg.GetBotToken())
	if err != nil {
		return fmt.Errorf("new bot api: %w", err)
	}

	m.bot = newBot(ctx, m.repo, m.soundcloud, m.cfg.GetBotName(), tgapi, m.cfg.GetChatIDForLogs(), m.cfg.GetElysiumFmID(), m.cfg.GetElysiumForumID(), m.cfg.GetElysiumFmCommentID(), m.cfg.GetTracksDbChannel(), m.cfg.GetCurrentTrackMessageID(), m.logger)
	go func() {
		m.bot.Run(ctx)
	}()

	return nil
}
