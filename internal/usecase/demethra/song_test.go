package demethra_test

import (
	"arimadj-helper/internal/application/config"
	"arimadj-helper/internal/application/config/config_module"
	"arimadj-helper/internal/application/env"
	"arimadj-helper/internal/application/env/test"
	"arimadj-helper/internal/application/logger"
	"arimadj-helper/internal/repository/postgres"
	"arimadj-helper/internal/repository/postgres/elysium"
	"arimadj-helper/internal/repository/redis"
	"arimadj-helper/internal/usecase/demethra"
	"context"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"testing"
)

var module *demethra.Module

func testDemethraConfig(t *testing.T) (*config_module.Demethra, *config_module.Postgres, *config_module.Redis) {
	t.Helper()

	t.Setenv("ENV", "test")
	testEnv := test.New()
	envModule := env.New(testEnv)
	storage, err := envModule.Init()
	if err != nil {
		t.Fatalf("Failed to initialize env module: %v", err)
	}

	demethraCfg := config_module.NewDemethraConfig()
	postgresCfg := config_module.NewPostgresConfig()
	redisCfg := config_module.NewRedisConfig()
	cfg := config.New(demethraCfg, postgresCfg, redisCfg)
	err = cfg.Init(storage)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	return demethraCfg, postgresCfg, redisCfg
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

	ctx, cancel := context.WithCancel(context.Background())

	demethraCfg, postgresCfg, redisCfg := testDemethraConfig(t)

	elysiumRepo := elysium.NewRepository()
	redisRepo := redis.New(redisCfg)
	err := redisRepo.Init(ctx, loggerModule)
	require.NoError(t, err)

	postgresql := postgres.New(postgresCfg, elysiumRepo)
	err = postgresql.Init(ctx, loggerModule)
	require.NoError(t, err)

	demethraUC := demethra.New(demethraCfg, elysiumRepo, redisRepo, nil)
	err = demethraUC.Init(ctx, loggerModule)
	require.NoError(t, err)

	module = demethraUC
	return func(t *testing.T) {
		cancel()
		elysiumRepo.Close()
		postgresql.Close()
		redisRepo.Close()
	}
}

func TestSendSongByTrackLink(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	t.Run("Success", func(t *testing.T) {
		link := "https://soundcloud.com/slagwerk/bambinodj-high-as-ever-still"
		userID := 5

		err := module.SendSongByTrackLink(context.Background(), userID, link)
		require.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		link := "https://soundcloud.com/slagwerk/bambinodj-high-as-ever-still"
		userID := -100

		err := module.SendSongByTrackLink(context.Background(), userID, link)
		require.Error(t, err)
	})
}
