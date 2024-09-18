package redis_test

import (
	"context"
	"elysium/internal/entity"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSaveOrUpdateLayoutSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	layout := entity.UserLayout{
		ID: 123,
		Background: entity.Background{
			"type":  "color",
			"value": "#FFFFFF",
		},
		Elements: []entity.LayoutElement{
			{
				RootElement: entity.RootElement{
					ID: 1,
				},
				Properties: map[string]interface{}{
					"size": "large",
				},
				Position: entity.Position{X: 10, Y: 20},
			},
			{
				RootElement: entity.RootElement{
					ID: 2,
				},
				Properties: map[string]interface{}{
					"size": "small",
				},
				Position: entity.Position{X: 30, Y: 40},
			},
		},
		Creator: 12345,
		Editors: []int{12345},
	}

	err := redisModule.SaveOrUpdateLayout(context.Background(), layout)
	require.NoError(t, err)
}

func TestGetLayoutSucceeds(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	// Preparing a layout to be retrieved
	layout := entity.UserLayout{
		ID:   1234,
		Name: "default_layout_4",
		Background: entity.Background{
			"type":  "image",
			"value": "https://example.com/image.jpg",
		},
		Elements: []entity.LayoutElement{},
		Creator:  12345,
		Editors:  []int{12345},
	}
	err := redisModule.SaveOrUpdateLayout(context.Background(), layout)
	require.NoError(t, err)

	// Retrieving the layout
	retrievedLayout, err := redisModule.GetLayout(context.Background(), layout.ID)
	require.NoError(t, err)
	require.Equal(t, layout.Background["type"], retrievedLayout.Background["type"])
	require.Equal(t, layout.Background["value"], retrievedLayout.Background["value"])

	retrievedLayout, err = redisModule.GetLayoutByName(context.Background(), layout.Name)
	require.NoError(t, err)
	require.Equal(t, layout.Background["type"], retrievedLayout.Background["type"])
	require.Equal(t, layout.Background["value"], retrievedLayout.Background["value"])
}

func TestGetLayoutReturnsErrorWhenNoLayout(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	// Attempting to retrieve a layout when none exists
	_, err := redisModule.GetLayout(context.Background(), -1)
	require.Error(t, err)
	require.Equal(t, entity.ErrLayoutNotFound, err)
}

func TestSaveOrUpdateLayoutUpdatesExistingLayout(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	initialLayout := entity.UserLayout{
		ID: 123,
		Background: entity.Background{
			"type":  "color",
			"value": "#000000",
		},
		Elements: []entity.LayoutElement{},
		Creator:  12345,
		Editors:  []int{12345},
	}

	err := redisModule.SaveOrUpdateLayout(context.Background(), initialLayout)
	require.NoError(t, err)

	updatedLayout := entity.UserLayout{
		ID: 123,
		Background: entity.Background{
			"type":  "image",
			"value": "https://example.com/new_image.jpg",
		},
		Elements: []entity.LayoutElement{},
		Creator:  12345,
		Editors:  []int{12345, 67890},
	}

	err = redisModule.SaveOrUpdateLayout(context.Background(), updatedLayout)
	require.NoError(t, err)

	retrievedLayout, err := redisModule.GetLayout(context.Background(), updatedLayout.ID)
	require.NoError(t, err)
	require.Equal(t, updatedLayout.Background["type"], retrievedLayout.Background["type"])
	require.Equal(t, updatedLayout.Background["value"], retrievedLayout.Background["value"])
	require.Equal(t, updatedLayout.Editors, retrievedLayout.Editors)

}

//
//func TestGetLayoutSucceeds(t *testing.T) {
//	teardown := setupTest(t)
//	defer teardown(t)
//
//	layout := entity.UserLayout{
//		UserID:   "user123",
//		LayoutID: "layout123",
//		Background: entity.Background{
//			Type:  "color",
//			Value: "#FFFFFF",
//		},
//		Layout:  []entity.LayoutElement{},
//		Creator: 12345,
//		Editors: []int{12345},
//	}
//
//	err := redisModule.SaveOrUpdateLayout(context.Background(), layout)
//	require.NoError(t, err)
//
//	retrievedLayout, err := redisModule.GetLayout(context.Background(), layout.LayoutID)
//	require.NoError(t, err)
//	require.Equal(t, layout.UserID, retrievedLayout.UserID)
//	require.Equal(t, layout.LayoutID, retrievedLayout.LayoutID)
//	require.Equal(t, layout.Background.Type, retrievedLayout.Background.Type)
//	require.Equal(t, layout.Background.Value, retrievedLayout.Background.Value)
//	require.Equal(t, layout.Creator, retrievedLayout.Creator)
//	require.Equal(t, layout.Editors, retrievedLayout.Editors)
//}
