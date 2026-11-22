package database

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	Host     string `env:"DATABASE_HOST"     envDefault:"localhost"`
	Port     int    `env:"DATABASE_PORT"     envDefault:"5432"`
	Name     string `env:"DATABASE_NAME"     envDefault:"purgebot"`
	User     string `env:"DATABASE_USER"     envDefault:"purgebot"`
	Password string `env:"DATABASE_PASSWORD"`
}

func (c Config) URL() string {
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.User, c.Password),
		Host:   net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:   "/" + c.Name,
	}
	return u.String()
}

type Database struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, connStr string) (*Database, error) {
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &Database{pool: pool}, nil
}

func (db *Database) Close() {
	db.pool.Close()
}
