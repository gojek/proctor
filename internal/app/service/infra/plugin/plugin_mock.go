package plugin

import (
	"github.com/stretchr/testify/mock"
	"plugin"
)

type GoPluginMock struct {
	mock.Mock
}

func (g *GoPluginMock) Load(pluginBinary string, exportedName string) (plugin.Symbol, error) {
	args := g.Called(pluginBinary, exportedName)
	return args.Get(0).(plugin.Symbol), args.Error(1)
}
