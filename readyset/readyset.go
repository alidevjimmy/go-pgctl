package readyset

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RS struct {
	ConnPool *pgxpool.Pool
}

func NewRS(dsn string) (*RS, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	config.MinConns = 20
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	return &RS{ConnPool: pool}, nil
}

func (c *RS) Close() {
	c.ConnPool.Close()
}
