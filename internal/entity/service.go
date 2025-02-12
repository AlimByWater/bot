package entity

import (
	"fmt"
	"time"
)

const (
	ServiceTypeEmojiGenerator = "emoji_generator"
)

var (
	ErrServiceNotFound = fmt.Errorf("service not found")
)

// easyjson:json
type Service struct {
	ID          int       `json:"id" db:"id"`
	BotID       int64     `json:"bot_id" db:"bot_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Price       int       `json:"price" db:"price"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
