package plugin

import (
	"fmt"
	"plugin"
	"proctor/internal/app/service/infra/logger"
)

type GoPlugin interface {
	Load(pluginBinary string, exportedName string) (plugin.Symbol, error)
}

type goPlugin struct{}

func (g *goPlugin) Load(pluginBinary string, exportedName string) (plugin.Symbol, error) {
	binary, err := plugin.Open(pluginBinary)
	logger.LogErrors(err, "load auth plugin binary from location: ", pluginBinary)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin binary from location: %s", pluginBinary)
	}

	raw, err := binary.Lookup(exportedName)
	logger.LogErrors(err, "Lookup ", pluginBinary, " for ", exportedName)
	if err != nil {
		return nil, fmt.Errorf("failed to Lookup plugin binary from location: %s with Exported Name: %s", pluginBinary, exportedName)
	}
	return raw, nil
}

func NewGoPlugin() GoPlugin {
	return &goPlugin{}
}
