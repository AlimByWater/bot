package downloader

import (
	"context"
	"elysium/pkg/proto"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log/slog"
	"net"
	"strconv"
)

type config interface {
	GetGrpcHost() string
	GetGrpcPort() int
	Validate() error
}

type Module struct {
	ctx      context.Context
	logger   *slog.Logger
	grpcConn *grpc.ClientConn
	client   proto.FileDownloadServiceClient
	cfg      config
}

func New(cfg config) *Module {
	return &Module{
		cfg: cfg,
	}
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) error {
	m.ctx = ctx
	m.logger = logger.With(slog.String("module", "📂 downloader"))

	if err := m.cfg.Validate(); err != nil {
		return fmt.Errorf("downloader config validate: %w", err)
	}

	grpcAddress := net.JoinHostPort(m.cfg.GetGrpcHost(), strconv.Itoa(m.cfg.GetGrpcPort()))

	// Создаем клиент
	cc, err := grpc.NewClient(grpcAddress,
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(200*1024*1024), // 200 MB
		),
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		return fmt.Errorf("grpc server connection failed: %w", err)
	}

	m.grpcConn = cc

	m.client = proto.NewFileDownloadServiceClient(cc)
	return nil
}

func (m *Module) Close() (err error) {
	if m.grpcConn != nil {
		err = m.grpcConn.Close()
		if err != nil {
			return fmt.Errorf("grpc server close: %w", err)
		}
	}

	return nil
}
