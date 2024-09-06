package entity

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserLayout_MarshalUnmarshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	layout := UserLayout{
		ID:   2,
		Name: "Test Layout",
		Background: Background{
			"color": "#FFFFFF",
			"image": "https://example.com/image.jpg",
		},
		Elements: []LayoutElement{
			{
				ID:            1,
				RootElementID: 100,
				Position: Position{
					X:      10,
					Y:      20,
					Width:  300,
					Height: 200,
					Z:      1,
				},
				Properties: Properites{
					"text":  "Hello, World!",
					"color": "#000000",
				},
				IsPublic:    true,
				IsRemovable: false,
			},
		},
		Creator:   123,
		Editors:   []int{456, 789},
		CreatedAt: now,
		UpdatedAt: now,
	}

	t.Run("Standard JSON Marshal/Unmarshal", func(t *testing.T) {
		data, err := json.Marshal(layout)
		require.NoError(t, err)
		var unmarshaled UserLayout
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, layout, unmarshaled)
	})

	t.Run("EasyJSON Marshal/Unmarshal", func(t *testing.T) {
		data, err := layout.MarshalJSON()
		require.NoError(t, err)

		fmt.Println(string(data))
		var unmarshaled UserLayout
		err = unmarshaled.UnmarshalJSON(data)
		require.NoError(t, err)

		assert.Equal(t, layout, unmarshaled)
	})
}

func TestUserLayout_MarshalUnmarshal_Empty(t *testing.T) {
	layout := UserLayout{}

	t.Run("Standard JSON Marshal/Unmarshal Empty", func(t *testing.T) {
		data, err := json.Marshal(layout)
		require.NoError(t, err)

		var unmarshaled UserLayout
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, layout, unmarshaled)
	})

	t.Run("EasyJSON Marshal/Unmarshal Empty", func(t *testing.T) {
		data, err := layout.MarshalJSON()
		require.NoError(t, err)

		var unmarshaled UserLayout
		err = unmarshaled.UnmarshalJSON(data)
		require.NoError(t, err)

		assert.Equal(t, layout, unmarshaled)
	})
}

func TestUserLayout_MarshalUnmarshal_NullFields(t *testing.T) {
	layout := UserLayout{
		ID:         2,
		Background: nil,
		Elements:   nil,
		Editors:    nil,
	}

	t.Run("Standard JSON Marshal/Unmarshal Null Fields", func(t *testing.T) {
		data, err := json.Marshal(layout)
		require.NoError(t, err)

		var unmarshaled UserLayout
		err = json.Unmarshal(data, &unmarshaled)
		require.NoError(t, err)

		assert.Equal(t, layout, unmarshaled)
	})

	t.Run("EasyJSON Marshal/Unmarshal Null Fields", func(t *testing.T) {
		data, err := layout.MarshalJSON()
		require.NoError(t, err)

		var unmarshaled UserLayout
		err = unmarshaled.UnmarshalJSON(data)
		require.NoError(t, err)

		assert.Equal(t, layout, unmarshaled)
	})
}
