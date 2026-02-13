package db

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/linskybing/platform-go/internal/config"
)

func RunMigrations() error {
	migrationDSN, err := buildMigrationDSN()
	if err != nil {
		return err
	}

	m, err := migrate.New(config.DbMigrationsPath, migrationDSN)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			slog.Error("failed to close migration source", "error", sourceErr)
		}
		if dbErr != nil {
			slog.Error("failed to close migration db", "error", dbErr)
		}
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("apply migrations: %w", err)
	}

	return nil
}

func buildMigrationDSN() (string, error) {
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(config.DbUser, config.DbPassword),
		Host:   fmt.Sprintf("%s:%s", config.DbHost, config.DbPort),
		Path:   config.DbName,
	}

	query := u.Query()
	query.Set("sslmode", config.DbSSLMode)
	u.RawQuery = query.Encode()

	return u.String(), nil
}
