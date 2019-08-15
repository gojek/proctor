package postgresql

import (
	"fmt"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"
	"time"

	"github.com/jmoiron/sqlx"
	//postgres driver
	_ "github.com/lib/pq"
)

type Client interface {
	NamedExec(string, interface{}) (int64, error)
	Select(interface{}, string, ...interface{}) error
	Close() error
	GetDB() *sqlx.DB
}

type client struct {
	db *sqlx.DB
}

func NewClient() Client {
	dataSourceName := fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=disable", config.Config().PostgresDatabase, config.Config().PostgresUser, config.Config().PostgresPassword, config.Config().PostgresHost)

	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		panic(err.Error())
	}

	db.SetMaxIdleConns(config.Config().PostgresMaxConnections)
	db.SetMaxOpenConns(config.Config().PostgresMaxConnections)
	db.SetConnMaxLifetime(time.Duration(config.Config().PostgresConnectionMaxLifetime) * time.Minute)

	return &client{
		db: db,
	}
}

func (client *client) NamedExec(query string, data interface{}) (int64, error) {
	result, err := client.db.NamedExec(query, data)
	if result == nil {
		return int64(0), err
	}

	rowsAffected, err := result.RowsAffected()
	return rowsAffected, err
}

func (client *client) Select(destination interface{}, query string, args ...interface{}) error {
	return client.db.Select(destination, query, args...)
}

func (client *client) Close() error {
	logger.Info("Closing connections to db")
	return client.db.Close()
}

func (client *client) GetDB() *sqlx.DB {
	return client.db
}
