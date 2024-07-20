package entity

import "github.com/google/uuid"

const (
	ChatTypeSuperGroup = "supergroup"
)

type EventType string

const (
	EventTypeInitApp        EventType = "init_app"
	EventTypeCloseApp       EventType = "close_app"
	EventTypeStartRadio     EventType = "start_radio"
	EventTypeStartAnimation EventType = "start_animation"
	EventTypeMinimizeApp    EventType = "minimize_app"
	EventTypeMaximizeApp    EventType = "maximize_app"
	EventTypePauseAnimation EventType = "pause_animation"
)

type WebAppEvent struct {
	EventType      EventType   `json:"event_type"`
	TelegramUserID int64       `json:"telegram_user_id"`
	Payload        interface{} `json:"payload"`
	SessionID      uuid.UUID   `json:"session_id"`
}
