package layout

import (
	"arimadj-helper/internal/entity"
	"context"
	"log/slog"
	"sync"
)

// cacheUC интерфейс для работы с кэшем
type cacheUC interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
}

// repoUC интерфейс для работы с репозиторием макетов
type repoUC interface {
	LayoutByUserID(ctx context.Context, userID int) (entity.UserLayout, error)
	LayoutByID(ctx context.Context, layoutID string) (entity.UserLayout, error)
	UpdateLayout(ctx context.Context, layout entity.UserLayout) error
	LogLayoutChange(ctx context.Context, change entity.LayoutChange) error
	IsLayoutOwner(ctx context.Context, layoutID string, userID int) (bool, error)
	AddLayoutEditor(ctx context.Context, layoutID string, editorID int) error
	RemoveLayoutEditor(ctx context.Context, layoutID string, editorID int) error
	GetDefaultLayout(ctx context.Context) (entity.UserLayout, error)
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
