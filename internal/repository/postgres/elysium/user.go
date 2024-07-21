package elysium

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `db:"id"`
	Username  string    `db:"username"`
	Email     string    `db:"email"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func (r *Repository) CreateUser(ctx context.Context, username, email string) (*User, error) {
	query := `
		INSERT INTO users (id, username, email, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, username, email, created_at, updated_at
	`

	user := &User{
		ID:        uuid.New(),
		Username:  username,
		Email:     email,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := r.db.QueryRowContext(ctx, query,
		user.ID, user.Username, user.Email, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID, &user.Username, &user.Email, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return user, nil
}
