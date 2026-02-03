package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/thatreguy/trade.re/internal/config"
)

// DB wraps the connection pool
type DB struct {
	Pool *pgxpool.Pool
}

// New creates a new database connection pool
func New(cfg config.DatabaseConfig) (*DB, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable&pool_max_conns=%d",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.MaxConnections,
	)

	poolCfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("parsing connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, fmt.Errorf("creating connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the connection pool
func (db *DB) Close() {
	db.Pool.Close()
}

// Ping checks if the database is reachable
func (db *DB) Ping(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}
