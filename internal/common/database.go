package common

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InitDatabase initializes a PostgreSQL connection pool
func InitDatabase(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	dbPool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}

	if err := dbPool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	log.Println("Connected to PostgreSQL")
	return dbPool, nil
}
