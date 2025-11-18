package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func runGooseMigrations() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/pxbox?sslmode=disable"
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Set goose dialect
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	// Determine migrations directory
	migrationsDir := os.Getenv("GOOSE_MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	// Check if directory exists
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory not found: %s", migrationsDir)
	}

	// Run migrations
	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

