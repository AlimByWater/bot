package redis

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
)

type config interface {
	GetHost() string
	GetPassword() string
	GetPort() int
	GetDB() int
	Validate() error
}

// Module - структура модуля репозитория
type Module struct {
	cfg    config
	client *redis.Client
}

// New - создает новый модуль, на входе конфигурация и таблицы
func New(cfg config) *Module {
	return &Module{
		cfg: cfg,
	}
}

func (m *Module) Init(ctx context.Context, _ *slog.Logger) (err error) {
	if err := m.cfg.Validate(); err != nil {
		return fmt.Errorf("redis config validate: %w", err)
	}

	m.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", m.cfg.GetHost(), m.cfg.GetPort()),
		Username: "default",
		Password: m.cfg.GetPassword(),
		DB:       m.cfg.GetDB(),
	})

	// Ping the Redis server to check the connection
	pong, err := m.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}

	_ = pong

	return nil
}

// Close - закрывает модуль при завершении работы приложения
func (m *Module) Close() (err error) {
	if m.client != nil {
		err = m.client.Close()
		if err != nil {
			return fmt.Errorf("redis close: %w", err)
		}
	}
	return
}
