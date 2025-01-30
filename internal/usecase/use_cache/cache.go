package use_cache

import (
	"context"
	"log/slog"
	"sync"
)

type Module struct {
	cache sync.Map
}

func New() *Module {
	return &Module{}
}

func (m *Module) Init(_ context.Context, _ *slog.Logger) error {
	return nil
}

func (m *Module) Store(key string, value any) {
	m.cache.Store(key, value)
	return
}

func (m *Module) LoadAndDelete(key string) (value any, loaded bool) {
	value, loaded = m.cache.LoadAndDelete(key)
	return
}

func (m *Module) Load(key string) (value any, loaded bool) {
	value, loaded = m.cache.Load(key)
	return
}
