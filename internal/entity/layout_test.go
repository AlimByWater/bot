package entity

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fastjson"
	"testing"
)

func TestMarshalFastJSONWithValidData(t *testing.T) {
	ul := &UserLayout{
		UserID:   "user123",
		LayoutID: "layout123",
		Background: Background{
			Type:  "color",
			Value: "#FFFFFF",
		},
		Layout: []LayoutElement{
			{
				ElementID: "elem1",
				Type:      "clickable_navigable",
				Position: Position{
					Row:    1,
					Column: 1,
					Height: 1,
					Width:  1,
				},
				Properties: Properties{
					Icon:          "path/to/icon1.png",
					Title:         "Navigation Item",
					NavigationURL: "/some-page",
				},
				Public:    true,
				Removable: true,
			},
		},
		Creator: 123,
		Editors: []int{123, 456},
	}

	a := fastjson.Arena{}
	v := ul.MarshalFastJSON(&a)

	require.Equal(t, "user123", string(v.GetStringBytes("userId")))
	require.Equal(t, "layout123", string(v.GetStringBytes("layoutId")))
	require.Equal(t, "color", string(v.Get("background").GetStringBytes("type")))
	require.Equal(t, "#FFFFFF", string(v.Get("background").GetStringBytes("value")))
	require.Equal(t, "elem1", string(v.GetArray("layout")[0].GetStringBytes("elementId")))
	require.Equal(t, "clickable_navigable", string(v.GetArray("layout")[0].GetStringBytes("type")))
	require.Equal(t, 1, v.GetArray("layout")[0].Get("position").GetInt("row"))
	require.Equal(t, 1, v.GetArray("layout")[0].Get("position").GetInt("column"))
	require.Equal(t, 1, v.GetArray("layout")[0].Get("position").GetInt("height"))
	require.Equal(t, 1, v.GetArray("layout")[0].Get("position").GetInt("width"))
	require.Equal(t, "path/to/icon1.png", string(v.GetArray("layout")[0].Get("properties").GetStringBytes("icon")))
	require.Equal(t, "Navigation Item", string(v.GetArray("layout")[0].Get("properties").GetStringBytes("title")))
	require.Equal(t, "/some-page", string(v.GetArray("layout")[0].Get("properties").GetStringBytes("navigationUrl")))
	require.True(t, v.GetArray("layout")[0].GetBool("public"))
	require.True(t, v.GetArray("layout")[0].GetBool("removable"))
	require.Equal(t, 123, v.GetInt("creator"))
	require.Equal(t, 123, v.GetArray("editors")[0].GetInt())
	require.Equal(t, 456, v.GetArray("editors")[1].GetInt())
}

func TestMarshalFastJSONWithEmptyLayout(t *testing.T) {
	ul := &UserLayout{
		UserID:   "user123",
		LayoutID: "layout123",
		Background: Background{
			Type:  "color",
			Value: "#FFFFFF",
		},
		Layout:  []LayoutElement{},
		Creator: 123,
		Editors: []int{123, 456},
	}

	a := fastjson.Arena{}
	v := ul.MarshalFastJSON(&a)

	require.Equal(t, "user123", string(v.GetStringBytes("userId")))
	require.Equal(t, "layout123", string(v.GetStringBytes("layoutId")))
	require.Equal(t, "color", string(v.Get("background").GetStringBytes("type")))
	require.Equal(t, "#FFFFFF", string(v.Get("background").GetStringBytes("value")))
	require.Empty(t, v.GetArray("layout"))
	require.Equal(t, 123, v.GetInt("creator"))
	require.Equal(t, 123, v.GetArray("editors")[0].GetInt())
	require.Equal(t, 456, v.GetArray("editors")[1].GetInt())
}

func TestMarshalFastJSONWithNilEditors(t *testing.T) {
	ul := &UserLayout{
		UserID:   "user123",
		LayoutID: "layout123",
		Background: Background{
			Type:  "color",
			Value: "#FFFFFF",
		},
		Layout:  []LayoutElement{},
		Creator: 123,
		Editors: nil,
	}

	a := fastjson.Arena{}
	v := ul.MarshalFastJSON(&a)

	require.Equal(t, "user123", string(v.GetStringBytes("userId")))
	require.Equal(t, "layout123", string(v.GetStringBytes("layoutId")))
	require.Equal(t, "color", string(v.Get("background").GetStringBytes("type")))
	require.Equal(t, "#FFFFFF", string(v.Get("background").GetStringBytes("value")))
	require.Empty(t, v.GetArray("layout"))
	require.Equal(t, 123, v.GetInt("creator"))
	require.Empty(t, v.GetArray("editors"))
}

