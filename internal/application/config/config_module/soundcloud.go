package config_module

type Soundcloud struct {
	ProxyURL string
}

func NewSoundcloudConfig() *Soundcloud {
	return &Soundcloud{}
}

func (c Soundcloud) GetProxyURL() string { return c.ProxyURL }
func (c Soundcloud) Validate() error     { return nil }
