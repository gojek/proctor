package middleware

import (
	"github.com/gorilla/mux"
	newrelic "github.com/newrelic/go-agent"
	"github.com/newrelic/go-agent/_integrations/nrgorilla/v1"
	"proctor/internal/app/service/infra/config"
)

var newRelicApp newrelic.Application

func InitNewRelic() error {
	appName := config.Config().NewRelicAppName
	licenceKey := config.Config().NewRelicLicenseKey
	newRelicConfig := newrelic.NewConfig(appName, licenceKey)
	newRelicConfig.Enabled = true
	app, err := newrelic.NewApplication(newRelicConfig)
	if err != nil {
		return err
	}
	newRelicApp = app
	return nil
}

func InstrumentNewRelic(r *mux.Router) *mux.Router {
	return nrgorilla.InstrumentRoutes(r, newRelicApp)
}
