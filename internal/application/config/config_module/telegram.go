package config_module

import validation "github.com/go-ozzo/ozzo-validation"

type Telegram struct {
	UserBotAppID   int
	UserBotAppHash string
	UserBotTgPhone string
}

func NewTelegramConfig() *Telegram {
	appID, _ := strconv.Atoi(os.Getenv("TELEGRAM_APP_ID"))
	appHash := os.Getenv("TELEGRAM_APP_HASH")
	phone := os.Getenv("TELEGRAM_PHONE_NUMBER")

	return &Telegram{
		UserBotAppID:   appID,
		UserBotAppHash: appHash,
		UserBotTgPhone: phone,
	}
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
	)
}
