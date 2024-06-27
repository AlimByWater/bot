package config_module

import validation "github.com/go-ozzo/ozzo-validation"

type ArimaDJ struct {
	BotToken      string
	BotName       string
	ChatIDForLogs int64
}

func NewArimaDJConfig() *ArimaDJ {
	return &ArimaDJ{}
}

func (c ArimaDJ) GetBotToken() string     { return c.BotToken }
func (c ArimaDJ) GetBotName() string      { return c.BotName }
func (c ArimaDJ) GetChatIDForLogs() int64 { return c.ChatIDForLogs }

func (c ArimaDJ) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.BotToken, validation.Required),
		validation.Field(&c.BotName, validation.Required),
		validation.Field(&c.ChatIDForLogs, validation.Required),
	)
}
