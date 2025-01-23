package auth

import (
	"context"
	"elysium/internal/entity"
	"fmt"
	"log/slog"
	"sync"
)

type config interface {
	GetSecret() string
	GetTelegramBotToken() string
	Validate() error
}

type repository interface {
	GetUserByID(ctx context.Context, userID int) (entity.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID int64) (entity.User, error)
	CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)
	//SaveOrUpdateToken(ctx context.Context, token entity.Token) error
	//TokenByUserID(ctx context.Context, userID int) (entity.Token, error)
}

type externalCache interface {
	GetToken(ctx context.Context, userID int) (entity.Token, error)
	SetToken(token entity.Token) error
	AllTokens(ctx context.Context) ([]entity.Token, error)
}

type userCreator interface {
	CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)
}

type Module struct {
	ctx       context.Context
	cfg       config
	logger    *slog.Logger
	jwtSecret []byte

	mu        sync.RWMutex
	tokensMap sync.Map
	cache     externalCache

	repo  repository
	users userCreator
}

func NewModule(cfg config, cache externalCache, repo repository, users userCreator) *Module {
	return &Module{
		cfg:       cfg,
		repo:      repo,
		users:     users,
		tokensMap: sync.Map{},
		cache:     cache,
		mu:        sync.RWMutex{},
	}
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.String("module", "🔐 "+"auth_methods"))
	if err := m.cfg.Validate(); err != nil {
		return err
	}

	m.jwtSecret = []byte(m.cfg.GetSecret())

	tokens, err := m.cache.AllTokens(ctx)
	if err != nil {
		return fmt.Errorf("all tokens: %w", err)
	}

	for _, token := range tokens {
		m.tokensMap.Store(token.UserID, token)
	}

	return nil
}
