package config_module

import validation "github.com/go-ozzo/ozzo-validation"

type Redis struct {
	Host     string
	Port     int
	Password string
	DB       int
}

func NewRedisConfig() *Redis {
	return &Redis{}
}

func (c Redis) GetHost() string     { return c.Host }
func (c Redis) GetPort() int        { return c.Port }
func (c Redis) GetPassword() string { return c.Password }
func (c Redis) GetDB() int          { return c.DB }

func (c Redis) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Host, validation.Required),
		validation.Field(&c.Port, validation.Required),
		validation.Field(&c.Password, validation.Required),
	)
}
