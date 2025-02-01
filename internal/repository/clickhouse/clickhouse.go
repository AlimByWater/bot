package clickhouse

import (
	"context"
	"errors"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"log/slog"
	"net"
	"reflect"
	"time"
)

type tableInsert interface {
	Init(save func(len int, query string, data []any) (err error), size int, saveTimeout int)
	QueryInit() string
	QueryInsert() string
	Len() int
	Data() []any
	Clean()
}

type tableQuery interface {
	AddConn(conn driver.Conn)
}

type config interface {
	GetUser() string
	GetPassword() string
	GetIP() []string
	GetDbname() string
	GetSize() int
	GetSaveTime() int
	GetDebug() bool
	GetMaxExecutionTime() int
	GetDialTimeout() int
	GetMaxOpenConns() int
	GetMaxIdleConns() int
	GetConnMaxLifetime() int
	GetBlockBufferSize() uint8
}

type Module struct {
	cfg     config
	ctx     context.Context
	logger  *slog.Logger
	inserts []tableInsert
	queries []tableQuery
	conn    driver.Conn
}

func New(cfg config) *Module {
	return &Module{
		cfg: cfg,
	}
}

func (m *Module) AddInsertTable(inserts ...tableInsert) {
	m.inserts = inserts
}

func (m *Module) AddQueriesTable(queries ...tableQuery) {
	m.queries = queries
}

func (m *Module) Init(ctx context.Context, logger *slog.Logger) (err error) {
	m.ctx = ctx
	m.logger = logger.With(slog.String("module", reflect.Indirect(reflect.ValueOf(m)).Type().PkgPath()))
	dialCount := 0
	if m.conn, err = clickhouse.Open(&clickhouse.Options{
		Addr: m.cfg.GetIP(),
		Auth: clickhouse.Auth{
			Database: m.cfg.GetDbname(),
			Username: m.cfg.GetUser(),
			Password: m.cfg.GetPassword(),
		},
		DialContext: func(ctx context.Context, addr string) (net.Conn, error) {
			dialCount++
			var d net.Dialer
			return d.DialContext(ctx, "tcp", addr)
		},
		Debug: m.cfg.GetDebug(),
		Debugf: func(format string, v ...interface{}) {
			m.logger.Debug(fmt.Sprintf(format, v))
		},
		Settings: clickhouse.Settings{
			"max_execution_time": m.cfg.GetMaxExecutionTime(),
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:      time.Duration(m.cfg.GetDialTimeout()) * time.Second,
		MaxOpenConns:     m.cfg.GetMaxOpenConns(),
		MaxIdleConns:     m.cfg.GetMaxIdleConns(),
		ConnMaxLifetime:  time.Duration(m.cfg.GetConnMaxLifetime()) * time.Minute,
		ConnOpenStrategy: clickhouse.ConnOpenRoundRobin,
		BlockBufferSize:  m.cfg.GetBlockBufferSize(),
	}); err != nil {
		return
	}
	if err = m.conn.Ping(m.ctx); err != nil {
		return
	}
	for i := range m.inserts {
		if err = m.conn.Exec(m.ctx, m.inserts[i].QueryInit()); err != nil {
			return
		}
		m.inserts[i].Init(m.save, m.cfg.GetSize(), m.cfg.GetSaveTime())
	}
	for i := range m.queries {
		m.queries[i].AddConn(m.conn)
	}
	return
}

func (m *Module) Close() (errs error) {
	if m.conn != nil {
		for i := range m.inserts {
			if m.inserts[i].Len() > 0 {
				err := m.save(m.inserts[i].Len(), m.inserts[i].QueryInsert(), m.inserts[i].Data())
				if err != nil {
					errs = errors.Join(err)
					continue
				}
				m.inserts[i].Clean()
			}
		}
		err := m.conn.Close()
		if err != nil {
			errs = errors.Join(err)
		}
	}
	return
}

func (m *Module) save(len int, query string, data []any) (err error) {
	if err = m.conn.Ping(m.ctx); err != nil {
		return
	}
	if len < 1 {
		return
	}
	var batch driver.Batch
	batch, err = m.conn.PrepareBatch(m.ctx, query)
	if err != nil {
		return
	}
	for i, v := range data {
		err = batch.Column(i).Append(v)
		if err != nil {
			return
		}
	}
	err = batch.Send()
	if err != nil {
		return
	}
	return
}
