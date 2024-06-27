package config_module

import validation "github.com/go-ozzo/ozzo-validation"

type Demethra struct {
	BotToken      string
	BotName       string
	ChatIDForLogs int64
}

func NewDemethraConfig() *Demethra {
	return &Demethra{}
}

func (c Demethra) GetBotToken() string     { return c.BotToken }
func (c Demethra) GetBotName() string      { return c.BotName }
func (c Demethra) GetChatIDForLogs() int64 { return c.ChatIDForLogs }

func (c Demethra) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.BotToken, validation.Required),
		validation.Field(&c.BotName, validation.Required),
		validation.Field(&c.ChatIDForLogs, validation.Required),
	)
}
