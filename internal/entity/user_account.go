package entity

import "time"

type UserAccount struct {
	UserID    int       `db:"user_id" json:"user_id"`
	Balance   int       `db:"balance" json:"balance"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
