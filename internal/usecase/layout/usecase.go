package layout

import (
	"arimadj-helper/internal/entity"
	"context"
	"log/slog"
	"sync"
)

type cacheUC interface {
}

type repoUC interface {
	LayoutByUserID(ctx context.Context, userID int) (entity.UserLayout, error)
	LayoutByID(ctx context.Context, layoutID string) (entity.UserLayout, error)
	UpdateLayout(ctx context.Context, layout entity.UserLayout) error
	LogLayoutChange(ctx context.Context, change entity.LayoutChange) error
}

type Module struct {
	logger *slog.Logger
	ctx    context.Context

	cache cacheUC
	repo  repoUC
	mu    sync.Mutex
}

func New(cache cacheUC, repo repoUC) *Module {
	return &Module{
		cache: cache,
		repo:  repo,
	}
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.StringValue("📱 LAYOUT"))

	return nil
}
