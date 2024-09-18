package demethra_test

import (
	"bufio"
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
	"elysium/internal/usecase/demethra"
	"fmt"
	"github.com/essentialkaos/go-icecast"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"os"
	"strconv"
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

func TestIceacst(t *testing.T) {
	api, err := icecast.NewAPI("http://91.206.15.29:8000", "alim", "hackme8")
	require.NoError(t, err)

	stats, err := api.GetStats()
	require.NoError(t, err)
	fmt.Println(stats)
	//ticke := time.NewTicker(time.Second * 1)
	//for range ticke.C {
	//	err = api.UpdateMeta("/stream", icecast.TrackMeta{
	//		Title:   "test",
	//		Artist:  "Arima DJ",
	//		URL:     "https://soundcloud.com/uiceheidd/tell-me-you-love-me",
	//		Artwork: "https://i1.sndcdn.com/artworks-oQRvHcKyeO921Eve-FeUQMA-t50x50.jpg",
	//	})
	//
	//	require.NoError(t, err)
	//}

}

func TestUpdateSongMetadataFile(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	t.Run("Success", func(t *testing.T) {
		track := entity.TrackInfo{
			TrackTitle: "test-title",
			ArtistName: "test-artist",
			TrackLink:  "test-link",
			CoverLink:  "test-cover-link",
		}
		module.UpdateSongMetadataFile(track)

		track = entity.TrackInfo{
			TrackTitle: "test-title23",
			ArtistName: "test-artist23",
			TrackLink:  "test-link3",
			CoverLink:  "test-cover-link",
		}
		module.UpdateSongMetadataFile(track)

	})
}

func TestIcecastMetadata(t *testing.T) {
	client := &http.Client{}
	streamUrl := "https://elysiumfm.ru/stream"
	req, _ := http.NewRequest("GET", streamUrl, nil)
	req.Header.Set("Icy-MetaData", "1")
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	// We sent "Icy-MetaData", we should have a "icy-metaint" in return
	ih := resp.Header.Get("icy-metaint")
	require.NotEmpty(t, ih)
	// "icy-metaint" is how often (in bytes) should we receive the meta
	ib, err := strconv.Atoi(ih)
	require.NoError(t, err)

	reader := bufio.NewReader(resp.Body)

	// skip the first mp3 frame
	c, err := reader.Discard(ib)
	require.NoError(t, err)
	// If we didn't received ib bytes, the stream is ended
	if c != ib {
		t.Fatal("stream ended prematurally")
	}

	// get the size byte, that is the metadata length in bytes / 16
	sb, err := reader.ReadByte()
	require.NoError(t, err)
	ms := int(sb * 16)

	// read the ms first bytes it will contain metadata
	m, err := reader.Peek(ms)
	require.NoError(t, err)

	t.Log(string(m))
}
