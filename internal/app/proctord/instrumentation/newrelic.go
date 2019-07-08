package instrumentation

import (
	newrelic "github.com/newrelic/go-agent"
	"net/http"
	"proctor/internal/app/service/infra/config"
)

var NewRelicApp newrelic.Application

func InitNewRelic() error {
	appName := config.NewRelicAppName()
	licenceKey := config.NewRelicLicenceKey()
	newRelicConfig := newrelic.NewConfig(appName, licenceKey)
	newRelicConfig.Enabled = true
	app, err := newrelic.NewApplication(newRelicConfig)
	if err != nil {
		return err
	}
	NewRelicApp = app
	return nil
}

func Wrap(pattern string, handlerFunc http.HandlerFunc) (string, func(http.ResponseWriter, *http.Request)) {
	return newrelic.WrapHandleFunc(NewRelicApp, pattern, handlerFunc)
}
