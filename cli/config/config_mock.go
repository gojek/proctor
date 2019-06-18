package config

import "github.com/stretchr/testify/mock"

type MockLoader struct {
	mock.Mock
}

func (m *MockLoader) Load() (ProctorConfig, ConfigError) {
	args := m.Called()
	return args.Get(0).(ProctorConfig), args.Get(1).(ConfigError)
}
