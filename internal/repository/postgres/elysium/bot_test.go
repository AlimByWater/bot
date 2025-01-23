package elysium_test

import (
	"context"
	"elysium/internal/entity"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRepository_GetAllBots(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		testBots := []*entity.Bot{
			{ID: 1001, Name: "TestBot1", Token: "token1", Purpose: "test1", Test: true, Enabled: true},
			{ID: 1002, Name: "TestBot2", Token: "token2", Purpose: "test2", Test: true, Enabled: true},
		}

		for _, bot := range testBots {
			err := elysiumRepo.CreateBot(ctx, bot)
			require.NoError(t, err)
		}

		t.Cleanup(func() {
			for _, bot := range testBots {
				err := elysiumRepo.DeleteBotHard(ctx, bot.ID)
				require.NoError(t, err)
			}
		})

		bots, err := elysiumRepo.GetAllBots(ctx)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(bots), len(testBots))
	})

	t.Run("EmptyResult", func(t *testing.T) {
		// Предварительно удаляем всех тестовых ботов
		testBot := &entity.Bot{ID: 1003, Enabled: false}
		err := elysiumRepo.CreateBot(ctx, testBot)
		require.NoError(t, err)

		t.Cleanup(func() {
			err := elysiumRepo.DeleteBotHard(ctx, testBot.ID)
			require.NoError(t, err)
		})

		bots, err := elysiumRepo.GetAllBots(ctx)
		require.NoError(t, err)
		require.NotContains(t, bots, testBot)
	})
}

func TestRepository_GetBotByID(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		testBot := &entity.Bot{
			ID:      2001,
			Name:    "TestBot3",
			Token:   "token3",
			Purpose: "test3",
			Test:    true,
			Enabled: true,
		}

		err := elysiumRepo.CreateBot(ctx, testBot)
		require.NoError(t, err)

		t.Cleanup(func() {
			err := elysiumRepo.DeleteBotHard(ctx, testBot.ID)
			require.NoError(t, err)
		})

		bot, err := elysiumRepo.GetBotByID(ctx, testBot.ID)
		require.NoError(t, err)
		require.Equal(t, testBot.ID, bot.ID)
		require.Equal(t, testBot.Name, bot.Name)
	})

	t.Run("NotFound", func(t *testing.T) {
		_, err := elysiumRepo.GetBotByID(ctx, 9999)
		require.Error(t, err)
	})

	t.Run("DisabledBot", func(t *testing.T) {
		testBot := &entity.Bot{
			ID:      2002,
			Enabled: false,
		}

		err := elysiumRepo.CreateBot(ctx, testBot)
		require.NoError(t, err)

		t.Cleanup(func() {
			err := elysiumRepo.DeleteBotHard(ctx, testBot.ID)
			require.NoError(t, err)
		})

		_, err = elysiumRepo.GetBotByID(ctx, testBot.ID)
		require.Error(t, err)
	})
}

func TestRepository_CreateBot(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		testBot := &entity.Bot{
			ID:      3001,
			Name:    "TestBot4",
			Token:   "token4",
			Purpose: "test4",
			Test:    true,
			Enabled: true,
		}

		t.Cleanup(func() {
			err := elysiumRepo.DeleteBotHard(ctx, testBot.ID)
			require.NoError(t, err)
		})

		err := elysiumRepo.CreateBot(ctx, testBot)
		require.NoError(t, err)

		bot, err := elysiumRepo.GetBotByID(ctx, testBot.ID)
		require.NoError(t, err)
		require.Equal(t, testBot.ID, bot.ID)
	})

	t.Run("Update", func(t *testing.T) {
		testBot := &entity.Bot{
			ID:      3002,
			Name:    "TestBot5",
			Token:   "token5",
			Purpose: "test5",
			Test:    true,
			Enabled: true,
		}

		t.Cleanup(func() {
			err := elysiumRepo.DeleteBotHard(ctx, testBot.ID)
			require.NoError(t, err)
		})

		err := elysiumRepo.CreateBot(ctx, testBot)
		require.NoError(t, err)

		testBot.Name = "UpdatedBot5"
		err = elysiumRepo.CreateBot(ctx, testBot)
		require.NoError(t, err)

		bot, err := elysiumRepo.GetBotByID(ctx, testBot.ID)
		require.NoError(t, err)
		require.Equal(t, "UpdatedBot5", bot.Name)
	})
}

