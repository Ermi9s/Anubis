package config

import (
	"embed"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed pg_db/migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(dbURL string) error {
	d, err := iofs.New(migrationsFS, "pg_db/migrations")
	if err != nil {
		return err
	}

	// Create migrate instance from iofs source
	log.Printf("driver created: %s", d)


	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		return err
	}

	log.Println("migrate instance")

	// Run "up" migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}