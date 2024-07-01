package postgres

import (
	"context"
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log/slog"
)

var initElysiumSchema = `CREATE SCHEMA IF NOT EXISTS elysium;`

type config interface {
	GetDriver() string
	GetUser() string
	GetPassword() string
	GetHost() string
	GetPort() int
	GetName() string
	GetMaxConn() int
	GetSSLMode() string
}

// table - интерфейс таблицы БД для добавления в репозиторий
type table interface {
	AddDb(db *sqlx.DB) // AddDb - добавляет к таблице подключение к БД
	//QueryInit() (bool, string) // QueryInit - возвращает строку для создания таблицы в БД
}

// Module - структура модуля репозитория
type Module struct {
	cfg    config
	db     *sqlx.DB
	tables []table
}

// New - создает новый модуль, на входе конфигурация и таблицы
func New(cfg config, t ...table) *Module {
	return &Module{
		cfg:    cfg,
		tables: t,
	}
}

func (m *Module) Init(ctx context.Context, _ *slog.Logger) (err error) {
	connStr := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=%s",
		m.cfg.GetDriver(),
		m.cfg.GetUser(),
		m.cfg.GetPassword(),
		m.cfg.GetHost(),
		m.cfg.GetPort(),
		m.cfg.GetName(),
		m.cfg.GetSSLMode(),
	)

	db, err := sqlx.Open(m.cfg.GetDriver(), connStr)
	if err != nil {
		return
	}

	m.db = db

	// создаем схему Elysium если ее еще нет
	_, err = m.db.ExecContext(ctx, initElysiumSchema)
	if err != nil {
		return
	}

	for i := range m.tables {
		m.tables[i].AddDb(m.db)
		//if query := m.tables[i].QueryInit(); query != "" {
		//	_, err = m.db.Exec(ctx, query)
		//	if err != nil {
		//		return
		//	}
		//}
	}
	return
}

// Close - закрывает модуль при завершении работы приложения
func (m *Module) Close() (err error) {
	if m.db != nil {
		m.db.Close()
	}
	return
}
