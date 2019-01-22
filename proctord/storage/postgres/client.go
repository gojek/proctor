package postgres

import (
	"fmt"
	"time"

	"github.com/gojektech/proctor/proctord/config"
	"github.com/gojektech/proctor/proctord/logger"
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
	dataSourceName := fmt.Sprintf("dbname=%s user=%s password=%s host=%s sslmode=disable", config.PostgresDatabase(), config.PostgresUser(), config.PostgresPassword(), config.PostgresHost())

	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		panic(err.Error())
	}

	db.SetMaxIdleConns(config.PostgresMaxConnections())
	db.SetMaxOpenConns(config.PostgresMaxConnections())
	db.SetConnMaxLifetime(time.Duration(config.PostgresConnectionMaxLifetime()) * time.Minute)

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
