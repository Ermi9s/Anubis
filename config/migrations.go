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

	m, err := migrate.NewWithSourceInstance("iofs", d, dbURL)
	if err != nil {
		return err
	}

	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Printf("[Anubis] migration source close error: %v", srcErr)
		}
		if dbErr != nil {
			log.Printf("[Anubis] migration db close error: %v", dbErr)
		}
	}()

	log.Printf("[Anubis] running migrations on %s", dbURL)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("[Anubis] all migrations are up!!!")
	return nil
}

