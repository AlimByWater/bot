package auth_test

import (
	"context"
	"elysium/internal/application/config"
	"elysium/internal/application/config/config_module"
	"elysium/internal/application/env"
	"elysium/internal/application/env/test"
	"elysium/internal/application/logger"
	"elysium/internal/entity"
	"elysium/internal/repository/postgres"
	"elysium/internal/repository/postgres/elysium"
	"elysium/internal/repository/redis"
	"elysium/internal/usecase/auth"
	"elysium/internal/usecase/layout"
	"elysium/internal/usecase/users"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"log"
	"log/slog"
	"os"
	"testing"
)

var module *auth.Module

func testConfig(t *testing.T) (*config_module.Postgres, *config_module.Auth, *config_module.Redis) {
	t.Helper()

	t.Setenv("ENV", "test")
	testEnv := test.New()
	envModule := env.New(testEnv)
	storage, err := envModule.Init()
	if err != nil {
		t.Fatalf("Failed to initialize env module: %v", err)
	}

	postgresCfg := config_module.NewPostgresConfig()
	authCfg := config_module.NewAuthConfig()
	redisCfg := config_module.NewRedisConfig()
	cfg := config.New(postgresCfg, authCfg, redisCfg)
	err = cfg.Init(storage)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	return postgresCfg, authCfg, redisCfg
}

func setupTest(t *testing.T) func(t *testing.T) {
	// Initialize necessary environment for tests
	// Example: Initialize a mock Redis client, database connection, etc.
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

	postgresCfg, authCfg, redisCfg := testConfig(t)
	elysiumRepo := elysium.NewRepository()
	repo := postgres.New(postgresCfg, elysiumRepo)
	err := repo.Init(ctx, loggerModule)
	require.NoError(t, err)

	redisModule := redis.New(redisCfg)
	err = redisModule.Init(ctx, loggerModule)
	require.NoError(t, err)

	layoutUC := layout.New(redisModule, elysiumRepo)
	layoutUC.Init(ctx, loggerModule)

	userUC := users.New(redisModule, elysiumRepo, layoutUC)
	err = userUC.Init(ctx, loggerModule)
	require.NoError(t, err)

	module = auth.NewModule(authCfg, redisModule, elysiumRepo, userUC)
	err = module.Init(ctx, loggerModule)
	require.NoError(t, err)

	// Return a teardown function to clean up after the test
	return func(t *testing.T) {
		repo.Close()
		redisModule.Close()
	}
}

func TestGenerateTokenForTelegramSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	telegramLogin := entity.TelegramLoginInfo{
		TelegramID: 431778623,
		InitData:   "user=%7B%22id%22%3A431778623%2C%22first_name%22%3A%22Troll%22%2C%22last_name%22%3A%22%22%2C%22username%22%3A%22doev03%22%2C%22language_code%22%3A%22en%22%2C%22is_premium%22%3Atrue%2C%22allows_write_to_pm%22%3Atrue%7D&chat_instance=7030645666426643400&chat_type=sender&auth_date=1722098726&hash=23601cd1f0d1f8d8c02021f755460c88ee715b211451bc04311ffe6874bb4712",
	}

	token, err := module.GenerateTokenForTelegram(context.Background(), telegramLogin)
	require.NoError(t, err)
	require.NotEmpty(t, token.AccessToken)
	require.NotEmpty(t, token.RefreshToken)

	j, _ := json.MarshalIndent(token, "", "  ")
	log.Println(string(j))
}

func TestGenerateTokenForTelegramFailsOnInvalidInitData(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	telegramLogin := entity.TelegramLoginInfo{
		TelegramID: 5534121833,
		InitData:   "invalid_init_data",
	}

	_, err := module.GenerateTokenForTelegram(context.Background(), telegramLogin)
	require.Error(t, err)
}

func TestGenerateTokenForTelegramFailsOnUserNotFound(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	telegramLogin := entity.TelegramLoginInfo{
		TelegramID: 131415151,
		InitData:   "user=%7B%22id%22%3A5534121833%2C%22first_name%22%3A%22Alim%22%2C%22last_name%22%3A%22%22%2C%22username%22%3A%22itsreallyalim%22%2C%22language_code%22%3A%22ru%22%7D&chat_instance=-845713120415025374&chat_type=private&auth_date=1721986575&hash=6d1030e364e8f8df20efb7c519358a3047d37134bab708f01c48534b3593a2c7",
	}

	_, err := module.GenerateTokenForTelegram(context.Background(), telegramLogin)
	require.Error(t, err)
}

func TestGen(t *testing.T) {
	nums := []int{1, 2, 3}

	addNum(nums[0:2])
	fmt.Println(nums) // 1 2 4

	addNums(nums[0:2])
	fmt.Println(nums) // ?
}

func addNum(nums []int) {
	nums = append(nums, 4)
}

func addNums(nums []int) {
	nums = append(nums, 5, 6)
}

func TestModule_CheckAccessTokenByUserID(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	token := entity.Token{
		UserID:       123,
		AccessToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXBlIjoiYWNjZXNzIiwic3ViIjoiNSIsImV4cCI6MTcyNTY1ODUyMCwiaWF0IjoxNzI1NjUxMzIwfQ.YGuF-Js6YVprTgxi78DPeLSWCAXis5TpHAsi4evkB38",
		RefreshToken: "5",
	}

	_, err := module.CheckAccessTokenByUserID(context.Background(), token.AccessToken, token.UserID)
	require.NoError(t, err)
}
