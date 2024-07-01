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
		db dbtx
	}

	UserRepository struct {
		db *sqlx.DB
	}
)

func NewUserRepository() *UserRepository {
	return &UserRepository{}
}

// AddDb - добавляет в структуру пул соединений с СУБД
func (t *UserRepository) AddDb(db *sqlx.DB) {
	t.db = db
}

func newQueries(db dbtx) *queries {
	return &queries{db: db}
}

func (t *UserRepository) execTX(ctx context.Context, fn func(*queries) error) error {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := newQueries(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %w, rb err: %s", err, rbErr.Error())
		}

		return fmt.Errorf("tx error: %w", err)
	}

	return tx.Commit()
}

func (t *UserRepository) Close() error {
	return t.db.Close()
}
