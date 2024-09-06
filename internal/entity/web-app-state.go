package entity

import (
	"encoding/json"
	"time"
)

type StateType string

const (
	StateTypeInit      StateType = "init"
	StateTypeAction    StateType = "action"
	StateTypeInterface StateType = "interface"
)

// easyjson:json
type WebAppState struct {
	StateType  StateType       `json:"state_type"`
	SessionID  string          `json:"session_id"`
	TelegramID int64           `json:"telegram_id"`
	Payload    json.RawMessage `json:"payload"`
	Timestamp  time.Time       `json:"timestamp"`
}

// easyjson:json
type InitStatePayload struct {
}

// easyjson:json
type ActionStatePayload struct {
}

// easyjson:json
type InterfaceStatePayload struct {
}
