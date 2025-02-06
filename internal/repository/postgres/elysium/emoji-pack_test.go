package elysium_test

import (
	"context"
	"elysium/internal/application/logger"
	"elysium/internal/controller/telegram/httpcaller"
	"elysium/internal/entity"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mymmrac/telego"
	"golang.org/x/time/rate"

	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, ctx context.Context, telegramID int64) entity.User {
	t.Helper()
	user := entity.User{
		TelegramID:       telegramID,
		TelegramUsername: fmt.Sprintf("testuser_%d", telegramID),
		Firstname:        "Test User",
		DateCreate:       time.Now(),
	}

	createdUser, err := elysiumRepo.CreateOrUpdateUser(ctx, user)
	require.NoError(t, err)
	return createdUser
}

func TestCreateNewEmojiPack(t *testing.T) {
	t.Skip()

	ctx := context.Background()

	t.Run("Create new emoji pack", func(t *testing.T) {
		// Создаем пользователя
		creatorID := int64(1000001)
		user := createTestUser(t, ctx, creatorID)

		pack := entity.EmojiPack{
			Bot:               entity.Bot{ID: -1007894673045},
			PackLink:          uuid.New().String(),
			CreatorTelegramID: creatorID,
			EmojiCount:        5,
			Deleted:           false,
			CreatedAt:         time.Now(),
		}

		createdPack, err := elysiumRepo.CreateNewEmojiPack(ctx, pack)
		require.NoError(t, err)
		require.NotZero(t, createdPack.ID)

		t.Cleanup(func() {
			// Удаляем пак
			query := `DELETE FROM emoji_packs WHERE id = $1`
			_, err := elysiumRepo.DB().ExecContext(ctx, query, createdPack.ID)
			require.NoError(t, err)

			// Удаляем пользователя
			err = elysiumRepo.DeleteUser(ctx, user.ID)
			require.NoError(t, err)
		})
	})
}

func TestGetEmojiPackByPackLink(t *testing.T) {
	t.Skip()

	ctx := context.Background()

	t.Run("Get emoji pack by pack link", func(t *testing.T) {
		// Создаем пользователя
		creatorID := int64(1000002)
		user := createTestUser(t, ctx, creatorID)

		pack := entity.EmojiPack{
			Bot:               entity.Bot{ID: -1007894673045},
			PackLink:          uuid.New().String(),
			CreatorTelegramID: creatorID,
			EmojiCount:        5,
			Deleted:           false,
			CreatedAt:         time.Now(),
		}

		createdPack, err := elysiumRepo.CreateNewEmojiPack(ctx, pack)
		require.NoError(t, err)

		fetchedPack, err := elysiumRepo.GetEmojiPackByPackLink(ctx, pack.PackLink)
		require.NoError(t, err)
		require.Equal(t, createdPack.ID, fetchedPack.ID)

		t.Cleanup(func() {
			query := `DELETE FROM emoji_packs WHERE id = $1`
			_, err := elysiumRepo.DB().ExecContext(ctx, query, createdPack.ID)
			require.NoError(t, err)

			err = elysiumRepo.DeleteUser(ctx, user.ID)
			require.NoError(t, err)
		})
	})
}

func TestSetUnsetEmojiPackDeleted(t *testing.T) {
	t.Skip()

	ctx := context.Background()

	t.Run("Set emoji pack as deleted", func(t *testing.T) {
		// Создаем пользователя
		creatorID := int64(1000003)
		user := createTestUser(t, ctx, creatorID)

		pack := entity.EmojiPack{
			Bot:               entity.Bot{ID: -1007894673045},
			PackLink:          uuid.New().String(),
			CreatorTelegramID: creatorID,
			EmojiCount:        5,
			Deleted:           false,
			CreatedAt:         time.Now(),
		}

		createdPack, err := elysiumRepo.CreateNewEmojiPack(ctx, pack)
		require.NoError(t, err)

		err = elysiumRepo.SetEmojiPackDeleted(ctx, pack.PackLink)
		require.NoError(t, err)

		updatedPack, err := elysiumRepo.GetEmojiPackByPackLink(ctx, pack.PackLink)
		require.NoError(t, err)
		require.True(t, updatedPack.Deleted)

		err = elysiumRepo.UnsetEmojiPackDeleted(ctx, pack.PackLink)
		require.NoError(t, err)

		updatedPack, err = elysiumRepo.GetEmojiPackByPackLink(ctx, pack.PackLink)
		require.NoError(t, err)
		require.False(t, updatedPack.Deleted)

		t.Cleanup(func() {
			query := `DELETE FROM emoji_packs WHERE id = $1`
			_, err := elysiumRepo.DB().ExecContext(ctx, query, createdPack.ID)
			require.NoError(t, err)

			err = elysiumRepo.DeleteUser(ctx, user.ID)
			require.NoError(t, err)
		})
	})
}

