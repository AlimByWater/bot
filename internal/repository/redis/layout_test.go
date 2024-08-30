package redis_test

import (
	"arimadj-helper/internal/application/config"
	"arimadj-helper/internal/application/config/config_module"
	"arimadj-helper/internal/application/env"
	"arimadj-helper/internal/application/env/test"
	"arimadj-helper/internal/entity"
	redisRepo "arimadj-helper/internal/repository/redis"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
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
		t.Fatalf("Failed to initialize redis repository: %v", err)
	}

	return func(t *testing.T) {
		err = redisModule.Close()
	}
}

func TestSaveLayoutSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	layout := entity.Layout{
		Background: entity.Background{
			Type:  "color",
			Value: "#FFFFFF",
		},
	}

	err := redisModule.SaveLayout(context.Background(), layout)
	require.NoError(t, err)
}

func TestGetLayoutSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	// Preparing a layout to be retrieved
	layout := entity.Layout{
		Background: entity.Background{
			Type:  "image",
			Value: "https://example.com/image.jpg",
		},
	}
	err := redisModule.SaveLayout(context.Background(), layout)
	require.NoError(t, err)

	// Retrieving the layout
	retrievedLayout, err := redisModule.GetLayout(context.Background())
	require.NoError(t, err)
	require.Equal(t, layout.Background.Type, retrievedLayout.Background.Type)
	require.Equal(t, layout.Background.Value, retrievedLayout.Background.Value)
}

func TestGetLayoutReturnsErrorWhenNoLayout(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	// Attempting to retrieve a layout when none exists
	_, err := redisModule.GetLayout(context.Background())
	require.Error(t, err)
	require.Equal(t, redisRepo.ErrLayoutNotFound, err)
}

func TestSaveLayoutUpdatesExistingLayout(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	initialLayout := entity.Layout{
		Background: entity.Background{
			Type:  "color",
			Value: "#000000",
		},
	}

	err := redisModule.SaveLayout(context.Background(), initialLayout)
	require.NoError(t, err)

	updatedLayout := entity.Layout{
		Background: entity.Background{
			Type:  "image",
			Value: "https://example.com/new_image.jpg",
		},
	}

	err = redisModule.SaveLayout(context.Background(), updatedLayout)
	require.NoError(t, err)

	retrievedLayout, err := redisModule.GetLayout(context.Background())
	require.NoError(t, err)
	require.Equal(t, updatedLayout.Background.Type, retrievedLayout.Background.Type)
	require.Equal(t, updatedLayout.Background.Value, retrievedLayout.Background.Value)
}
