package http

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/gin-gonic/gin"
	"log/slog"
	"net/http"
	"reflect"
	"runtime"
	"time"
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

	var certFilePath string
	var keyFilePath string

	switch runtime.GOOS {
	case "windows":
		certFilePath = `C:\ssl\cert.pem`
		keyFilePath = `C:\ssl\key.pem`
	default:
		certFilePath = "/app/ssl/cert.pem"
		keyFilePath = "/app/ssl/key.pem"
	}

	serverTLSCert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
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
	if err := m.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
