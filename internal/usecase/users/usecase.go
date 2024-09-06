package users

import (
	"arimadj-helper/internal/entity"
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
)

type cacheUC interface {
	GetListenersCount(ctx context.Context) (int64, error)
	GetAllCurrentListeners(ctx context.Context) ([]entity.ListenerCache, error)
}

type repository interface {
	CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)
}

type layoutCreator interface {
	GenerateAndSaveDefaultLayout(ctx context.Context, userID int, username string) (entity.UserLayout, error)
}

type Module struct {
	logger *slog.Logger
	ctx    context.Context

	cache  cacheUC
	repo   repository
	layout layoutCreator

	onlineUsersCount atomic.Int64
	mu               sync.Mutex
}

func New(cache cacheUC, repo repository, layout layoutCreator) *Module {
	return &Module{
		cache:  cache,
		repo:   repo,
		layout: layout,
	}
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.StringValue("🏓 USERS"))

	go m.updateOnlineUsersCountLoop()
	return nil
}
