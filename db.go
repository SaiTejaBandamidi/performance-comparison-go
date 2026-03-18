package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MetricsStore struct {
	db *pgxpool.Pool

	restInflight    int64
	graphQLInflight int64
}

func NewMetricsStore(db *pgxpool.Pool) *MetricsStore {
	return &MetricsStore{db: db}
}

func (m *MetricsStore) Increment(requestType string) int {
	switch requestType {
	case "rest":
		return int(atomic.AddInt64(&m.restInflight, 1))
	case "graphql":
		return int(atomic.AddInt64(&m.graphQLInflight, 1))
	default:
		return 0
	}
}

func (m *MetricsStore) Decrement(requestType string) {
	switch requestType {
	case "rest":
		atomic.AddInt64(&m.restInflight, -1)
	case "graphql":
		atomic.AddInt64(&m.graphQLInflight, -1)
	}
}

func (m *MetricsStore) InsertMetric(
	ctx context.Context,
	requestType string,
	load int,
	requestTime time.Time,
	responseTime time.Time,
	totalTimeMS int64,
) error {
	query := `
		INSERT INTO api_request_metrics (
			request_type,
			load,
			request_time,
			response_time,
			total_time_ms
		)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := m.db.Exec(ctx, query, requestType, load, requestTime, responseTime, totalTimeMS)
	if err != nil {
		return fmt.Errorf("failed to insert metric: %w", err)
	}

	return nil
}

func NewPostgresPool(cfg DatabaseConfig) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}
