package elysium_test

import (
	"context"
	"elysium/internal/application/config"
	"elysium/internal/application/config/config_module"
	"elysium/internal/application/env"
	"elysium/internal/application/env/test"
	"elysium/internal/application/logger"
	"elysium/internal/repository/clickhouse"
	"elysium/internal/repository/clickhouse/transaction_audit_insert"
	"elysium/internal/repository/postgres"
	"elysium/internal/repository/postgres/elysium"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"testing"
)

var (
	elysiumRepo *elysium.Repository
	postgresql  *postgres.Module
)

func testConfig(t *testing.T) (*config_module.Postgres, *config_module.Clickhouse) {
	t.Helper()

	t.Setenv("ENV", "test")
	testEnv := test.New()
	envModule := env.New(testEnv)
	storage, err := envModule.Init()
	if err != nil {
		t.Fatalf("Failed to initialize env module: %v", err)
	}

	postgresCfg := config_module.NewPostgresConfig()
	clickhouseCfg := config_module.NewClickhouseConfig()
	cfg := config.New(postgresCfg, clickhouseCfg)
	err = cfg.Init(storage)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	return postgresCfg, clickhouseCfg
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

	t := &testing.T{}
	// Инициализация репозиториев
	cfg, clickCfg := testConfig(t)

	clickhouseRepo := clickhouse.New(clickCfg)
	transactionsAuditInsert := transaction_audit_insert.NewInsertTable()
	clickhouseRepo.AddInsertTable(transactionsAuditInsert)
	err := clickhouseRepo.Init(ctx, log)
	require.NoError(t, err)

	elysiumRepo = elysium.NewRepository(transactionsAuditInsert)
	postgresql = postgres.New(cfg, elysiumRepo)
	err = postgresql.Init(ctx, log)
	if err != nil {
		log.Error("Failed to initialize postgres repository", "error", err)
		os.Exit(1)
	}

	// Запуск тестов
	code := m.Run()

	// Очистка после всех тестов
	postgresql.Close()
	clickhouseRepo.Close()

	os.Exit(code)
}

func setupTest(t *testing.T) func(t *testing.T) {
	t.Helper()

	loggerModule := logger.New(
		logger.Options{
			AppName: "test-bot-manager",
			Writer:  os.Stdout,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		},
	)

	cfg, clickCfg := testConfig(t)

	clickhouseRepo := clickhouse.New(clickCfg)
	transactionsAuditInsert := transaction_audit_insert.NewInsertTable()
	clickhouseRepo.AddInsertTable(transactionsAuditInsert)
	err := clickhouseRepo.Init(context.Background(), loggerModule)
	require.NoError(t, err)

	elysiumRepo = elysium.NewRepository(transactionsAuditInsert)
	postgresql = postgres.New(cfg, elysiumRepo)
	err = postgresql.Init(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to initialize postgres repository: %v", err)

	}

	// Возвращаем функцию очистки
	return func(t *testing.T) {
		if err := postgresql.Close(); err != nil {
			t.Errorf("Failed to close postgresql connection: %v", err)
		}

		if err := clickhouseRepo.Close(); err != nil {
			t.Errorf("Failed to close clickhouse connection: %v", err)
		}
	}
}
