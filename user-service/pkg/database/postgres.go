package database

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
)

func NewPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return db, nil
}

func RunMigrations(db *sql.DB) error {
	// Retrieve the DSN from the underlying driver connection.
	// We pass the DSN directly via the migrate URL instead.
	return nil // called separately with DSN below
}

func RunMigrationsWithDSN(dsn string) error {
	m, err := migrate.New("file://migrations", dsn)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate.Up: %w", err)
	}
	log.Println("migrations applied successfully")
	return nil
}
