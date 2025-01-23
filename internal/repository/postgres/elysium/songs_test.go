package elysium_test

import (
	"context"
	"elysium/internal/entity"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSong(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	t.Run("CreateSong with valid data", func(t *testing.T) {
		song := entity.Song{
			URL:                       "test-url",
			ArtistName:                "test-artist",
			Title:                     "test-title",
			CoverLink:                 "test-cover-link",
			CoverTelegramFileID:       "test-cover-telegram-file-id",
			SongTelegramMessageID:     123,
			SongTelegramMessageChatID: 123,
			Tags:                      []string{"test-tag1", "test-tag2"},
		}

		song, err := elysiumRepo.CreateSong(context.Background(), song)
		require.NoError(t, err)

		err = elysiumRepo.RemoveSong(context.Background(), song.ID)
		require.NoError(t, err)
	})

	t.Run("CreateSong with missing required fields", func(t *testing.T) {
		song := entity.Song{
			URL: "test-artist",
		}

		_, err := elysiumRepo.CreateSong(context.Background(), song)
		require.Error(t, err)
	})

	t.Run("CreateSong with existing URL", func(t *testing.T) {
		song := entity.Song{
			URL:                       "existing-url",
			ArtistName:                "test-artist",
			Title:                     "test-title",
			CoverLink:                 "test-cover-link",
			CoverTelegramFileID:       "test-cover-telegram-file-id",
			SongTelegramMessageID:     123,
			SongTelegramMessageChatID: 123,
			Tags:                      []string{"test-tag1", "test-tag2"},
		}

		song, err := elysiumRepo.CreateSong(context.Background(), song)
		require.NoError(t, err)

		_, err = elysiumRepo.CreateSong(context.Background(), song)
		require.Error(t, err)

		err = elysiumRepo.RemoveSong(context.Background(), song.ID)
		require.NoError(t, err)
	})
}

func TestSongByUrl(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	t.Run("SongByUrl with valid URL", func(t *testing.T) {
		song := entity.Song{
			URL:                       "test-url",
			ArtistName:                "test-artist",
			Title:                     "SongByUrl with valid URL",
			CoverLink:                 "test-cover-link",
			CoverTelegramFileID:       "test-cover-telegram-file-id",
			SongTelegramMessageID:     123,
			SongTelegramMessageChatID: 123,
			Tags:                      []string{"test-tag1", "test-tag2"},
		}

		createdSong, err := elysiumRepo.CreateSong(context.Background(), song)
		assert.NoError(t, err)

		foundSong, err := elysiumRepo.SongByUrl(context.Background(), "test-url")
		assert.NoError(t, err)
		assert.Equal(t, createdSong.URL, foundSong.URL)

		err = elysiumRepo.RemoveSong(context.Background(), createdSong.ID)
		require.NoError(t, err)
	})

	t.Run("SongByUrl with non-existing URL", func(t *testing.T) {
		_, err := elysiumRepo.SongByUrl(context.Background(), "non-existing-url")
		require.Error(t, err)
	})

	t.Run("SongByUrl with empty URL", func(t *testing.T) {
		_, err := elysiumRepo.SongByUrl(context.Background(), "")
		require.Error(t, err)
	})
}

func TestLogSongDownloaded(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	err := elysiumRepo.LogSongDownload(context.Background(), 1, 1, entity.SongDownloadSourceBot)
	require.NoError(t, err)
}

func TestSongByID(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	song, err := elysiumRepo.SongByID(context.Background(), 241)
	require.NoError(t, err)
	t.Log(song.SongTelegramMessageID)
}