func TestUpdateEmojiCount(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	ctx := context.Background()

	t.Run("Update emoji count", func(t *testing.T) {
		// Создаем пользователя
		creatorID := int64(1000004)
		user := createTestUser(t, ctx, creatorID)

		pack := entity.EmojiPack{
			Bot:               entity.Bot{ID: -1007894673045},
			PackLink:          uuid.New().String(),
			CreatorTelegramID: creatorID,
			EmojiCount:        5,
			Deleted:           false,
			CreatedAt:         time.Now(),
		}

		createdPack, err := elysiumRepo.CreateNewEmojiPack(ctx, pack)
		require.NoError(t, err)

		err = elysiumRepo.UpdateEmojiCount(ctx, createdPack.ID, 10)
		require.NoError(t, err)

		updatedPack, err := elysiumRepo.GetEmojiPackByPackLink(ctx, pack.PackLink)
		require.NoError(t, err)
		require.Equal(t, 10, updatedPack.EmojiCount)

		t.Cleanup(func() {
			query := `DELETE FROM emoji_packs WHERE id = $1`
			_, err := elysiumRepo.DB().ExecContext(ctx, query, createdPack.ID)
			require.NoError(t, err)

			err = elysiumRepo.DeleteUser(ctx, user.ID)
			require.NoError(t, err)
		})
	})
}

func TestGetEmojiPacksByCreator(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	ctx := context.Background()

	t.Run("Get emoji packs by creator", func(t *testing.T) {
		// Создаем пользователя
		creatorID := int64(1000005)
		user := createTestUser(t, ctx, creatorID)

		pack1 := entity.EmojiPack{
			Bot:               entity.Bot{ID: -1007894673045},
			PackLink:          uuid.New().String(),
			CreatorTelegramID: creatorID,
			EmojiCount:        5,
			Deleted:           false,
			CreatedAt:         time.Now(),
		}
		pack2 := entity.EmojiPack{
			Bot:               entity.Bot{ID: -1007894673045},
			PackLink:          uuid.New().String(),
			CreatorTelegramID: creatorID,
			EmojiCount:        3,
			Deleted:           true,
			CreatedAt:         time.Now(),
		}

		createdPack1, err := elysiumRepo.CreateNewEmojiPack(ctx, pack1)
		require.NoError(t, err)
		createdPack2, err := elysiumRepo.CreateNewEmojiPack(ctx, pack2)
		require.NoError(t, err)

		packs, err := elysiumRepo.GetEmojiPacksByCreator(ctx, creatorID, false)
		require.NoError(t, err)
		require.Len(t, packs, 1)
		require.Equal(t, createdPack1.ID, packs[0].ID)

		packs, err = elysiumRepo.GetEmojiPacksByCreator(ctx, creatorID, true)
		require.NoError(t, err)
		require.Len(t, packs, 1)
		require.Equal(t, createdPack2.ID, packs[0].ID)

		t.Cleanup(func() {
			query := `DELETE FROM emoji_packs WHERE id IN ($1, $2)`
			_, err := elysiumRepo.DB().ExecContext(ctx, query, createdPack1.ID, createdPack2.ID)
			require.NoError(t, err)

			err = elysiumRepo.DeleteUser(ctx, user.ID)
			require.NoError(t, err)
		})
	})
}

func TestDeleteEmojiPacks(t *testing.T) {
	ctx := context.TODO()

	creatorID := int64(251636949)

	loggerModule := logger.New(
		logger.Options{
			AppName: "driptech-bot-test",
			Writer:  os.Stdout,
			HandlerOptions: &slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		},
	)

	bots, err := elysiumRepo.GetAllBots(ctx)
	require.NoError(t, err)

	packs1, err := elysiumRepo.GetEmojiPacksByCreator(ctx, creatorID, false)
	require.NoError(t, err)

	packsByBotID := make(map[string][]entity.EmojiPack)
	for _, pack := range packs1 {
		packsByBotID[fmt.Sprintf("%d", pack.BotID)] = append(packsByBotID[fmt.Sprintf("%d", pack.BotID)], pack)
	}

	for _, b := range bots {
		rl := rate.NewLimiter(rate.Every(1*time.Second), 80)
		bot, err := telego.NewBot(b.Token,
			telego.WithAPICaller(httpcaller.NewFastHttpCallerWithLimiter(rl, loggerModule)),
		)
		if err != nil {
			t.Log("Newbot", b.Name)
		}

		// Get bot info for middleware
		me, err := bot.GetMe()
		if err != nil {
			continue
		}

		packs, ok := packsByBotID[fmt.Sprintf("-100%d", me.ID)]
		if !ok {
			continue
		}

		for _, pack := range packs {
			err = bot.DeleteStickerSet(&telego.DeleteStickerSetParams{Name: pack.PackLink})
			if err != nil {
				t.Log("DeleteStickerSet", pack.ID)
			} else {
				err2 := elysiumRepo.DeleteEmojiPackHard(ctx, pack.ID)
				if err2 != nil {
					t.Log("DeleteEmojiPackHard", pack.ID)
				}
			}
		}
	}

}
