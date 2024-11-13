package entity

import (
	"encoding/json"
	"time"
)

type EventType string

const (
	EventTypeInitApp       EventType = "init_app"
	EventTypeStartApp      EventType = "start_app"
	EventTypeStartAction   EventType = "start_action"
	EventTypeCollapseApp   EventType = "collapse_app"
	EventTypeExpandApp     EventType = "expand_app"
	EventTypeCloseAction   EventType = "close_action"
	EventTypeChangedStream EventType = "changed_stream"
)

// easyjson:json
type WebAppEvent struct {
	EventType  EventType       `json:"event_type"`
	SessionID  string          `json:"session_id"`
	TelegramID int64           `json:"telegram_id"`
	Payload    json.RawMessage `json:"payload"`
	Timestamp  time.Time       `json:"timestamp"`
	StreamSlug string          `json:"stream"`
}

// easyjson:json
type InitAppPayload struct {
	RawInitData string `json:"raw_init_data"`
	StreamSlug  string `json:"stream"`
}

// easyjson:json
type StartAppPayload struct {
}

// easyjson:json
type StartActionPayload struct {
	ActionID string `json:"action_id"`
}

// easyjson:json
type CollapseAppPayload struct {
	ActionID string `json:"action_id"`
}

// easyjson:json
type ExpandAppPayload struct {
	ActionID string `json:"action_id"`
}

// easyjson:json
type CloseActionPayload struct {
	ActionID string `json:"action_id"`
}

// easyjson:json
type ChangedStreamPayload struct {
	StreamSlug string `json:"stream"`
}
