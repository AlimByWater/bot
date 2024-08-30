package auth

import (
	"arimadj-helper/internal/entity"
	"context"
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

	AllTokens(ctx context.Context) ([]entity.Token, error)
	//SaveOrUpdateToken(ctx context.Context, token entity.Token) error
	//TokenByUserID(ctx context.Context, userID int) (entity.Token, error)
}

type Module struct {
	ctx       context.Context
	cfg       config
	logger    *slog.Logger
	jwtSecret []byte

	mu        sync.RWMutex
	tokensMap sync.Map
	//tokens    map[int]entity.Token

	repo repository
	users *users.Module
}

func NewModule(cfg config, repo repository, users *users.Module) *Module {
	return &Module{
		cfg:       cfg,
		repo:      repo,
		users:     users,
		tokensMap: sync.Map{},
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

	tokens, err := m.repo.AllTokens(ctx)
	if err != nil {
		return fmt.Errorf("all tokens: %w", err)
	}

	for _, token := range tokens {
		m.tokensMap.Store(token.UserID, token)
	}

	return nil
}
