package postgres

import (
	"fmt"

	"github.com/gojektech/proctor/proctord/config"
	"github.com/gojektech/proctor/proctord/logger"
	"github.com/mattes/migrate"
	//postgres driver
	_ "github.com/mattes/migrate/database/postgres"
	//driver for reading migrations from file
	_ "github.com/mattes/migrate/source/file"
)

var migrationsPath, postgresConnectionURL string

func init() {
	postgresConnectionURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", config.PostgresUser(), config.PostgresPassword(), config.PostgresHost(), config.PostgresPort(), config.PostgresDatabase())
	migrationsPath = "file://./migrations"
}

func Up() error {
	m, err := migrate.New(migrationsPath, postgresConnectionURL)
	if err != nil {
		return err
	}

	err = m.Up()
	if err == migrate.ErrNoChange {
		logger.Info("No migrations run")
		return nil
	}

	return err
}

func DownOneStep() error {
	m, err := migrate.New(migrationsPath, postgresConnectionURL)
	if err != nil {
		return err
	}

	return m.Steps(-1)
}
