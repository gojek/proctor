package redis

import "github.com/stretchr/testify/mock"

type MockClient struct {
	mock.Mock
}

func (m *MockClient) GET(key string) ([]byte, error) {
	args := m.Called(key)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockClient) SET(key string, value []byte) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockClient) KEYS(regex string) ([]string, error) {
	args := m.Called(regex)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockClient) MGET(keys ...interface{}) ([][]byte, error) {
	args := m.Called(keys...)
	return args.Get(0).([][]byte), args.Error(1)
}
