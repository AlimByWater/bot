package soundcloud

import (
	"arimadj-helper/internal/application/config"
	"arimadj-helper/internal/application/config/config_module"
	"arimadj-helper/internal/application/env"
	"arimadj-helper/internal/application/env/test"
	"arimadj-helper/internal/application/logger"
	"arimadj-helper/internal/entity"
	"context"
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"testing"
)

var sc *Module

func testConfig(t *testing.T) *config_module.Soundcloud {
	t.Helper()

	t.Setenv("ENV", "test")
	testEnv := test.New()
	envModule := env.New(testEnv)
	storage, err := envModule.Init()
	if err != nil {
		t.Fatalf("Failed to initialize env module: %v", err)
	}

	scConfig := config_module.NewSoundcloudConfig()
	cfg := config.New(scConfig)
	err = cfg.Init(storage)
	if err != nil {
		t.Fatalf("Failed to initialize config: %v", err)
	}

	return scConfig
}

func testLogger(t *testing.T) *slog.Logger {
	t.Helper()

	loggerModule := logger.New(
		logger.Options{
			AppName: "bot-manager",
			Writer:  os.Stdout,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		},
	)

	return loggerModule
}

func setupTest(t *testing.T) func(t *testing.T) {
	t.Helper()

	cfg := testConfig(t)
	log := testLogger(t)

	sc = New(cfg)
	ctx := context.Background()

	err := sc.Init(ctx, log)
	if err != nil {
		t.Fatalf("Failed to initialize soundcloud repository: %v", err)
	}

	return func(t *testing.T) {

	}
}

func TestModule_DownloadTrackByURL(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	ctx := context.Background()
	trackUrl := "https://soundcloud.com/eande/kelela-enemy-dj-e-edit"
	//trackUrl := "https://soundcloud.com/love_again/dj-lostboi-x-young-thug?in=sungodarima/sets/copy-of-related-tracks-500-dj/s-zzFAANRSzyr&si=15b323716ebc4df691fa147b589201aa&utm_source=clipboard&utm_medium=text&utm_campaign=social_sharing"
	info := entity.TrackInfo{}

	trackPath, err := sc.DownloadTrackByURL(ctx, trackUrl, info)
	require.NoError(t, err)
	require.NotEmpty(t, trackPath)
}

func TestDownloadTrackByURLV2(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)
	//
	//trackUrl := "https://soundcloud.com/love_again/dj-lostboi-x-young-thug?in=sungodarima/sets/copy-of-related-tracks-500-dj/s-zzFAANRSzyr&si=15b323716ebc4df691fa147b589201aa&utm_source=clipboard&utm_medium=text&utm_campaign=social_sharing"
	//info := entity.TrackInfo{
	//	TrackTitle: "Chou Chou ",
	//	ArtistName: "do you want to feel something real?",
	//	TrackLink:  trackUrl,
	//	CoverLink:  "https://i1.sndcdn.com/artworks-nAHorXThQpzWN8pQ-CfwHEw-t500x500.jpg",
	//}
	//
	//songPath, err := m.sc.DownloadByUrl(trackUrl, "/Users/admin/go/src/arimadj-helper/", info)
	//require.NoError(t, err)
	//require.NotEmpty(t, songPath)
}
