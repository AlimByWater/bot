package redis_test

import (
	"arimadj-helper/internal/application/config"
	"arimadj-helper/internal/application/config/config_module"
	"arimadj-helper/internal/application/env"
	"arimadj-helper/internal/application/env/test"
	"arimadj-helper/internal/entity"
	redisRepo "arimadj-helper/internal/repository/redis"
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var redisModule *redisRepo.Module

func testConfig(t *testing.T) *config_module.Redis {
	t.Helper()

	t.Setenv("ENV", "test")
	testEnv := test.New()
	envModule := env.New(testEnv)
	storage, err := envModule.Init()
	if err != nil {
		t.Fatalf("Failed to initialize env module: %v", err)
	}

	redisCfg := config_module.NewRedisConfig()
	cfg := config.New(redisCfg)
	err = cfg.Init(storage)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	return redisCfg
}

func setupTest(t *testing.T) func(t *testing.T) {
	t.Helper()

	cfg := testConfig(t)
	redisModule = redisRepo.New(cfg)
	err := redisModule.Init(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to initialize postgres repository: %v", err)

	}

	return func(t *testing.T) {
		err = redisModule.Close()
	}
}

func TestSaveListenerSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	listenerCache := entity.ListenerCache{
		TelegramID: 123456789,
		Payload:    entity.ListenerCachePayload{InitTimestamp: time.Now().Add(-1 * time.Hour).Unix(), LastActivity: time.Now().Unix()},
	}

	err := redisModule.SaveOrUpdateListener(context.Background(), listenerCache)
	require.NoError(t, err)
}

func TestSaveListenerFailsOnEmptyTelegramID(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	listenerCache := entity.ListenerCache{
		TelegramID: 0,
		Payload:    entity.ListenerCachePayload{InitTimestamp: time.Now().Add(-1 * time.Hour).Unix(), LastActivity: time.Now().Unix()},
	}

	err := redisModule.SaveOrUpdateListener(context.Background(), listenerCache)
	require.Error(t, err)
}

func TestSaveListenerSuccessOnUpdate(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	telegramId := int64(123456789)
	listenerCache1 := entity.ListenerCache{
		TelegramID: telegramId,
		Payload: entity.ListenerCachePayload{
			InitTimestamp: time.Now().Add(-1 * time.Hour).Unix(),
			LastActivity:  time.Now().Add(-30 * time.Minute).Unix(),
		},
	}

	err := redisModule.SaveOrUpdateListener(context.Background(), listenerCache1)
	require.NoError(t, err)

	listenerCache2 := entity.ListenerCache{
		TelegramID: telegramId,
		Payload: entity.ListenerCachePayload{
			InitTimestamp: time.Now().Add(-1 * time.Hour).Unix(),
			LastActivity:  time.Now().Unix(),
		},
	}

	err = redisModule.SaveOrUpdateListener(context.Background(), listenerCache2)
	require.NoError(t, err)

	listenerCache, err := redisModule.GetListenerByTelegramID(context.Background(), telegramId)
	require.NoError(t, err)
	require.Equal(t, listenerCache.Payload.LastActivity, listenerCache2.Payload.LastActivity)
}

func TestGetListenerByTelegramIDSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	// Preparing a listener cache to be retrieved
	listenerCache := entity.ListenerCache{
		TelegramID: 123456789,
		Payload:    entity.ListenerCachePayload{InitTimestamp: time.Now().Unix(), LastActivity: time.Now().Unix()},
	}
	err := redisModule.SaveOrUpdateListener(context.Background(), listenerCache)
	require.NoError(t, err)

	// Retrieving the listener cache
	retrievedCache, err := redisModule.GetListenerByTelegramID(context.Background(), 123456789)
	require.NoError(t, err)
	require.Equal(t, listenerCache.TelegramID, retrievedCache.TelegramID)
	require.Equal(t, listenerCache.Payload.InitTimestamp, retrievedCache.Payload.InitTimestamp)
	require.Equal(t, listenerCache.Payload.LastActivity, retrievedCache.Payload.LastActivity)
}

func TestGetListenerByTelegramIDFailsOnNonexistentID(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	// Attempting to retrieve a listener cache with a Telegram ID that does not exist
	_, err := redisModule.GetListenerByTelegramID(context.Background(), 0)
	require.Error(t, err)
	require.Equal(t, err, redisRepo.ErrTelegramIDRequired)
}

func TestModule_GetAllListeners(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	// Preparing a listener cache to be retrieved
	listenerCache := entity.ListenerCache{
		TelegramID: 123456789,
		Payload:    entity.ListenerCachePayload{InitTimestamp: time.Now().Unix(), LastActivity: time.Now().Unix()},
	}
	err := redisModule.SaveOrUpdateListener(context.Background(), listenerCache)
	require.NoError(t, err)

	// Retrieving all listener caches
	count, err := redisModule.GetListenersCount(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, count)
}

func TestGetListenerLastActivitySucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	// Preparing a listener cache with last activity
	telegramID := int64(123456789)
	listenerCache := entity.ListenerCache{
		TelegramID: telegramID,
		Payload:    entity.ListenerCachePayload{InitTimestamp: time.Now().Unix(), LastActivity: time.Now().Unix()},
	}
	err := redisModule.SaveOrUpdateListener(context.Background(), listenerCache)
	require.NoError(t, err)

	lastActivity, err := redisModule.GetListenerLastActivityByTelegramID(context.Background(), telegramID)
	require.NoError(t, err)
	require.NotZero(t, lastActivity)
}

func TestGetListenerLastActivityFailsOnMissingTelegramID(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	_, err := redisModule.GetListenerLastActivityByTelegramID(context.Background(), 0)
	require.Error(t, err)
	require.Equal(t, redisRepo.ErrTelegramIDRequired, err)

}

func TestGetListenerLastActivityFailsOnNonexistentListener(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	_, err := redisModule.GetListenerLastActivityByTelegramID(context.Background(), int64(987654321))
	require.Error(t, err)
	require.Equal(t, true, errors.Is(err, redisRepo.ErrListenerNotFound))
}

func TestSaveOrUpdateListenerSucceedsOnNewListener(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	listenerCache := entity.ListenerCache{
		TelegramID: 1234567891,
		Payload:    entity.ListenerCachePayload{InitTimestamp: time.Now().Unix(), LastActivity: time.Now().Unix()},
	}

	err := redisModule.SaveOrUpdateListener(context.Background(), listenerCache)
	require.NoError(t, err)

	retrievedCache, err := redisModule.GetListenerByTelegramID(context.Background(), listenerCache.TelegramID)
	require.NoError(t, err)
	require.Equal(t, listenerCache.TelegramID, retrievedCache.TelegramID)
	require.Equal(t, listenerCache.Payload.InitTimestamp, retrievedCache.Payload.InitTimestamp)
	require.Equal(t, listenerCache.Payload.LastActivity, retrievedCache.Payload.LastActivity)
}

func TestSaveOrUpdateListenerUpdatesExistingListener(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	telegramID := int64(123456789)
	initialListenerCache := entity.ListenerCache{
		TelegramID: telegramID,
		Payload:    entity.ListenerCachePayload{InitTimestamp: time.Now().Add(-24 * time.Hour).Unix(), LastActivity: time.Now().Add(-24 * time.Hour).Unix()},
	}

	err := redisModule.SaveOrUpdateListener(context.Background(), initialListenerCache)
	require.NoError(t, err)

	updatedListenerCache := entity.ListenerCache{
		TelegramID: telegramID,
		Payload:    entity.ListenerCachePayload{LastActivity: time.Now().Unix()},
	}

	err = redisModule.SaveOrUpdateListener(context.Background(), updatedListenerCache)
	require.NoError(t, err)

	retrievedCache, err := redisModule.GetListenerByTelegramID(context.Background(), telegramID)
	require.NoError(t, err)
	require.Equal(t, updatedListenerCache.Payload.LastActivity, retrievedCache.Payload.LastActivity)
}

func TestSaveOrUpdateListenerFailsOnMissingTelegramID(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	listenerCache := entity.ListenerCache{}

	err := redisModule.SaveOrUpdateListener(context.Background(), listenerCache)
	require.Error(t, err)
	require.Equal(t, redisRepo.ErrTelegramIDRequired, err)
}

func TestGetAllCurrentListenersReturnsListenersSuccessfully(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	// Preparing multiple listener caches to be retrieved
	listenerCaches := []entity.ListenerCache{
		{
			TelegramID: 123456789,
			Payload:    entity.ListenerCachePayload{InitTimestamp: time.Now().Unix(), LastActivity: time.Now().Unix()},
		},
		{
			TelegramID: 987654321,
			Payload:    entity.ListenerCachePayload{InitTimestamp: time.Now().Unix(), LastActivity: time.Now().Unix()},
		},
	}
	for _, cache := range listenerCaches {
		err := redisModule.SaveOrUpdateListener(context.Background(), cache)
		require.NoError(t, err)
	}

	// Retrieving all listener caches
	retrievedListeners, err := redisModule.GetAllCurrentListeners(context.Background())
	require.NoError(t, err)
	require.Len(t, retrievedListeners, len(listenerCaches))
}

func TestGetAllCurrentListenersReturnsEmptySliceWhenNoListeners(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	// Ensuring no listeners are present
	retrievedListeners, err := redisModule.GetAllCurrentListeners(context.Background())
	require.NoError(t, err)
	require.Empty(t, retrievedListeners)
}
