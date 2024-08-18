package config_module

import validation "github.com/go-ozzo/ozzo-validation"

type Auth struct {
	Secret           string
	TelegramBotToken string
	ApiKey           string
}

func NewAuthConfig() *Auth {
	return &Auth{}
}

func (c Auth) GetSecret() string           { return c.Secret }
func (c Auth) GetTelegramBotToken() string { return c.TelegramBotToken }
func (c Auth) GetApiKey() string           { return c.ApiKey }

func (c Auth) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Secret, validation.Required),
		validation.Field(&c.TelegramBotToken, validation.Required),
		validation.Field(&c.ApiKey, validation.Required),
	)
}
