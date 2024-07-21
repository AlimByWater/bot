package elysium_test

import (
	"arimadj-helper/internal/entity"
	"context"
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
