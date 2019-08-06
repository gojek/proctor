package server

import (
	"context"
	"github.com/urfave/negroni"
	"net/http"
	"os"
	"os/signal"
	"proctor/internal/app/service/infra/config"
	"proctor/internal/app/service/infra/logger"
	"proctor/internal/app/service/server/middleware"
	"syscall"
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
	httpServer := &http.Server{
		Addr:    appPort,
		Handler: server,
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)

		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint

		if shutdownErr := httpServer.Shutdown(context.Background()); shutdownErr != nil {
			logger.Error("Received an Interrupt Signal", shutdownErr)
		}
	}()

	if err = httpServer.ListenAndServe(); err != nil {
		logger.Error("HTTP Server Failed ", err)
		close(idleConnsClosed)
	}

	<-idleConnsClosed

	_ = postgresClient.Close()
	logger.Info("Stopped server gracefully")
	return nil
}
