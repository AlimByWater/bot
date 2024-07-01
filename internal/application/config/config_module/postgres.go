package config_module

import validation "github.com/go-ozzo/ozzo-validation"

type Postgres struct {
	Driver   string
	Name     string
	Host     string
	Port     int
	Password string
	User     string
	MaxConn  int
	SSLMode  string
}

func NewPostgresConfig() *Postgres {
	return &Postgres{}
}

func (c Postgres) GetDriver() string   { return c.Driver }
func (c Postgres) GetName() string     { return c.Name }
func (c Postgres) GetHost() string     { return c.Host }
func (c Postgres) GetPort() int        { return c.Port }
func (c Postgres) GetPassword() string { return c.Password }
func (c Postgres) GetUser() string     { return c.User }
func (c Postgres) GetMaxConn() int     { return c.MaxConn }
func (c Postgres) GetSSLMode() string  { return c.SSLMode }

func (c Postgres) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Driver, validation.Required),
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.Host, validation.Required),
		validation.Field(&c.Port, validation.Required),
		validation.Field(&c.Password, validation.Required),
		validation.Field(&c.User, validation.Required),
		validation.Field(&c.SSLMode, validation.Required),
		validation.Field(&c.MaxConn, validation.Required),
	)
}
