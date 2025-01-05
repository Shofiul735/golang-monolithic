package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// Config holds the database configuration
type Config struct {
	Host        string
	Port        int
	User        string
	Password    string
	Database    string
	MaxPoolSize int32
	MinPoolSize int32
	MaxIdleTime time.Duration
	MaxLifetime time.Duration
	HealthCheck time.Duration
	SSLMode     string // Added for SSL configuration
}

// DB represents our database connection
type DB struct {
	pool *pgxpool.Pool
	cfg  Config
}

func (c *Config) GetConnectionURL() string {
	// If SSLMode is not set, default to disable for local development
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "disable"
	}

	// Format: postgresql://username:password@host:port/dbname?sslmode=disable
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		sslMode,
	)
}

// NewConnection establishes a new database connection pool
func NewConnection(cfg Config) (*DB, error) {
	// Construct connection URL
	connString := fmt.Sprintf(
		"postgresql://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database,
	)

	// Configure the connection pool
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("error parsing connection string: %v", err)
	}

	// Set pool configuration
	poolConfig.MaxConns = cfg.MaxPoolSize
	poolConfig.MinConns = cfg.MinPoolSize
	poolConfig.MaxConnLifetime = cfg.MaxLifetime
	poolConfig.MaxConnIdleTime = cfg.MaxIdleTime

	// Create the connection pool
	pool, err := pgxpool.ConnectConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("error connecting to the database: %v", err)
	}

	db := &DB{
		pool: pool,
		cfg:  cfg,
	}

	// Start health check if configured
	if cfg.HealthCheck > 0 {
		go db.startHealthCheck()
	}

	return db, nil
}

// Health check routine
func (db *DB) startHealthCheck() {
	ticker := time.NewTicker(db.cfg.HealthCheck)
	for range ticker.C {
		err := db.Ping(context.Background())
		if err != nil {
			fmt.Printf("Database health check failed: %v\n", err)
		}
	}
}

// Ping verifies a connection to the database is still alive
func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

// Close closes the database connection pool
func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

// GetPool returns the underlying connection pool
func (db *DB) GetPool() *pgxpool.Pool {
	return db.pool
}

// Transaction represents a database transaction
type Transaction struct {
	tx pgx.Tx
}

// BeginTx starts a new transaction
func (db *DB) BeginTx(ctx context.Context) (*Transaction, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("error beginning transaction: %v", err)
	}
	return &Transaction{tx: tx}, nil
}

// Commit commits the transaction
func (t *Transaction) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

// ExecContext executes a query without returning any rows
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	return db.pool.Exec(ctx, query, args...)
}

// QueryContext executes a query that returns rows
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	return db.pool.Query(ctx, query, args...)
}

// QueryRowContext executes a query that returns a single row
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) pgx.Row {
	return db.pool.QueryRow(ctx, query, args...)
}

// Example usage of transactions
func (db *DB) ExampleTransaction(ctx context.Context) error {
	tx, err := db.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // Rollback if not committed

	// Execute queries within transaction
	_, err = tx.tx.Exec(ctx, "INSERT INTO users (name) VALUES ($1)", "John")
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit(ctx)
}
