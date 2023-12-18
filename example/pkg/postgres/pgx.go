package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	User = "postgres"
	Pass = "123456"
	Host = "localhost"
	Port = "5433"
	Name = "mq"
)

// Create new pool
func NewPool(names ...string) (pool *pgxpool.Pool, err error) {
	var (
		name string
	)
	if len(names) > 0 {
		name = names[0]
	} else {
		name = Name
	}

	dsn := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		User, Pass, name, Host, Port)
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return pool, err
	}

	config.MaxConns = int32(2)
	config.MaxConnLifetime = time.Minute
	config.MaxConnIdleTime = time.Minute
	pool, err = pgxpool.ConnectConfig(context.Background(), config)
	return pool, err
}
