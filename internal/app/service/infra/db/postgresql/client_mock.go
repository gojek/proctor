package postgresql

import (
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/mock"
)

type ClientMock struct {
	mock.Mock
}

func (m ClientMock) NamedExec(query string, data interface{}) (int64, error) {
	args := m.Called(query, data)
	return args.Get(0).(int64), args.Error(1)
}

func (m ClientMock) Select(destination interface{}, query string, arguments ...interface{}) error {
	jobName := arguments[0]
	args := m.Called(destination, query, jobName)
	return args.Error(0)
}

func (m ClientMock) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m ClientMock) GetDB() *sqlx.DB {
	args := m.Called()
	return args.Get(0).(*sqlx.DB)
}
