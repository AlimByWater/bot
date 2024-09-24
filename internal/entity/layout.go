package entity

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

var (
	ErrLayoutNotFound           = fmt.Errorf("layout not found")
	ErrNoPermissionToEditLayout = fmt.Errorf("no permission to edit layout")
)

// RootElement представляет собой структуру корневого элемента
//
//easyjson:json
type RootElement struct {
	ID                int        `json:"id" redis:"id"`
	Type              string     `json:"type" redis:"type"`
	Name              string     `json:"name" redis:"name"`
	URL               string     `json:"url" redis:"url"`
	External          bool       `json:"external" redis:"external"`
	AppType           string     `json:"app_type" redis:"app_type"`
	Description       string     `json:"description" redis:"description"`
	DefaultProperties Properites `json:"default_properties" redis:"default_properties"`
	IsPublic          bool       `json:"is_public" redis:"is_public"`
	IsPaid            bool       `json:"is_paid" redis:"is_paid"`
	CreatedAt         time.Time  `json:"created_at,omitempty" redis:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at,omitempty" redis:"updated_at"`
}

// UserLayout представляет собой структуру макета пользователя
//
//easyjson:json
type UserLayout struct {
	ID         int             `json:"id" redis:"id"`
	Name       string          `json:"name" redis:"name"`
	Background Background      `json:"background" redis:"background"`
	StreamURL  string          `json:"stream_url" redis:"stream_url"`
	Elements   []LayoutElement `json:"elements" redis:"elements"`
	Creator    int             `json:"creator" redis:"creator"`
	Editors    []int           `json:"editors" redis:"editors"`
	CreatedAt  time.Time       `json:"created_at" redis:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at" redis:"updated_at"`
}

type Background map[string]interface{}

func (b Background) Value() (driver.Value, error) {
	return json.Marshal(b)
}

func (b *Background) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	return json.Unmarshal(bytes, &b)

}

// Position представляет собой структуру позиции элемента в макете
//
//easyjson:json
type Position struct {
	X      int `json:"x" redis:"x"`
	Y      int `json:"y" redis:"y"`
	Z      int `json:"z" redis:"z"`
	Width  int `json:"width" redis:"width"`
	Height int `json:"height" redis:"height"`
}

// LayoutElement представляет собой структуру элемента макета
//
//easyjson:json
type LayoutElement struct {
	ID          int64       `json:"id" redis:"id"`
	RootElement RootElement `json:"root_element" redis:"root_element"`
	OnGridID    int64       `json:"on_grid_id" redis:"on_grid_id"`
	Position    Position    `json:"position" redis:"position"`
	Properties  Properites  `json:"properties" redis:"properties"`
	IsPublic    bool        `json:"is_public" redis:"is_public"`
	IsRemovable bool        `json:"is_removable" redis:"is_removable"`
}

type Properites map[string]interface{}

func (p Properites) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *Properites) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &p)
}

// LayoutChange представляет собой структуру для логирования изменений макета
type LayoutChange struct {
	ID         int                    `json:"id" redis:"id"`
	UserID     int                    `json:"user_id" redis:"user_id"`
	LayoutID   int                    `json:"layout_id" redis:"layout_id"`
	ChangeType string                 `json:"change_type" redis:"change_type"`
	Details    map[string]interface{} `json:"details" redis:"details"`
	Timestamp  time.Time              `json:"timestamp" redis:"timestamp"`
}

type ChangeDetails map[string]interface{}

func (p ChangeDetails) Value() (driver.Value, error) {
	return json.Marshal(p)
}

func (p *ChangeDetails) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &p)
}
