package entity

import (
	"errors"
	"time"
	"github.com/valyala/fastjson"
)

var (
	// ErrNoPermission возникает, когда у пользователя нет прав на редактирование макета
	ErrNoPermission = errors.New("you don't have permission to edit this layout")
	// ErrLayoutNotFound возникает, когда запрашиваемый макет не найден
	ErrLayoutNotFound = errors.New("layout not found")
)

// UserLayout представляет собой структуру макета пользователя
type UserLayout struct {
	UserID     string          `json:"userId" redis:"userId"`         // Идентификатор пользователя
	LayoutID   string          `json:"layoutId" redis:"layoutId"`     // Уникальный идентификатор макета
	Background Background      `json:"background" redis:"background"` // Фон макета
	Layout     []LayoutElement `json:"layout" redis:"layout"`         // Элементы макета
	Creator    int             `json:"creator" redis:"creator"`       // Идентификатор создателя макета
	Editors    []int           `json:"editors" redis:"editors"`       // Список идентификаторов редакторов макета
}

// MarshalFastJSON сериализует UserLayout в fastjson.Value
func (ul *UserLayout) MarshalFastJSON(a *fastjson.Arena) *fastjson.Value {
	o := a.NewObject()
	o.Set("userId", a.NewString(ul.UserID))
	o.Set("layoutId", a.NewString(ul.LayoutID))
	
	bg := a.NewObject()
	bg.Set("type", a.NewString(ul.Background.Type))
	bg.Set("value", a.NewString(ul.Background.Value))
	o.Set("background", bg)
	
	layout := a.NewArray()
	for _, elem := range ul.Layout {
		elemObj := a.NewObject()
		elemObj.Set("elementId", a.NewString(elem.ElementID))
		elemObj.Set("type", a.NewString(elem.Type))
		
		pos := a.NewObject()
		pos.Set("row", a.NewNumberInt(elem.Position.Row))
		pos.Set("column", a.NewNumberInt(elem.Position.Column))
		pos.Set("height", a.NewNumberInt(elem.Position.Height))
		pos.Set("width", a.NewNumberInt(elem.Position.Width))
		elemObj.Set("position", pos)
		
		props := a.NewObject()
		props.Set("icon", a.NewString(elem.Properties.Icon))
		props.Set("title", a.NewString(elem.Properties.Title))
		props.Set("navigationUrl", a.NewString(elem.Properties.NavigationURL))
		props.Set("currentValue", a.NewNumberInt(elem.Properties.CurrentValue))
		props.Set("minValue", a.NewNumberInt(elem.Properties.MinValue))
		props.Set("maxValue", a.NewNumberInt(elem.Properties.MaxValue))
		props.Set("value", a.NewNumberInt(elem.Properties.Value))
		elemObj.Set("properties", props)
		
		elemObj.Set("public", a.NewBool(elem.Public))
		elemObj.Set("removable", a.NewBool(elem.Removable))
		
		layout.SetArrayItem(layout.Len(), elemObj)
	}
	o.Set("layout", layout)
	
	o.Set("creator", a.NewNumberInt(ul.Creator))
	
	editors := a.NewArray()
	for i, editor := range ul.Editors {
		editors.SetArrayItem(i, a.NewNumberInt(editor))
	}
	o.Set("editors", editors)
	
	return o
}

