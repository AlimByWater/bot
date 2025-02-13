package users_test

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
	"elysium/internal/repository/redis"
	"elysium/internal/usecase/services"
	"elysium/internal/usecase/users"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"testing"
)

var module *users.Module

func TestMain(m *testing.M) {
	ctx := context.Background()
	loggerModule := logger.New(
		logger.Options{
			AppName: "test-bot-manager",
			Writer:  os.Stdout,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		},
	)

	t := &testing.T{}

	postgresCfg, redisCfg, clickCfg := testUsersConfig(t)
	clickhouseRepo := clickhouse.New(clickCfg)
	transactionsAuditInsert := transaction_audit_insert.NewInsertTable()
	clickhouseRepo.AddInsertTable(transactionsAuditInsert)
	err := clickhouseRepo.Init(ctx, loggerModule)
	require.NoError(t, err)

	elysiumRepo := elysium.NewRepository(transactionsAuditInsert)
	redisRepo := redis.New(redisCfg)

	err = redisRepo.Init(ctx, loggerModule)
	require.NoError(t, err)

	postgresql := postgres.New(postgresCfg, elysiumRepo)
	err = postgresql.Init(ctx, loggerModule)
	require.NoError(t, err)

	servicesUC := services.New(redisRepo, elysiumRepo)
	err = servicesUC.Init(ctx, loggerModule)
	require.NoError(t, err)

	userUC := users.New(redisRepo, elysiumRepo, servicesUC)
	err = userUC.Init(ctx, loggerModule)
	require.NoError(t, err)

	module = userUC
	code := m.Run()

	// Очистка после всех тестов
	postgresql.Close()
	redisRepo.Close()
	clickhouseRepo.Close()

	os.Exit(code)

}

func testUsersConfig(t *testing.T) (*config_module.Postgres, *config_module.Redis, *config_module.Clickhouse) {
	t.Helper()

	t.Setenv("ENV", "test")
	testEnv := test.New()
	envModule := env.New(testEnv)
	storage, err := envModule.Init()
	if err != nil {
		t.Fatalf("Failed to initialize env module: %v", err)
	}

	postgresCfg := config_module.NewPostgresConfig()
	redisCfg := config_module.NewRedisConfig()
	clickCfg := config_module.NewClickhouseConfig()
	cfg := config.New(postgresCfg, redisCfg, clickCfg)
	err = cfg.Init(storage)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	return postgresCfg, redisCfg, clickCfg
}
