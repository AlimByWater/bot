package elysium_test

import (
	"context"
	"elysium/internal/application/config"
	"elysium/internal/application/config/config_module"
	"elysium/internal/application/env"
	"elysium/internal/application/env/test"
	"elysium/internal/application/logger"
	"elysium/internal/repository/postgres"
	"elysium/internal/repository/postgres/elysium"
	"log/slog"
	"os"
	"testing"
)

var (
	elysiumRepo *elysium.Repository
	postgresql  *postgres.Module
)

func testConfig(t *testing.T) *config_module.Postgres {
	t.Helper()

	t.Setenv("ENV", "test")
	testEnv := test.New()
	envModule := env.New(testEnv)
	storage, err := envModule.Init()
	if err != nil {
		t.Fatalf("Failed to initialize env module: %v", err)
	}

	postgresCfg := config_module.NewPostgresConfig()
	cfg := config.New(postgresCfg)
	err = cfg.Init(storage)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	return postgresCfg
}

func TestMain(m *testing.M) {
	// Настройка перед всеми тестами
	ctx := context.Background()
	log := logger.New(
		logger.Options{
			AppName: "test-bot-manager",
			Writer:  os.Stdout,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		},
	)

	// Инициализация репозиториев
	cfg := testConfig(&testing.T{})
	elysiumRepo = elysium.NewRepository()
	postgresql = postgres.New(cfg, elysiumRepo)
	err := postgresql.Init(ctx, log)
	if err != nil {
		log.Error("Failed to initialize postgres repository", "error", err)
		os.Exit(1)
	}

	// Запуск тестов
	code := m.Run()

	// Очистка после всех тестов
	postgresql.Close()

	os.Exit(code)
}

func setupTest(t *testing.T) func(t *testing.T) {
	t.Helper()

	cfg := testConfig(t)
	elysiumRepo = elysium.NewRepository()
	postgresql = postgres.New(cfg, elysiumRepo)
	err := postgresql.Init(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to initialize postgres repository: %v", err)

	}

	// Возвращаем функцию очистки
	return func(t *testing.T) {
		if err := postgresql.Close(); err != nil {
			t.Errorf("Failed to close postgresql connection: %v", err)
		}
	}
}
