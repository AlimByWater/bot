package elysium_test

import (
	"context"
	"elysium/internal/entity"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCreateOrUpdateUser(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	testCases := []struct {
		name        string
		user        entity.User
		expectError bool
	}{
		{
			name: "Create new user",
			user: entity.User{
				TelegramID:       9000000001,
				TelegramUsername: "testuser1",
				Firstname:        "Test",
				DateCreate:       time.Now(),
			},
			expectError: false,
		},
		{
			name: "Update existing user",
			user: entity.User{
				TelegramID:       9000000001,
				TelegramUsername: "updateduser1",
				Firstname:        "Updated",
				DateCreate:       time.Now(),
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			createdUser, err := elysiumRepo.CreateOrUpdateUser(context.Background(), tc.user)
			if tc.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotZero(t, createdUser.ID)

			t.Cleanup(func() {
				err := elysiumRepo.DeleteUser(context.Background(), createdUser.ID)
				require.NoError(t, err)
			})

			fetchedUser, err := elysiumRepo.GetUserByID(context.Background(), createdUser.ID)
			require.NoError(t, err)
			require.Equal(t, tc.user.TelegramUsername, fetchedUser.TelegramUsername)
			require.Equal(t, tc.user.Firstname, fetchedUser.Firstname)
		})
	}
}

func TestGetUserByTelegramID(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	testCases := []struct {
		name        string
		telegramID  int64
		shouldExist bool
	}{
		{
			name:        "Existing user",
			telegramID:  9000000002,
			shouldExist: true,
		},
		{
			name:        "Non-existing user",
			telegramID:  9000000999,
			shouldExist: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.shouldExist {
				user := entity.User{
					TelegramID:       tc.telegramID,
					TelegramUsername: "testuser2",
					Firstname:        "Test",
					DateCreate:       time.Now(),
				}
				createdUser, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user)
				require.NoError(t, err)
				t.Cleanup(func() {
					err := elysiumRepo.DeleteUser(context.Background(), createdUser.ID)
					require.NoError(t, err)
				})
			}

			fetchedUser, err := elysiumRepo.GetUserByTelegramID(context.Background(), tc.telegramID)
			if tc.shouldExist {
				require.NoError(t, err)
				require.Equal(t, tc.telegramID, fetchedUser.TelegramID)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	cleanup := setupTest(t)
	defer cleanup(t)

	ctx := context.Background()

	// Сначала создаем пользователя
	testUser := entity.User{
		TelegramID:       123456789,
		TelegramUsername: "testuser",
		Firstname:        "Test User",
		DateCreate:       time.Now(),
		Balance:          100,
		Permissions: entity.Permissions{
			PrivateGeneration: true,
			UseByChannelName:  true,
			Vip:               false,
		},
	}

	createdUser, err := elysiumRepo.CreateOrUpdateUser(ctx, testUser)
	require.NoError(t, err)
	require.NotZero(t, createdUser.ID)

	// Теперь пытаемся получить пользователя
	fetchedUser, err := elysiumRepo.GetUserByID(ctx, createdUser.ID)
	require.NoError(t, err)
	require.Equal(t, createdUser.ID, fetchedUser.ID)
	require.Equal(t, testUser.TelegramID, fetchedUser.TelegramID)
	require.Equal(t, testUser.TelegramUsername, fetchedUser.TelegramUsername)

	// Очистка данных
	err = elysiumRepo.DeleteUser(ctx, createdUser.ID)
	require.NoError(t, err)
}

func TestUpdatePermissions(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	user := entity.User{
		TelegramID:       9000000004,
		TelegramUsername: "testuser4",
		Firstname:        "Test",
		DateCreate:       time.Now(),
	}
	createdUser, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := elysiumRepo.DeleteUser(context.Background(), createdUser.ID)
		require.NoError(t, err)
	})

	newPermissions := entity.Permissions{
		PrivateGeneration: true,
		UseByChannelName:  true,
		Vip:               true,
	}

	err = elysiumRepo.UpdatePermissions(context.Background(), createdUser.ID, newPermissions)
	require.NoError(t, err)

	updatedUser, err := elysiumRepo.GetUserByID(context.Background(), createdUser.ID)
	require.NoError(t, err)
	require.Equal(t, newPermissions.PrivateGeneration, updatedUser.Permissions.PrivateGeneration)
	require.Equal(t, newPermissions.UseByChannelName, updatedUser.Permissions.UseByChannelName)
	require.Equal(t, newPermissions.Vip, updatedUser.Permissions.Vip)
}

func TestDeleteUser(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	user := entity.User{
		TelegramID:       9000000005,
		TelegramUsername: "testuser5",
		Firstname:        "Test",
		DateCreate:       time.Now(),
	}
	createdUser, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user)
	require.NoError(t, err)

	err = elysiumRepo.DeleteUser(context.Background(), createdUser.ID)
	require.NoError(t, err)

	_, err = elysiumRepo.GetUserByID(context.Background(), createdUser.ID)
	require.Error(t, err)
}

func TestGetUsersByTelegramID(t *testing.T) {
	t.Skip()
	//teardown := setupTest(t)
	//defer teardown(t)

	user1 := entity.User{
		TelegramID:       9000000006,
		TelegramUsername: "testuser6",
		Firstname:        "Test",
		DateCreate:       time.Now(),
	}
	user2 := entity.User{
		TelegramID:       9000000007,
		TelegramUsername: "testuser7",
		Firstname:        "Test",
		DateCreate:       time.Now(),
	}

	createdUser1, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user1)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := elysiumRepo.DeleteUser(context.Background(), createdUser1.ID)
		require.NoError(t, err)
	})

	createdUser2, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user2)
	require.NoError(t, err)
	t.Cleanup(func() {
		err := elysiumRepo.DeleteUser(context.Background(), createdUser2.ID)
		require.NoError(t, err)
	})

	telegramIDs := []int64{user1.TelegramID, user2.TelegramID}

	users, err := elysiumRepo.GetUsersByTelegramID(context.Background(), telegramIDs)
	require.NoError(t, err)
	require.Len(t, users, 2)

	receivedIDs := []int64{users[0].TelegramID, users[1].TelegramID}
	require.ElementsMatch(t, telegramIDs, receivedIDs)
}
