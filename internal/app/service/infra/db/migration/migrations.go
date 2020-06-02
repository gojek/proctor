package migration

import (
	"fmt"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"

	"github.com/mattes/migrate"
	//postgres driver
	_ "github.com/mattes/migrate/database/postgres"
	//driver for reading migrations from file
	_ "github.com/mattes/migrate/source/file"
)

var migrationsPath, postgresConnectionURL string

func init() {
	postgresConnectionURL = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", config.Config().PostgresUser, config.Config().PostgresPassword, config.Config().PostgresHost, config.Config().PostgresPort, config.Config().PostgresDatabase)
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
