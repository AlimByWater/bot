package entity

import (
	"errors"
	"time"
)

var (
	// ErrNoPermission возникает, когда у пользователя нет прав на редактирование макета
	ErrNoPermission = errors.New("you don't have permission to edit this layout")
	// ErrLayoutNotFound возникает, когда запрашиваемый макет не найден
	ErrLayoutNotFound = errors.New("layout not found")
)

// UserLayout представляет собой структуру макета пользователя
type UserLayout struct {
	UserID     string          `json:"userId"`     // Идентификатор пользователя
	LayoutID   string          `json:"layoutId"`   // Уникальный идентификатор макета
	Background Background      `json:"background"` // Фон макета
	Layout     []LayoutElement `json:"layout"`     // Элементы макета
	Creator    int             `json:"creator"`    // Идентификатор создателя макета
	Editors    []int           `json:"editors"`    // Список идентификаторов редакторов макета
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
	Type  string `json:"type"`  // Тип фона (например, "цвет", "изображение")
	Value string `json:"value"` // Значение фона (например, "#FFFFFF" для цвета или URL для изображения)
}

// LayoutElement представляет собой структуру элемента макета
type LayoutElement struct {
	ElementID  string     `json:"elementId"`  // Уникальный идентификатор элемента
	Type       string     `json:"type"`       // Тип элемента (например, "кнопка", "текст")
	Position   Position   `json:"position"`   // Позиция элемента в макете
	Properties Properties `json:"properties"` // Свойства элемента
	Public     bool       `json:"public"`     // Флаг публичности элемента
	Removable  bool       `json:"removable"`  // Флаг возможности удаления элемента
}

// Position представляет собой структуру позиции элемента в макете
type Position struct {
	Row    int `json:"row"`    // Номер строки
	Column int `json:"column"` // Номер столбца
	Height int `json:"height"` // Высота элемента
	Width  int `json:"width"`  // Ширина элемента
}

// Properties представляет собой структуру свойств элемента макета
type Properties struct {
	Icon          string `json:"icon"`                    // Иконка элемента
	Title         string `json:"title"`                   // Заголовок элемента
	NavigationURL string `json:"navigationUrl,omitempty"` // URL для навигации (если применимо)
	CurrentValue  int    `json:"currentValue,omitempty"`  // Текущее значение (если применимо)
	MinValue      int    `json:"minValue,omitempty"`      // Минимальное значение (если применимо)
	MaxValue      int    `json:"maxValue,omitempty"`      // Максимальное значение (если применимо)
	Value         int    `json:"value,omitempty"`         // Значение элемента (если применимо)
}
