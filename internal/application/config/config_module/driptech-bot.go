package config_module

func NewDripTechBotConfig() *DripTechBot {
	return &DripTechBot{}
}

type DripTechBot struct {
	Token string
	Name  string
}

func (c DripTechBot) GetToken() string { return c.Token }

func (c DripTechBot) GetName() string { return c.Name }
