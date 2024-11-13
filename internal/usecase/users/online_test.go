package users_test

import (
	"context"
	"elysium/internal/application/config"
	"elysium/internal/application/config/config_module"
	"elysium/internal/application/env"
	"elysium/internal/application/env/test"
	"elysium/internal/application/logger"
	"elysium/internal/repository/postgres"
	"elysium/internal/repository/postgres/elysium"
	"elysium/internal/repository/redis"
	"elysium/internal/usecase/layout"
	"elysium/internal/usecase/users"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"testing"
	"time"
)

var module *users.Module

func testUsersConfig(t *testing.T) (*config_module.Postgres, *config_module.Redis) {
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
	cfg := config.New(postgresCfg, redisCfg)
	err = cfg.Init(storage)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	return postgresCfg, redisCfg
}

func setupTest(t *testing.T) func(t *testing.T) {
	t.Helper()

	loggerModule := logger.New(
		logger.Options{
			AppName: "test-Bot-manager",
			Writer:  os.Stdout,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		},
	)

	ctx, cancel := context.WithCancel(context.Background())

	postgresCfg, redisCfg := testUsersConfig(t)

	elysiumRepo := elysium.NewRepository()
	redisRepo := redis.New(redisCfg)

	err := redisRepo.Init(ctx, loggerModule)
	require.NoError(t, err)

	postgresql := postgres.New(postgresCfg, elysiumRepo)
	err = postgresql.Init(ctx, loggerModule)
	require.NoError(t, err)

	layoutUC := layout.New(redisRepo, elysiumRepo)
	err = layoutUC.Init(ctx, loggerModule)
	require.NoError(t, err)

	userUC := users.New(redisRepo, elysiumRepo, layoutUC)
	err = userUC.Init(ctx, loggerModule)
	require.NoError(t, err)

	module = userUC
	return func(t *testing.T) {
		cancel()
		elysiumRepo.Close()
		postgresql.Close()
		redisRepo.Close()
	}
}

func TestGetOnlineUsersCount(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	t.Run("Success", func(t *testing.T) {
		time.Sleep(time.Second * 15)
		onlineUsersCount := module.GetOnlineUsersCount()
		require.NotEmpty(t, onlineUsersCount)
		t.Log(onlineUsersCount)
	})
}
