package users

import (
	"context"
	"elysium/internal/entity"
	"log/slog"
	"sync"
	"sync/atomic"
)

type cacheUC interface {
	GetListenersCount(ctx context.Context) (map[string]int64, error)
	GetAllCurrentListeners(ctx context.Context) ([]entity.ListenerCache, error)
	GetUserByTelegramIDCache(ctx context.Context, telegramID int64) (entity.User, error)
	SaveOrUpdateUserCache(ctx context.Context, user entity.User) error
}

type repository interface {
	CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID int64) (entity.User, error)
	SetUserToBotActive(ctx context.Context, userID int, botID int64) error
}
type Module struct {
	logger *slog.Logger
	ctx    context.Context

	cache cacheUC
	repo  repository

	onlineUsersCount   atomic.Int64
	mu                 sync.RWMutex
	streamsOnlineCount map[string]int64
}

func New(cache cacheUC, repo repository) *Module {
	return &Module{
		cache: cache,
		repo:  repo,
	}
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.String("module", "🏓 USERS"))

	go m.updateOnlineUsersCountLoop()
	return nil
}
