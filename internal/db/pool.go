package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Pool struct {
	*pgxpool.Pool
	*Queries
	log *zap.Logger
}

func NewPool(databaseURL string) (*Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger, _ := zap.NewProduction()
	return &Pool{
		Pool:    pool,
		Queries: NewQueries(pool),
		log:     logger,
	}, nil
}

func (p *Pool) Close() {
	p.Pool.Close()
}

