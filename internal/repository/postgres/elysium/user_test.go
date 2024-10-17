package elysium_test

import (
	"context"
	"elysium/internal/entity"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCreateUserSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	user := entity.User{
		TelegramID:       723456789,
		TelegramUsername: "testuser",
		Firstname:        "Test",
		DateCreate:       time.Now(),
	}

	createdUser, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user)
	require.NoError(t, err)
	require.NotZero(t, createdUser.ID)
	require.Equal(t, user.TelegramUsername, createdUser.TelegramUsername)

	// remove user
	err = elysiumRepo.DeleteUser(context.Background(), createdUser.ID)
	require.NoError(t, err)
}

func TestCreateThenUpdateUser(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	user := entity.User{
		TelegramID:       723456789,
		TelegramUsername: "testuser",
		Firstname:        "Test",
		DateCreate:       time.Now(),
	}

	_, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user)
	require.NoError(t, err)

	user.TelegramUsername = "testuser2"
	// Attempt to create a second user with the same TelegramID
	updatedUser, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user)
	require.NoError(t, err)
	require.Equal(t, user.TelegramUsername, updatedUser.TelegramUsername)

	err = elysiumRepo.DeleteUser(context.Background(), user.ID)
	require.NoError(t, err)
}

func TestCreateUserFailsOnMissingRequiredFields(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	user := entity.User{
		// Missing TelegramID and other required fields
	}

	_, err := elysiumRepo.CreateOrUpdateUser(context.Background(), user)
	require.Error(t, err)
}

func TestGetUserByTelegramID(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	ids := []int64{86700706, 251636949}
	users, err := elysiumRepo.GetUsersByTelegramID(context.Background(), ids)
	require.NoError(t, err)
	require.Len(t, users, 2)
	t.Log(users)

}
