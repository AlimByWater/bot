package domain

import "time"

type User struct {
	Id         string    `json:"id"`
	TelegramId int64     `json:"telegram_id"`
	UserName   string    `json:"username,omitempty"`
	FirstName  string    `json:"first_name,omitempty"`
	Phone      string    `json:"phone,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
