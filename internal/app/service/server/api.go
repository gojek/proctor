package server

import (
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/app/service/server/middleware"
	"time"

	"github.com/tylerb/graceful"
	"github.com/urfave/negroni"
)

func Start() error {
	err := middleware.InitNewRelic()
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
