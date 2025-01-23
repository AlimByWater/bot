package config_module

import validation "github.com/go-ozzo/ozzo-validation"

type Telegram struct {
	UserBotAppID   int
	UserBotAppHash string
	UserBotTgPhone string
	Local          bool
	SessionDir     string
}

func NewTelegramConfig() *Telegram {
	return &Telegram{}
}

func (c Telegram) IsLocal() bool {
	return c.Local
}

func (c Telegram) GetSessionDir() string {
	return c.SessionDir
}

func (c Telegram) GetUserBotAppID() int {
	return c.UserBotAppID
}

func (c Telegram) GetUserBotAppHash() string {
	return c.UserBotAppHash
}

func (c Telegram) GetUserBotTgPhone() string {
	return c.UserBotTgPhone
}

func (c Telegram) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.UserBotAppID, validation.Required),
		validation.Field(&c.UserBotAppHash, validation.Required),
		validation.Field(&c.UserBotTgPhone, validation.Required),
		validation.Field(&c.SessionDir, validation.Required),
	)
}
