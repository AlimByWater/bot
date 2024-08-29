package entity

import "errors"

var (
	ErrNoPermission   = errors.New("you don't have permission to edit this layout")
	ErrLayoutNotFound = errors.New("layout not found")
)

type UserLayout struct {
	UserID     string          `json:"userId"`
	LayoutID   string          `json:"layoutId"`
	Background Background      `json:"background"`
	Layout     []LayoutElement `json:"layout"`
	Creator    int             `json:"creator"`
	Editors    []int           `json:"editors"`
}

type LayoutChange struct {
	UserID    int       `json:"userId"`
	LayoutID  string    `json:"layoutId"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
	Details   string    `json:"details"`
}

type Background struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type LayoutElement struct {
	ElementID  string     `json:"elementId"`
	Type       string     `json:"type"`
	Position   Position   `json:"position"`
	Properties Properties `json:"properties"`
	Visibility string     `json:"visibility"`
	Removable  bool       `json:"removable"`
}

type Position struct {
	Row    int `json:"row"`
	Column int `json:"column"`
	Height int `json:"height"`
	Width  int `json:"width"`
}

type Properties struct {
	Icon          string `json:"icon"`
	Title         string `json:"title"`
	NavigationURL string `json:"navigationUrl,omitempty"`
	CurrentValue  int    `json:"currentValue,omitempty"`
	MinValue      int    `json:"minValue,omitempty"`
	MaxValue      int    `json:"maxValue,omitempty"`
	Value         int    `json:"value,omitempty"`
}
