package config_module

import validation "github.com/go-ozzo/ozzo-validation"

type UserBot struct {
	UserBotAppID   int
	UserBotAppHash string
	UserBotTgPhone string
	SessionDir     string
}

func NewUserBotConfig() *UserBot {
	return &UserBot{}
}

func (c UserBot) GetSessionDir() string {
	return c.SessionDir
}

func (c UserBot) GetUserBotAppID() int {
	return c.UserBotAppID
}

func (c UserBot) GetUserBotAppHash() string {
	return c.UserBotAppHash
}

func (c UserBot) GetUserBotTgPhone() string {
	return c.UserBotTgPhone
}

func (c UserBot) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.UserBotAppID, validation.Required),
		validation.Field(&c.UserBotAppHash, validation.Required),
		validation.Field(&c.UserBotTgPhone, validation.Required),
		validation.Field(&c.SessionDir, validation.Required),
	)
}
