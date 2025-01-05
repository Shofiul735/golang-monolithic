package migrations

import (
	"embed"
	"fmt"
	"log"
	"time"

	"example.com/monolithic/internal/platform/database"
	"github.com/golang-migrate/migrate/v4"
)

//go:embed*.sql
var migrationFiles embed.FS

// RunMigrations runs database migrations
func RunMigrations(dbURL string) error {
	m, err := migrate.New(
		"file://migrations",
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("error creating migrate instance: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error running migrations: %v", err)
	}

	return nil
}

// Example migration file: migrations/001_create_users_table.up.sql
/*
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
*/

// Example usage in main.go
func main() {
	cfg := database.Config{
		Host:        "localhost",
		Port:        5432,
		User:        "postgres",
		Password:    "password",
		Database:    "myapp",
		MaxPoolSize: 10,
		MinPoolSize: 2,
		MaxIdleTime: 15 * time.Minute,
		MaxLifetime: 1 * time.Hour,
		HealthCheck: 30 * time.Second,
	}

	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := RunMigrations(cfg.GetConnectionURL()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}
