package http

import (
	"context"
	"crypto/tls"
	"errors"
	"log/slog"
	"net/http"
	"reflect"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

type group interface {
	Path() string
	Handlers() []func() (method string, path string, handlerFunc gin.HandlerFunc)
	Auth() gin.HandlerFunc
}

type config interface {
	GetPort() string
	GetMode() string
}

// New коструктор http контролера
func New(cfg config, groups ...group) *Module {
	return &Module{cfg: cfg, groups: groups}
}

// Module структура http контроллера
type Module struct {
	server *http.Server
	ctx    context.Context
	stop   context.CancelFunc
	logger *slog.Logger
	cfg    config
	groups []group

	certFilePath string
	keyFilePath  string
}

// Init инициализатор контролера
func (m *Module) Init(ctx context.Context, stop context.CancelFunc, logger *slog.Logger) (err error) {
	m.ctx = ctx
	m.stop = stop
	m.logger = logger.With(slog.String("module", reflect.Indirect(reflect.ValueOf(m)).Type().PkgPath()))

	gin.SetMode(m.cfg.GetMode())
	router := gin.New()
	router.Use(ginRecovery(m.logger), cors(m.logger))

	// добавляем группы эндпоинтов
	for _, g := range m.groups {
		rg := router.Group(g.Path(), g.Auth())
		for _, handler := range g.Handlers() {
			rg.Handle(handler())
		}
	}

	switch runtime.GOOS {
	case "windows":
		m.certFilePath = `C:\ssl\cert.pem`
		m.keyFilePath = `C:\ssl\key.pem`
	case "linux":
		m.certFilePath = "/app/ssl/elysium.pem"
		m.keyFilePath = "/app/ssl/elysium_key.pem"
	case "darwin":
		m.certFilePath = "/Users/admin/ssl/cert.pem"
		m.keyFilePath = "/Users/admin/ssl/key.pem"
	}

	serverTLSCert, err := tls.LoadX509KeyPair(m.certFilePath, m.keyFilePath)
	if err != nil {
		m.logger.Warn("Error loading certificate and key file", slog.String("err", err.Error()))
		m.server = &http.Server{
			Addr:    m.cfg.GetPort(),
			Handler: router,
		}

	} else {
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{serverTLSCert},
		}

		m.server = &http.Server{
			Addr:      m.cfg.GetPort(),
			Handler:   router,
			TLSConfig: tlsConfig,
		}
	}

	return
}

func (m *Module) Run() {
	go m.run()
}

func (m *Module) run() {
	if err := m.server.ListenAndServeTLS(m.certFilePath, m.keyFilePath); err != nil && !errors.Is(err, http.ErrServerClosed) {
		m.logger.Error("Listen and serve", slog.String("err", err.Error()))
	}
	m.stop()
}

func (m *Module) Shutdown() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if m.server != nil {
		err = m.server.Shutdown(ctx)
	}
	return
}
