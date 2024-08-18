package config_module

type Http struct {
	Port string
	Mode string
}

func NewHttpConfig() *Http {
	return &Http{}
}

func (c Http) GetPort() string { return c.Port }
func (c Http) GetMode() string { return c.Mode }
