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

type Module struct {
	logger *slog.Logger
	ctx    context.Context

	cache cacheUC

	onlineUsersCount atomic.Int64
	mu               sync.Mutex
}

func New(cache cacheUC) *Module {
	return &Module{
		cache: cache,
	}
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.StringValue("🏓 USERS"))

	go m.updateOnlineUsersCountLoop()
	return nil
}
