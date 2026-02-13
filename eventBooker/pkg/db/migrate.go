package postgres

import (
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func MigrateUp(connString string, sourcePath string) error {
	m, err := migrate.New(sourcePath, connString)
	if err != nil {
		return fmt.Errorf("unable to create migrations table: %v", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("unable to apply migrations: %v", err)
	}
	return nil
}

func MigrateUpClickHouse(chDSN, sourcePath string) error {
	m, err := migrate.New(sourcePath, chDSN)
	if err != nil {
		return fmt.Errorf("create migrate: %w", err)
	}
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