func TestRepository_UpdateBot(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		testBot := &entity.Bot{
			ID:      4001,
			Name:    "TestBot6",
			Token:   "token6",
			Purpose: "test6",
			Test:    true,
			Enabled: true,
		}

		err := elysiumRepo.CreateBot(ctx, testBot)
		require.NoError(t, err)

		t.Cleanup(func() {
			err := elysiumRepo.DeleteBotHard(ctx, testBot.ID)
			require.NoError(t, err)
		})

		testBot.Name = "UpdatedBot6"
		err = elysiumRepo.UpdateBot(ctx, testBot)
		require.NoError(t, err)

		bot, err := elysiumRepo.GetBotByID(ctx, testBot.ID)
		require.NoError(t, err)
		require.Equal(t, "UpdatedBot6", bot.Name)
	})

	t.Run("NotFound", func(t *testing.T) {
		testBot := &entity.Bot{ID: 9999}
		err := elysiumRepo.UpdateBot(ctx, testBot)
		require.Error(t, err)
	})
}

func TestRepository_SetUserToBotActive(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		// Создаем тестового пользователя
		testUser := entity.User{
			TelegramID:       1001,
			TelegramUsername: "testuser1",
			Firstname:        "Test",
		}
		user, err := elysiumRepo.CreateOrUpdateUser(ctx, testUser)
		require.NoError(t, err)
		require.NotZero(t, user.ID)

		// Создаем тестового бота
		testBot := &entity.Bot{
			ID:      2001,
			Name:    "TestBot",
			Token:   "token",
			Purpose: "test",
			Test:    true,
			Enabled: true,
		}
		err = elysiumRepo.CreateBot(ctx, testBot)
		require.NoError(t, err)

		t.Cleanup(func() {
			err := elysiumRepo.DeleteBotHard(ctx, testBot.ID)
			require.NoError(t, err)
			err = elysiumRepo.DeleteUser(ctx, user.ID)
			require.NoError(t, err)
		})

		// Активируем бота для пользователя
		err = elysiumRepo.SetUserToBotActive(ctx, user.ID, testBot.ID)
		require.NoError(t, err)

		// Проверяем что бот появился в списке активных
		activeBots, err := elysiumRepo.GetUserActiveBots(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, activeBots, 1)
		require.Equal(t, testBot.ID, activeBots[0].ID)
	})

	t.Run("UpdateExisting", func(t *testing.T) {
		// Создаем тестового пользователя
		testUser := entity.User{
			TelegramID:       1002,
			TelegramUsername: "testuser2",
			Firstname:        "Test2",
		}
		user, err := elysiumRepo.CreateOrUpdateUser(ctx, testUser)
		require.NoError(t, err)
		require.NotZero(t, user.ID)

		// Создаем тестового бота
		testBot := &entity.Bot{
			ID:      2002,
			Name:    "TestBot2",
			Token:   "token2",
			Purpose: "test2",
			Test:    true,
			Enabled: true,
		}
		err = elysiumRepo.CreateBot(ctx, testBot)
		require.NoError(t, err)

		t.Cleanup(func() {
			err := elysiumRepo.DeleteBotHard(ctx, testBot.ID)
			require.NoError(t, err)
			err = elysiumRepo.DeleteUser(ctx, user.ID)
			require.NoError(t, err)
		})

		// Деактивируем бота для пользователя
		err = elysiumRepo.UnsetUserToBot(ctx, user.ID, testBot.ID)
		require.NoError(t, err)

		// Проверяем что бот не активен
		activeBots, err := elysiumRepo.GetUserActiveBots(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, activeBots, 0)

		// Активируем бота
		err = elysiumRepo.SetUserToBotActive(ctx, user.ID, testBot.ID)
		require.NoError(t, err)

		// Проверяем что бот стал активным
		activeBots, err = elysiumRepo.GetUserActiveBots(ctx, user.ID)
		require.NoError(t, err)
		require.Len(t, activeBots, 1)
		require.Equal(t, testBot.ID, activeBots[0].ID)
	})
}

func TestRepository_DeleteBot(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		testBot := &entity.Bot{
			ID:      5001,
			Name:    "TestBot7",
			Token:   "token7",
			Purpose: "test7",
			Test:    true,
			Enabled: true,
		}

		err := elysiumRepo.CreateBot(ctx, testBot)
		require.NoError(t, err)

		t.Cleanup(func() {
			err := elysiumRepo.DeleteBotHard(ctx, testBot.ID)
			require.NoError(t, err)
		})

		err = elysiumRepo.DeleteBot(ctx, testBot.ID)
		require.NoError(t, err)

		_, err = elysiumRepo.GetBotByID(ctx, testBot.ID)
		require.Error(t, err)
	})

	t.Run("NotFound", func(t *testing.T) {
		err := elysiumRepo.DeleteBot(ctx, 9999)
		require.Error(t, err)
	})

	t.Run("AlreadyDeleted", func(t *testing.T) {
		testBot := &entity.Bot{
			ID:      5002,
			Enabled: false,
		}

		err := elysiumRepo.CreateBot(ctx, testBot)
		require.NoError(t, err)

		t.Cleanup(func() {
			err := elysiumRepo.DeleteBotHard(ctx, testBot.ID)
			require.NoError(t, err)
		})

		err = elysiumRepo.DeleteBot(ctx, testBot.ID)
		require.Error(t, err)
	})
}
