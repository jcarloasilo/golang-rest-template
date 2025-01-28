package database

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	ErrRecordNotFound = errors.New("record not found")
)

const defaultTimeout = 3 * time.Second

func NewPool(connStr string) (*pgxpool.Pool, error) {

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	dbpool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, err
	}

	dbpool.Config().MaxConns = 25
	dbpool.Config().MaxConnIdleTime = 5 * time.Minute
	dbpool.Config().MaxConnLifetime = 2 * time.Hour

	return dbpool, nil
}
