package soundcloud

import (
	"arimadj-helper/internal/entity"
	"arimadj-helper/internal/repository/soundcloud/pkg/soundcloud"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
)

type configs interface {
	GetProxyURL() string
	GetDownloadPath() string
	Validate() error
}

type Module struct {
	cfg    configs
	logger *slog.Logger
	sc     *soundcloud.Soundcloud
}

func New(cfg configs) *Module {
	return &Module{
		cfg: cfg,
	}
}

func (m *Module) Init(ctx context.Context, log *slog.Logger) error {
	m.logger = log.With(slog.String("module", "☁️ soundcloud repo"))
	client := http.DefaultClient
	if m.cfg.GetProxyURL() != "" {
		proxyUrl, err := url.Parse(m.cfg.GetProxyURL())
		if err != nil {
			return fmt.Errorf("proxy parse: %w", err)
		}

		client.Transport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	}

	m.sc = soundcloud.NewClient("", m.cfg.GetDownloadPath(), m.cfg.GetProxyURL(), client, log)
	return nil
}

func (m *Module) Close() error {
	return nil
}

func (m *Module) DownloadTrackByURL(ctx context.Context, trackUrl string, info entity.TrackInfo) (string, error) {
	var err error
	var songPath string

	for i := 0; i < 3; i++ {
		songPath, err = m.sc.Download(ctx, trackUrl, info)
		if err != nil {
			fmt.Println(i)
			continue
		}
		break
	}

	if err != nil {
		return "", fmt.Errorf("download track: %w", err)
	}

	return songPath, nil
}
