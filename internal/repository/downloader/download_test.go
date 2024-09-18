package downloader_test

import (
	"bytes"
	"context"
	"elysium/internal/application/config"
	"elysium/internal/application/config/config_module"
	"elysium/internal/application/env"
	"elysium/internal/application/env/test"
	"elysium/internal/application/logger"
	"elysium/internal/repository/downloader"
	"github.com/bogem/id3v2"
	"github.com/stretchr/testify/require"
	"log/slog"
	"testing"
)

var downloaderModule *downloader.Module

func testConfig(t *testing.T) *config_module.Downloader {
	t.Helper()

	t.Setenv("ENV", "test")
	testEnv := test.New()
	envModule := env.New(testEnv)
	storage, err := envModule.Init()
	if err != nil {
		t.Fatalf("Failed to initialize env module: %v", err)
	}

	downloaderCfg := config_module.NewDownloaderConfig()
	cfg := config.New(downloaderCfg)
	err = cfg.Init(storage)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	return downloaderCfg
}

func setupTest(t *testing.T) func(t *testing.T) {
	t.Helper()

	cfg := testConfig(t)
	ctx := context.Background()
	logger := logger.New(
		logger.Options{
			AppName: "test-bot-manager",
			Writer:  nil,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		},
	)

	downloaderModule = downloader.New(cfg)
	err := downloaderModule.Init(ctx, logger)
	require.NoError(t, err)

	return func(t *testing.T) {
		err := downloaderModule.Close()
		if err != nil {
			t.Fatalf("Failed to shutdown downloader module: %v", err)
		}
	}
}

func TestModule_Download(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	t.Run("Success", func(t *testing.T) {
		fileName, data, err := downloaderModule.DownloadByLink(context.Background(), "https://www.youtube.com/watch?v=3QIGeNqi6l0&ab_channel=Slipknot", "mp3")
		require.NoError(t, err)

		require.NotEmpty(t, fileName)
		require.NotEmpty(t, data)

		file := bytes.NewReader(data)
		tag, err := id3v2.ParseReader(file, id3v2.Options{Parse: true})
		require.NoError(t, err)
		t.Log(tag.Title())
		t.Log(tag.Artist())

		err = downloaderModule.RemoveFile(context.Background(), fileName)
		require.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		_, _, err := downloaderModule.DownloadByLink(context.Background(), "https://example.com/file.mp3", "mp3")
		require.Error(t, err)
	})
}
