package config_module

type Soundcloud struct {
	ProxyURL     string
	DownloadPath string
}

func NewSoundcloudConfig() *Soundcloud {
	return &Soundcloud{}
}

func (c Soundcloud) GetProxyURL() string     { return c.ProxyURL }
func (c Soundcloud) GetDownloadPath() string { return c.DownloadPath }
func (c Soundcloud) Validate() error         { return nil }
