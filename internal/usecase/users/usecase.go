package users

import (
	"arimadj-helper/internal/entity"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

type cacheUC interface {
	GetListenersCount(ctx context.Context) (int64, error)
	GetAllCurrentListeners(ctx context.Context) ([]entity.ListenerCache, error)
}

type repoUC interface {
	CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)
}

type Module struct {
	logger *slog.Logger
	ctx    context.Context

	cache cacheUC
	repo  repoUC

	onlineUsersCount atomic.Int64
	mu               sync.Mutex
}

func New(cache cacheUC, repo repoUC) *Module {
	return &Module{
		cache: cache,
		repo:  repo,
	}
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.StringValue("🏓 USERS"))

	go m.updateOnlineUsersCountLoop()
	return nil
}

func (m *Module) CreateUser(ctx context.Context, user entity.User) (entity.User, error) {
	createdUser, err := m.repo.CreateOrUpdateUser(ctx, user)
	if err != nil {
		m.logger.Error("Failed to create user", slog.Any("error", err), slog.Any("user", user))
		return entity.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	m.logger.Info("User created successfully", slog.Any("user", createdUser))
	return createdUser, nil
}
