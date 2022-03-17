package postgres

import (
	"context"
	"fmt"
	"sizebot/internal/config"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Storage contains methods for retrieve bids from postgresql.
type Storage struct {
	mainDB *pgxpool.Pool
}

// New create new Storage object.
func New(ctx context.Context, conf *config.Postgres) (*Storage, error) {
	mainDB, err := pgxpool.Connect(ctx, conf.DbConnectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to create mainDB pgx pool: %w", err)
	}

	return &Storage{
		mainDB: mainDB,
	}, nil
}

// Close releases underlying db resources.
func (s *Storage) Close() error {
	s.mainDB.Close()

	return nil
}
