package demethra

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"log/slog"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
	GetTelegramBotApiServer() string
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
	SongByID(ctx context.Context, id int) (entity.Song, error)
	CreateSong(ctx context.Context, song entity.Song) (entity.Song, error)
	SongPlayed(ctx context.Context, streamSlug string, songID int) (entity.SongPlay, error)
	RemoveSong(ctx context.Context, songID int) error
	SetCoverTelegramFileIDForSong(ctx context.Context, songID int, fileID string) error
	GetPlayedCountByID(ctx context.Context, songID int) (int, error)
	GetPlayedCountByURL(ctx context.Context, url string) (int, error)
	GetAllPlaysByURL(ctx context.Context, url string) ([]entity.SongPlay, error)
	LogSongDownload(ctx context.Context, songID int, userID int, source string) error

	CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID int64) (entity.User, error)
	GetUsersByTelegramID(ctx context.Context, telegramIDs []int64) ([]entity.User, error)

	SaveWebAppEvent(ctx context.Context, event entity.WebAppEvent) error
	SaveWebAppEvents(ctx context.Context, events []entity.WebAppEvent) error
	GetEventsByTelegramUserID(ctx context.Context, telegramUserID int64, since time.Time) ([]entity.WebAppEvent, error)

	GetUserByID(ctx context.Context, userID int) (entity.User, error)
	SaveUserSessionDuration(ctx context.Context, sessionDuration entity.UserSessionDuration) error
	BatchAddSongToUserSongHistory(ctx context.Context, histories []entity.UserToSongHistory) error
}
type soundcloudDownloader interface {
	DownloadTrackByURL(ctx context.Context, trackUrl string, info entity.TrackInfo) (string, error)
}

type downloader interface {
	DownloadByLink(ctx context.Context, url string, format string) (string, []byte, error)
	RemoveFile(ctx context.Context, fileName string) error
}

type usersUseCase interface {
	GetOnlineUsersCount() map[string]int64
	GetAllCurrentListeners(ctx context.Context) ([]entity.ListenerCache, error)
}

type Stream struct {
	Slug         string
	PrevTrack    entity.TrackInfo
	CurrentTrack entity.TrackInfo

	LastPlayed  entity.SongPlay
	LastUpdated time.Time

	mu sync.RWMutex
}

type Module struct {
	ctx  context.Context
	Bot  *Bot
	cfg  config
	repo repository
	//soundcloud soundcloudDownloader
	downloader downloader
	users      usersUseCase
	cache      cache
	logger     *slog.Logger

	//prevTrack    entity.TrackInfo // Предыдущий трек
	//currentTrack entity.TrackInfo // Текущий трек

	mu sync.RWMutex
	//lastPlayed entity.SongPlay // Последний проигранный трек

	batchEventUpdate chan entity.WebAppEvent

	streams map[string]*Stream
}

// New конструктор ...
func New(cfg config, repo repository, cache cache, downloader downloader, users usersUseCase) *Module {
	return &Module{
		cfg:   cfg,
		repo:  repo,
		cache: cache,
		//soundcloud: sc,
		downloader:       downloader,
		users:            users,
		mu:               sync.RWMutex{},
		streams:          make(map[string]*Stream),
		batchEventUpdate: make(chan entity.WebAppEvent, batchItemsCount),
	}
}

func (m *Module) AddStream(slug string) {
	s := &Stream{
		Slug: slug,
	}

	m.streams[slug] = s
}

// Init инициализатор ...
func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.String("module", "🦕 "+m.cfg.GetBotName()))
	err := m.cfg.Validate()
	if err != nil {
		return err
	}

	var tgapi *tgbotapi.BotAPI
	if m.cfg.GetTelegramBotApiServer() != "" {
		tgapi, err = tgbotapi.NewBotAPIWithAPIEndpoint(m.cfg.GetBotToken(), m.cfg.GetTelegramBotApiServer())
		if err != nil {
			return fmt.Errorf("new Bot api with custom server: %w", err)
		}

	} else {
		tgapi, err = tgbotapi.NewBotAPI(m.cfg.GetBotToken())
		if err != nil {
			return fmt.Errorf("new Bot api: %w", err)
		}
	}

	m.Bot = newBot(ctx, m.repo, m.downloader, m.users, m.cfg.GetBotName(), tgapi, m.cfg.GetChatIDForLogs(), m.cfg.GetElysiumFmID(), m.cfg.GetElysiumForumID(), m.cfg.GetElysiumFmCommentID(), m.cfg.GetTracksDbChannel(), m.cfg.GetCurrentTrackMessageID(), m.logger)
	go func() {
		m.Bot.Run(ctx)
	}()

	if len(m.streams) == 0 {
		m.streams["main"] = &Stream{
			Slug: "main",
		}
	}

	return nil
}
