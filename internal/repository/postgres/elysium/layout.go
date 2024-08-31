package postgres

import (
	"arimadj-helper/internal/entity"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type LayoutRepository struct {
	db *sql.DB
}

func NewLayoutRepository(db *sql.DB) *LayoutRepository {
	return &LayoutRepository{db: db}
}
