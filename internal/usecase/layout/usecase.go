package layout

import (
	"context"
	"log/slog"
	"sync"
)

type cacheUC interface {
}

type Module struct {
	logger *slog.Logger
	ctx    context.Context

	cache cacheUC
	mu    sync.Mutex
}

func New(cache cacheUC) *Module {
	return &Module{
		cache: cache,
	}
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.StringValue("📱 LAYOUT"))

	return nil
}
