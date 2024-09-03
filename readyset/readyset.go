package postgres

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

type RS struct {
	Conn *pgx.Conn
}

func NewRS(dsn string) (*RS, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Printf("unable to connect to database: %v", err)
		return nil, err
	}
	return &RS{Conn: conn}, nil
}

func (c *RS) Close() {
	c.Conn.Close(context.Background())
}
