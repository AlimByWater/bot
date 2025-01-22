package entity

import botapi "github.com/go-telegram/bot"

const (
	ProjectEmojiGenCommon = "emoji-gen-common" // ДЛЯ ГЕНЕРАЦИИ ЭМОДЖИ ПАКОВ НА ФОРУМЕ/В ЧАТЕ
	ProjectEmojiGenVip    = "emoji-gen-vip"    // ДЛЯ ГЕНЕРАЦИИ ЭМОДЖИ ПАКОВ В ЛИЧНЫХ СООБЩЕНИЯХ
	ProjectEmojiGenDrip   = "emoji-gen-drip"
	ProjectThreeD         = "three-d-3d"
)

// easyjson:json
type Bot struct {
	ID      int64       `json:"id" db:"id"`
	Name    string      `json:"name" db:"name"`
	Token   string      `json:"token" db:"token"`
	Purpose string      `json:"purpose" db:"purpose"`
	Test    bool        `json:"test" db:"test"`
	Enabled bool        `json:"enabled" db:"enabled"`
	BotApi  *botapi.Bot `json:"-" db:"-"`
}
