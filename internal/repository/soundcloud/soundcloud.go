package soundcloud

import (
	"arimadj-helper/internal/repository/soundcloud/pkg/soundcloud"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

type config interface {
	GetProxyURL() string
	Validate() error
}

type Module struct {
	cfg    config
	logger *slog.Logger
	sc     *soundcloud.Soundcloud
}

func New(cfg config) *Module {
	return &Module{
		cfg: cfg,
	}
}

func (m *Module) Init(ctx context.Context, log *slog.Logger) error {
	m.logger = log.With(slog.StringValue("☁️ soundcloud repo"))
	client := http.DefaultClient
	if m.cfg.GetProxyURL() != "" {
		proxyUrl, err := url.Parse(m.cfg.GetProxyURL())
		if err != nil {
			return fmt.Errorf("proxy parse: %w", err)
		}

		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	}

	m.sc = soundcloud.NewClient("", client)
	return nil
}

func (m *Module) DownloadTrackByURL(_ context.Context, trackUrl string) {

	//m.sc.Download(trackUrl)
}
