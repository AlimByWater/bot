package users

import (
	"context"
	"elysium/internal/entity"
	"log/slog"
	"sync"
	"sync/atomic"
)

type cacheUC interface {
	GetListenersCount(ctx context.Context) (map[string]int64, error)
	GetAllCurrentListeners(ctx context.Context) ([]entity.ListenerCache, error)
	GetUserByTelegramIDCache(ctx context.Context, telegramID int64) (entity.User, error)
	SaveOrUpdateUserCache(ctx context.Context, user entity.User) error
	RemoveUserCache(ctx context.Context, userID int) error
	GetUserByID(ctx context.Context, userID int) (entity.User, error)
}

type repository interface {
	DeleteUser(ctx context.Context, userID int) error
	CreateOrUpdateUser(ctx context.Context, user entity.User) (entity.User, error)
	GetUserByTelegramID(ctx context.Context, telegramID int64) (entity.User, error)
	GetUserByID(ctx context.Context, userID int) (entity.User, error)
	SetUserToBotActive(ctx context.Context, userID int, botID int64) error
	GetUserActiveBots(ctx context.Context, userID int) ([]entity.Bot, error)

	CreateTransaction(ctx context.Context, txn entity.UserTransaction) (entity.UserTransaction, error)
	GetTransactionByID(ctx context.Context, txnID string) (entity.UserTransaction, error)
	ProcessTransaction(ctx context.Context, txnID string, userID int, balanceChange int, newStatus string) error
	UpdateTransactionExternalID(ctx context.Context, txnID string, externalID string) error

	// Методы для работы с балансом пользователя
	GetUserAccount(ctx context.Context, userID int) (entity.UserAccount, error)
	CreateUserAccount(ctx context.Context, account entity.UserAccount) error
	UpdateUserBalance(ctx context.Context, userID int, newBalance int) error
	GetUserBalance(ctx context.Context, userID int) (int, error)
}

type servicesUC interface {
	GetService(ctx context.Context, botID int64, serviceName string) (entity.Service, error)
	GetServices(ctx context.Context, botID int64) ([]entity.Service, error)
}

type Module struct {
	logger *slog.Logger
	ctx    context.Context

	cache     cacheUC
	repo      repository
	servicesM servicesUC

	onlineUsersCount   atomic.Int64
	mu                 sync.RWMutex
	streamsOnlineCount map[string]int64
}

func New(cache cacheUC, repo repository, servicesM servicesUC) *Module {
	return &Module{
		cache:     cache,
		repo:      repo,
		servicesM: servicesM,
	}
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.String("module", "🏓 USERS"))

	go m.updateOnlineUsersCountLoop()
	return nil
}
