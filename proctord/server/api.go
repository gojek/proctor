package server

import (
	"proctor/proctord/config"
	"proctor/proctord/instrumentation"
	"proctor/proctord/logger"
	"proctor/proctord/redis"
	"proctor/proctord/storage/postgres"
	"time"

	"github.com/tylerb/graceful"
	"github.com/urfave/negroni"
)

func Start() error {
	redisClient := redis.NewClient()
	postgresClient := postgres.NewClient()

	err := instrumentation.InitNewRelic()
	if err != nil {
		logger.Fatal(err)
	}
	appPort := ":" + config.AppPort()

	server := negroni.New(negroni.NewRecovery())
	router, err := NewRouter(postgresClient, redisClient)
	if err != nil {
		return err
	}
	server.UseHandler(router)

	logger.Info("Starting server on port", appPort)

	graceful.Run(appPort, 2*time.Second, server)

	postgresClient.Close()
	logger.Info("Stopped server gracefully")
	return nil
}
