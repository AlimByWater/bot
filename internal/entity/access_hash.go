package entity

import "time"

type AccessHash struct {
	ChatID    string    `db:"chat_id" json:"chat_id"`
	Hash      int64     `db:"hash" json:"hash"`
	PeerID    int64     `db:"peer_id" json:"peer_id"`
	Username  string    `db:"username" json:"username"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
