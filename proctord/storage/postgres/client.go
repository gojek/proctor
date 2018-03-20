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
	NamedExec(string, interface{}) error
	Close() error
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

func (client *client) NamedExec(query string, data interface{}) error {
	_, err := client.db.NamedExec(query, data)
	return err
}

func (client *client) Close() error {
	logger.Info("Closing connections to db")
	return client.db.Close()
}
