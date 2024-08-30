package redis_test

import (
	"arimadj-helper/internal/entity"
	redisRepo "arimadj-helper/internal/repository/redis"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

var redisModule *redisRepo.Module

func TestSaveLayoutSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	layout := entity.UserLayout{
		UserID:   "user123",
		LayoutID: "layout123",
		Background: entity.Background{
			Type:  "color",
			Value: "#FFFFFF",
		},
		Layout:  []entity.LayoutElement{},
		Creator: 12345,
		Editors: []int{12345},
	}

	err := redisModule.SaveLayout(context.Background(), layout)
	require.NoError(t, err)
}

func TestGetLayoutSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	// Preparing a layout to be retrieved
	layout := entity.UserLayout{
		UserID:   "user123",
		LayoutID: "layout123",
		Background: entity.Background{
			Type:  "image",
			Value: "https://example.com/image.jpg",
		},
		Layout:  []entity.LayoutElement{},
		Creator: 12345,
		Editors: []int{12345},
	}
	err := redisModule.SaveLayout(context.Background(), layout)
	require.NoError(t, err)

	// Retrieving the layout
	retrievedLayout, err := redisModule.GetLayout(context.Background(), "user123", "layout123")
	require.NoError(t, err)
	require.Equal(t, layout.Background.Type, retrievedLayout.Background.Type)
	require.Equal(t, layout.Background.Value, retrievedLayout.Background.Value)
}

func TestGetLayoutReturnsErrorWhenNoLayout(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	// Attempting to retrieve a layout when none exists
	_, err := redisModule.GetLayout(context.Background(), "nonexistent_user", "nonexistent_layout")
	require.Error(t, err)
	require.Equal(t, redisRepo.ErrLayoutNotFound, err)
}

func TestSaveLayoutUpdatesExistingLayout(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	initialLayout := entity.UserLayout{
		UserID:   "user123",
		LayoutID: "layout123",
		Background: entity.Background{
			Type:  "color",
			Value: "#000000",
		},
		Layout:  []entity.LayoutElement{},
		Creator: 12345,
		Editors: []int{12345},
	}

	err := redisModule.SaveLayout(context.Background(), initialLayout)
	require.NoError(t, err)

	updatedLayout := entity.UserLayout{
		UserID:   "user123",
		LayoutID: "layout123",
		Background: entity.Background{
			Type:  "image",
			Value: "https://example.com/new_image.jpg",
		},
		Layout:  []entity.LayoutElement{},
		Creator: 12345,
		Editors: []int{12345, 67890},
	}

	err = redisModule.SaveLayout(context.Background(), updatedLayout)
	require.NoError(t, err)

	retrievedLayout, err := redisModule.GetLayout(context.Background(), "user123", "layout123")
	require.NoError(t, err)
	require.Equal(t, updatedLayout.Background.Type, retrievedLayout.Background.Type)
	require.Equal(t, updatedLayout.Background.Value, retrievedLayout.Background.Value)
	require.Equal(t, updatedLayout.Editors, retrievedLayout.Editors)
}
