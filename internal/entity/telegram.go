package entity

import "time"

const (
	BotSelfCtxKey = "bot_self"
)

const (
	ChatTypeSuperGroup = "supergroup"
	ChatTypePrivate    = "private"
)

type BotUpdate struct {
	BotID      int64     `json:"bot_id" clickhouse:"bot_id"`
	UpdateTime time.Time `json:"update_time" clickhouse:"update_time"`
	Payload    string    `json:"payload" clickhouse:"payload"`
}

const (
	CacheKeyInitMessageToDelete = "init_message_to_delete"
)
