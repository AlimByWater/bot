package elysium

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRepository_CreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	repo := &Repository{db: sqlxDB}

	ctx := context.Background()
	username := "testuser"
	email := "test@example.com"

	t.Run("successful user creation", func(t *testing.T) {
		expectedUser := &User{
			ID:        uuid.New(),
			Username:  username,
			Email:     email,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO users").
			WithArgs(sqlmock.AnyArg(), username, email, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id", "username", "email", "created_at", "updated_at"}).
				AddRow(expectedUser.ID, expectedUser.Username, expectedUser.Email, expectedUser.CreatedAt, expectedUser.UpdatedAt))
		mock.ExpectCommit()

		user, err := repo.CreateUser(ctx, username, email)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.Username, user.Username)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.NotEqual(t, uuid.Nil, user.ID)
		assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
		assert.WithinDuration(t, time.Now(), user.UpdatedAt, time.Second)
	})

	t.Run("database error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("INSERT INTO users").
			WithArgs(sqlmock.AnyArg(), username, email, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(sqlmock.ErrCancelled)
		mock.ExpectRollback()

		user, err := repo.CreateUser(ctx, username, email)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.EqualError(t, err, "sqlmock: query canceled")
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
