package migrations

import (
	"database/sql"
	"fmt"
	"log"
)

// Migrate performs database migrations
func Migrate(db *sql.DB) error {
	log.Println("Running database migrations...")

	// Create migrations table if not exists
	if err := createMigrationsTable(db); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Run migrations
	migrations := []struct {
		name string
		up   string
	}{
		{
			name: "01_create_users_table",
			up:   createUsersTable,
		},
		{
			name: "02_create_products_table",
			up:   createProductsTable,
		},
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migrations
	for _, migration := range migrations {
		// Check if migration has already been applied
		var count int
		err := tx.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = $1", migration.name).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %w", err)
		}

		if count > 0 {
			log.Printf("Migration %s already applied", migration.name)
			continue
		}

		// Apply migration
		log.Printf("Applying migration: %s", migration.name)
		if _, err := tx.Exec(migration.up); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.name, err)
		}

		// Record migration
		if _, err := tx.Exec("INSERT INTO migrations (name) VALUES ($1)", migration.name); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migration.name, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// createMigrationsTable creates the migrations table if it doesn't exist
func createMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

// Migration SQL statements
const (
	createUsersTable = `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			email VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`

	createProductsTable = `
		CREATE TABLE IF NOT EXISTS products (
			id UUID PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			price DECIMAL(10, 2) NOT NULL,
			duration_months INTEGER NOT NULL,
			tax_rate DECIMAL(5, 2) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL
		)
	`
)
