package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func runMigrations() error {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/pxbox?sslmode=disable"
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer db.Close()

	// Create migrations table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get applied migrations
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return fmt.Errorf("failed to query migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int]bool)
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return err
		}
		applied[version] = true
	}

	// Read migration files
	migrationsDir := "migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		// Try relative path
		migrationsDir = "./migrations"
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.sql"))
	if err != nil {
		return fmt.Errorf("failed to read migrations: %w", err)
	}

	// Sort files by version number
	sort.Slice(files, func(i, j int) bool {
		vi := extractVersion(files[i])
		vj := extractVersion(files[j])
		return vi < vj
	})

	// Apply migrations
	for _, file := range files {
		version := extractVersion(file)
		if applied[version] {
			fmt.Printf("Migration %d already applied, skipping\n", version)
			continue
		}

		fmt.Printf("Applying migration %d: %s\n", version, filepath.Base(file))

		sql, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file: %w", err)
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(string(sql)); err != nil {
			// Check if error is due to relation already existing (idempotent migrations)
			if strings.Contains(err.Error(), "already exists") {
				fmt.Printf("Migration %d: relations already exist, marking as applied\n", version)
				tx.Rollback()
				// Mark as applied even though we rolled back
				if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT DO NOTHING", version); err != nil {
					return fmt.Errorf("failed to record migration %d: %w", version, err)
				}
				continue
			}
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %d: %w", version, err)
		}

		if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", version, err)
		}

		fmt.Printf("Migration %d applied successfully\n", version)
	}

	return nil
}

func extractVersion(filename string) int {
	base := filepath.Base(filename)
	parts := strings.Split(base, "_")
	if len(parts) > 0 {
		var version int
		fmt.Sscanf(parts[0], "%d", &version)
		return version
	}
	return 0
}