// UnmarshalUserLayoutFastJSON десериализует fastjson.Value в UserLayout
func UnmarshalUserLayoutFastJSON(v *fastjson.Value) (*UserLayout, error) {
	ul := &UserLayout{}
	
	ul.UserID = string(v.GetStringBytes("userId"))
	ul.LayoutID = string(v.GetStringBytes("layoutId"))
	
	bg := v.Get("background")
	ul.Background.Type = string(bg.GetStringBytes("type"))
	ul.Background.Value = string(bg.GetStringBytes("value"))
	
	layout := v.GetArray("layout")
	for _, elem := range layout {
		le := LayoutElement{}
		le.ElementID = string(elem.GetStringBytes("elementId"))
		le.Type = string(elem.GetStringBytes("type"))
		
		pos := elem.Get("position")
		le.Position.Row = pos.GetInt("row")
		le.Position.Column = pos.GetInt("column")
		le.Position.Height = pos.GetInt("height")
		le.Position.Width = pos.GetInt("width")
		
		props := elem.Get("properties")
		le.Properties.Icon = string(props.GetStringBytes("icon"))
		le.Properties.Title = string(props.GetStringBytes("title"))
		le.Properties.NavigationURL = string(props.GetStringBytes("navigationUrl"))
		le.Properties.CurrentValue = props.GetInt("currentValue")
		le.Properties.MinValue = props.GetInt("minValue")
		le.Properties.MaxValue = props.GetInt("maxValue")
		le.Properties.Value = props.GetInt("value")
		
		le.Public = elem.GetBool("public")
		le.Removable = elem.GetBool("removable")
		
		ul.Layout = append(ul.Layout, le)
	}
	
	ul.Creator = v.GetInt("creator")
	
	editors := v.GetArray("editors")
	for _, editor := range editors {
		ul.Editors = append(ul.Editors, editor.GetInt())
	}
	
	return ul, nil
}

// LayoutChange представляет собой структуру для логирования изменений макета
type LayoutChange struct {
	UserID    int       `json:"userId"`    // Идентификатор пользователя, внесшего изменения
	LayoutID  string    `json:"layoutId"`  // Идентификатор измененного макета
	Timestamp time.Time `json:"timestamp"` // Время внесения изменений
	Action    string    `json:"action"`    // Тип действия (например, "обновление", "добавление элемента")
	Details   string    `json:"details"`   // Дополнительные детали изменения
}

// Background представляет собой структуру фона макета
type Background struct {
	Type  string `json:"type" redis:"type"`   // Тип фона (например, "цвет", "изображение")
	Value string `json:"value" redis:"value"` // Значение фона (например, "#FFFFFF" для цвета или URL для изображения)
}

// LayoutElement представляет собой структуру элемента макета
type LayoutElement struct {
	ElementID  string     `json:"elementId" redis:"elementId"`   // Уникальный идентификатор элемента
	Type       string     `json:"type" redis:"type"`             // Тип элемента (например, "кнопка", "текст")
	Position   Position   `json:"position" redis:"position"`     // Позиция элемента в макете
	Properties Properties `json:"properties" redis:"properties"` // Свойства элемента
	Public     bool       `json:"public" redis:"public"`         // Флаг публичности элемента
	Removable  bool       `json:"removable" redis:"removable"`   // Флаг возможности удаления элемента
}

// Position представляет собой структуру позиции элемента в макете
type Position struct {
	Row    int `json:"row" redis:"row"`       // Номер строки
	Column int `json:"column" redis:"column"` // Номер столбца
	Height int `json:"height" redis:"height"` // Высота элемента
	Width  int `json:"width" redis:"width"`   // Ширина элемента
}

// Properties представляет собой структуру свойств элемента макета
type Properties struct {
	Icon          string `json:"icon" redis:"icon"`                             // Иконка элемента
	Title         string `json:"title" redis:"title"`                           // Заголовок элемента
	NavigationURL string `json:"navigationUrl,omitempty" redis:"navigationUrl"` // URL для навигации (если применимо)
	CurrentValue  int    `json:"currentValue,omitempty" redis:"currentValue"`   // Текущее значение (если применимо)
	MinValue      int    `json:"minValue,omitempty" redis:"minValue"`           // Минимальное значение (если применимо)
	MaxValue      int    `json:"maxValue,omitempty" redis:"maxValue"`           // Максимальное значение (если применимо)
	Value         int    `json:"value,omitempty" redis:"value"`                 // Значение элемента (если применимо)
}
