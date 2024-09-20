package elysium_test

import (
	"context"
	"elysium/internal/entity"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateLayout(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	t.Run("CreateLayout with valid data", func(t *testing.T) {
		layout := entity.UserLayout{
			Name:    "Test Layout",
			Creator: 2, // Предполагаемый ID пользователя
			Background: map[string]interface{}{
				"color": "#FFFFFF",
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
			},
			Editors: []int{2},
		}

		err := elysiumRepo.CreateLayout(context.Background(), layout)
		require.NoError(t, err)

		// Проверяем, что макет был создан
		createdLayout, err := elysiumRepo.LayoutByUserID(context.Background(), layout.Creator)
		require.NoError(t, err)
		assert.Equal(t, layout.Name, createdLayout.Name)
		assert.Equal(t, layout.Creator, createdLayout.Creator)
		assert.Equal(t, layout.Background, createdLayout.Background)
		assert.Len(t, createdLayout.Elements, 1)
		assert.Equal(t, layout.Elements[0].RootElement.ID, createdLayout.Elements[0].RootElement.ID)
		assert.Equal(t, layout.Elements[0].Properties, createdLayout.Elements[0].Properties)
		assert.Equal(t, layout.Elements[0].Position, createdLayout.Elements[0].Position)
		assert.Equal(t, layout.Editors, createdLayout.Editors)

		// Очистка
		err = elysiumRepo.DeleteLayout(context.Background(), createdLayout.ID)
		require.NoError(t, err)
	})

	t.Run("CreateLayout with missing required fields", func(t *testing.T) {
		layout := entity.UserLayout{
			// Отсутствует поле Name
			Creator: 1,
		}

		err := elysiumRepo.CreateLayout(context.Background(), layout)
		require.Error(t, err)
	})

	t.Run("CreateLayout with duplicate name", func(t *testing.T) {
		layout1 := entity.UserLayout{
			Name:    "Duplicate Layout",
			Creator: 1,
		}

		err := elysiumRepo.CreateLayout(context.Background(), layout1)
		require.NoError(t, err)

		layout2 := entity.UserLayout{
			Name:    "Duplicate Layout",
			Creator: 1,
		}

		err = elysiumRepo.CreateLayout(context.Background(), layout2)
		require.Error(t, err)

		// Очистка
		createdLayout, _ := elysiumRepo.LayoutByUserID(context.Background(), layout1.Creator)
		_ = elysiumRepo.DeleteLayout(context.Background(), createdLayout.ID)
	})
}

func TestLayoutByUserID(t *testing.T) {
	teardown := setupTest(t)
	defer teardown(t)

	t.Run("LayoutByUserID with existing layout", func(t *testing.T) {
		layout := entity.UserLayout{
			Name:    "Test User Layout",
			Creator: 1,
		}

		err := elysiumRepo.CreateLayout(context.Background(), layout)
		require.NoError(t, err)

		foundLayout, err := elysiumRepo.LayoutByUserID(context.Background(), layout.Creator)
		assert.NoError(t, err)
		assert.Equal(t, layout.Name, foundLayout.Name)
		assert.Equal(t, layout.Creator, foundLayout.Creator)

		// Очистка
		err = elysiumRepo.DeleteLayout(context.Background(), foundLayout.ID)
		require.NoError(t, err)
	})

	t.Run("LayoutByUserID with non-existing user", func(t *testing.T) {
		_, err := elysiumRepo.LayoutByUserID(context.Background(), 999999)
		require.Error(t, err)
		assert.Equal(t, entity.ErrLayoutNotFound, err)
	})
}
