package database

import (
	"context"
	"embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// Connect creates a new database connection pool and runs migrations
func Connect(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	// Run migrations
	migrations := []string{
		"migrations/001_initial_schema.up.sql",
		"migrations/002_add_location_to_jobs.up.sql",
		"migrations/003_add_authentication.up.sql",
		"migrations/004_application_state.up.sql",
	}

	for _, migration := range migrations {
		upSQL, err := migrationFS.ReadFile(migration)
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("failed to read migration %s: %w", migration, err)
		}

		if _, err = pool.Exec(ctx, string(upSQL)); err != nil {
			// Ignore if already exists errors
			errMsg := err.Error()
			if errMsg != "ERROR: relation \"user_profiles\" already exists (SQLSTATE 42P07)" &&
				errMsg != "ERROR: column \"location\" of relation \"jobs\" already exists (SQLSTATE 42701)" &&
				errMsg != "ERROR: column \"password_hash\" of relation \"user_profiles\" already exists (SQLSTATE 42701)" &&
				errMsg != "ERROR: column \"paused_at\" of relation \"applications\" already exists (SQLSTATE 42701)" {
				pool.Close()
				return nil, fmt.Errorf("failed to run migration %s: %w", migration, err)
			}
		}
	}

	return pool, nil
}
