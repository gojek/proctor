package server

import (
	"proctor/internal/app/proctord/instrumentation"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"
	"time"

	"github.com/tylerb/graceful"
	"github.com/urfave/negroni"
)

func Start() error {
	err := instrumentation.InitNewRelic()
	if err != nil {
		logger.Fatal(err)
	}
	appPort := ":" + config.AppPort()

	server := negroni.New(negroni.NewRecovery())
	router, err := NewRouter()
	if err != nil {
		return err
	}
	server.UseHandler(router)

	logger.Info("Starting server on port", appPort)

	graceful.Run(appPort, 2*time.Second, server)

	_ = postgresClient.Close()
	logger.Info("Stopped server gracefully")
	return nil
}
