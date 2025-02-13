package elysium

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"time"
)

const defaultQueryTimeout = 30 * time.Second

type (
	dbtx interface {
		ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
		PrepareContext(context.Context, string) (*sql.Stmt, error)
		QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
		QueryRowContext(context.Context, string, ...interface{}) *sql.Row
	}

	queries struct {
		db    dbtx
		audit auditer
	}

	auditer interface {
		SaveAudit(txnID string, payload string, version int, eventtime time.Time) (err error)
	}

	Repository struct {
		db    *sqlx.DB
		audit auditer
	}
)

func NewRepository(audit auditer) *Repository {
	return &Repository{
		audit: audit,
	}
}

// AddDb - добавляет в структуру пул соединений с СУБД
func (r *Repository) AddDb(db *sqlx.DB) {
	r.db = db
}

func (r *Repository) DB() *sqlx.DB {
	return r.db
}

func newQueries(db dbtx, audit auditer) *queries {
	return &queries{
		db:    db,
		audit: audit,
	}
}

func (r *Repository) execTX(ctx context.Context, fn func(*queries) error) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := newQueries(tx, r.audit)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %w, rb err: %s", err, rbErr.Error())
		}

		return fmt.Errorf("tx error: %w", err)
	}

	return tx.Commit()
}

func (r *Repository) Close() error {
	return r.db.Close()
}
