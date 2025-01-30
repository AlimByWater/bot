package config_module

import validation "github.com/go-ozzo/ozzo-validation"

type Message struct {
	DefaultLanguage string
	DirectoryName   string
}

func NewMessageConfig() *Message {
	return &Message{}
}

func (c Message) GetDefaultLanguage() string { return c.DefaultLanguage }
func (c Message) GetDirectoryName() string   { return c.DirectoryName }

func (c Message) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DefaultLanguage, validation.Required),
		validation.Field(&c.DirectoryName, validation.Required),
	)
}
