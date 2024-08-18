package entity

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	EventTypeInitApp     EventType = "init_app"
	EventTypeStartApp    EventType = "start_app"
	EventTypeStartAction EventType = "start_action"
	EventTypeCollapseApp EventType = "collapse_app"
	EventTypeExpandApp   EventType = "expand_app"
	EventTypeCloseAction EventType = "close_action"
)

type WebAppEvent struct {
	EventType  EventType       `json:"event_type"`
	SessionID  string          `json:"session_id"`
	TelegramID int64           `json:"telegram_user_id"`
	Payload    json.RawMessage `json:"payload"`
	Timestamp  time.Time       `json:"timestamp"`
}

type InitAppPayload struct {
	RawInitData string `json:"raw_init_data"`
}

type StartAppPayload struct {
}

type StartActionPayload struct {
	ActionID string `json:"action_id"`
}

type CollapseAppPayload struct {
	ActionID string `json:"action_id"`
}

type ExpandAppPayload struct {
	ActionID string `json:"action_id"`
}

type CloseActionPayload struct {
	ActionID string `json:"action_id"`
}
