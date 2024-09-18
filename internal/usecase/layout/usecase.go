package layout

import (
	"context"
	"elysium/internal/entity"
	"log/slog"
	"sync"
)

// cacheUC интерфейс для работы с кэшем
type cacheUC interface {
	GetLayout(ctx context.Context, layoutID int) (entity.UserLayout, error)
	SaveOrUpdateLayout(ctx context.Context, layout entity.UserLayout) error
	DeleteLayout(ctx context.Context, layoutID int) error
	GetLayoutByName(ctx context.Context, layoutName string) (entity.UserLayout, error)
	GetLayoutByUserID(ctx context.Context, userID int) (entity.UserLayout, error)
}

// repoUC интерфейс для работы с репозиторием макетов
type repoUC interface {
	LayoutByUserID(ctx context.Context, userID int) (entity.UserLayout, error)
	LayoutByID(ctx context.Context, layoutID int) (entity.UserLayout, error)
	UpdateLayoutFull(ctx context.Context, layout entity.UserLayout) error
	LogLayoutChange(ctx context.Context, change entity.LayoutChange) error
	IsLayoutOwner(ctx context.Context, layoutID int, userID int) (bool, error)
	AddLayoutEditor(ctx context.Context, layoutID int, editorID int) error
	RemoveLayoutEditor(ctx context.Context, layoutID int, editorID int) error
	GetDefaultLayout(ctx context.Context) (entity.UserLayout, error)
	LayoutByName(ctx context.Context, layoutName string) (entity.UserLayout, error)
	CreateLayout(ctx context.Context, layout entity.UserLayout) error
}

// Module представляет собой модуль для работы с макетами
type Module struct {
	logger *slog.Logger
	ctx    context.Context

	cache cacheUC
	repo  repoUC
	mu    sync.Mutex
}

// New создает новый экземпляр модуля макетов
func New(cache cacheUC, repo repoUC) *Module {
	return &Module{
		cache: cache,
		repo:  repo,
	}
}

// Init инициализирует модуль макетов
func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.StringValue("📱 LAYOUT"))

	return nil
}
