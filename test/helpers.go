package test

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// SetupTestDB sets up a test database connection
func SetupTestDB() (*sql.DB, error) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5433/pxbox_test?sslmode=disable"
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping test database: %w", err)
	}

	return db, nil
}

// RunMigrations runs migrations on test database
func RunMigrations(db *sql.DB) error {
	migrationsDir := "../migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = "./migrations"
	}

	// Simple migration runner for tests
	migrationSQL, err := os.ReadFile(migrationsDir + "/0001_init.sql")
	if err != nil {
		return fmt.Errorf("failed to read migration: %w", err)
	}

	if _, err := db.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("failed to run migration: %w", err)
	}

	return nil
}

// CleanupTestDB cleans up test database
func CleanupTestDB(db *sql.DB) error {
	tables := []string{"reminders", "responses", "requests", "flows", "entities", "schema_migrations"}
	for _, table := range tables {
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			// Ignore errors if table doesn't exist
		}
	}
	return nil
}

// StartTestServices starts docker-compose services for testing
func StartTestServices() error {
	cmd := exec.Command("docker-compose", "-f", "test/docker-compose.test.yaml", "up", "-d")
	cmd.Dir = ".."
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start test services: %w", err)
	}

	// Wait for services to be ready
	time.Sleep(5 * time.Second)
	return nil
}

// StopTestServices stops docker-compose services
func StopTestServices() error {
	cmd := exec.Command("docker-compose", "-f", "test/docker-compose.test.yaml", "down")
	cmd.Dir = ".."
	return cmd.Run()
}

