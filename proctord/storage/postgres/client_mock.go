package postgres

import "github.com/stretchr/testify/mock"

type ClientMock struct {
	mock.Mock
}

func (m ClientMock) NamedExec(query string, data interface{}) error {
	args := m.Called(query, data)
	return args.Error(0)
}

func (m ClientMock) Close() error {
	args := m.Called()
	return args.Error(0)
}
