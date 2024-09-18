package config_module

import validation "github.com/go-ozzo/ozzo-validation"

type Downloader struct {
	GrpcHost string
	GrpcPort int
}

func NewDownloaderConfig() *Downloader {
	return &Downloader{}
}

func (c Downloader) GetGrpcHost() string { return c.GrpcHost }
func (c Downloader) GetGrpcPort() int    { return c.GrpcPort }

func (c Downloader) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.GrpcHost, validation.Required),
		validation.Field(&c.GrpcPort, validation.Required),
	)
}