func BenchmarkMarshal(b *testing.B) {
	ul := &UserLayout{
		UserID:   "user123",
		LayoutID: "layout123",
		Background: Background{
			Type:  "color",
			Value: "#FFFFFF",
		},
		Layout: []LayoutElement{
			{
				ElementID: "elem1",
				Type:      "clickable_navigable",
				Position: Position{
					Row:    1,
					Column: 1,
					Height: 1,
					Width:  1,
				},
				Properties: Properties{
					Icon:          "path/to/icon1.png",
					Title:         "Navigation Item",
					NavigationURL: "/some-page",
				},
				Public:    true,
				Removable: true,
			},
		},
		Creator: 123,
		Editors: []int{123, 456},
	}

	b.Run("FastJSON", func(b *testing.B) {
		a := fastjson.Arena{}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ul.MarshalFastJSON(&a)
		}
	})

	b.Run("EncodingJSON", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			json.Marshal(ul)
		}
	})
}

func TestUnmarshalUserLayoutFastJSON_ValidData(t *testing.T) {
	jsonStr := `{
		"userId": "user123",
		"layoutId": "layout123",
		"background": {
			"type": "color",
			"value": "#FFFFFF"
		},
		"layout": [
			{
				"elementId": "elem1",
				"type": "clickable_navigable",
				"position": {
					"row": 1,
					"column": 1,
					"height": 1,
					"width": 1
				},
				"properties": {
					"icon": "path/to/icon1.png",
					"title": "Navigation Item",
					"navigationUrl": "/some-page"
				},
				"public": true,
				"removable": true
			}
		],
		"creator": 123,
		"editors": [123, 456]
	}`

	v, err := fastjson.Parse(jsonStr)
	require.NoError(t, err)

	ul, err := UnmarshalUserLayoutFastJSON(v)
	require.NoError(t, err)
	require.Equal(t, "user123", ul.UserID)
	require.Equal(t, "layout123", ul.LayoutID)
	require.Equal(t, "color", ul.Background.Type)
	require.Equal(t, "#FFFFFF", ul.Background.Value)
	require.Len(t, ul.Layout, 1)
	require.Equal(t, "elem1", ul.Layout[0].ElementID)
	require.Equal(t, "clickable_navigable", ul.Layout[0].Type)
	require.Equal(t, 1, ul.Layout[0].Position.Row)
	require.Equal(t, 1, ul.Layout[0].Position.Column)
	require.Equal(t, 1, ul.Layout[0].Position.Height)
	require.Equal(t, 1, ul.Layout[0].Position.Width)
	require.Equal(t, "path/to/icon1.png", ul.Layout[0].Properties.Icon)
	require.Equal(t, "Navigation Item", ul.Layout[0].Properties.Title)
	require.Equal(t, "/some-page", ul.Layout[0].Properties.NavigationURL)
	require.True(t, ul.Layout[0].Public)
	require.True(t, ul.Layout[0].Removable)
	require.Equal(t, 123, ul.Creator)
	require.Equal(t, []int{123, 456}, ul.Editors)
}

func TestUnmarshalUserLayoutFastJSON_MissingFields(t *testing.T) {
	jsonStr := `{
		"userId": "user123",
		"layoutId": "layout123",
		"background": {
			"type": "color"
		},
		"layout": [],
		"creator": 123
	}`

	v, err := fastjson.Parse(jsonStr)
	require.NoError(t, err)

	ul, err := UnmarshalUserLayoutFastJSON(v)
	require.NoError(t, err)
	require.Equal(t, "user123", ul.UserID)
	require.Equal(t, "layout123", ul.LayoutID)
	require.Equal(t, "color", ul.Background.Type)
	require.Empty(t, ul.Background.Value)
	require.Empty(t, ul.Layout)
	require.Equal(t, 123, ul.Creator)
	require.Empty(t, ul.Editors)
}

func TestUnmarshalUserLayoutFastJSON_InvalidDataType(t *testing.T) {
	jsonStr := `{
		"userId": 123,
		"layoutId": "layout123",
		"background": {
			"type": "color",
			"value": "#FFFFFF"
		},
		"layout": [],
		"creator": 123,
		"editors": [123, 456]
	}`

	v, err := fastjson.Parse(jsonStr)
	require.NoError(t, err)

	_, err = UnmarshalUserLayoutFastJSON(v)
	require.Error(t, err)
}
